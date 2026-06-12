import csv
import os
import sys
import time
from pathlib import Path
from typing import Optional

import requests

BASE_URL = "https://data.cityofnewyork.us/resource/erm2-nwe9.csv"

OUTPUT_PATH = Path("data/raw/nyc_311_3M.csv")
CHECKPOINT_PATH = Path("data/raw/nyc_311_checkpoint.txt")

APP_TOKEN = os.getenv("SOCRATA_APP_TOKEN")

TARGET_ROWS = 3_000_000
PAGE_SIZE = 100_000

MAX_RETRIES = 5
BASE_SLEEP_SECONDS = 2.0

SELECT_COLUMNS = [
    "unique_key",
    "created_date",
    "closed_date",
    "agency",
    "agency_name",
    "complaint_type",
    "descriptor",
    "location_type",
    "city",
    "borough",
    "latitude",
    "longitude",
]

def ensure_parent_dir(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)

def load_checkpoint() -> tuple[str, str]:
    """
    Returns:
        (last_created_date, last_unique_key)
    """

    default = (
        "9999-12-31T23:59:59.999",
        "999999999999",
    )

    if not CHECKPOINT_PATH.exists():
        return default

    content = CHECKPOINT_PATH.read_text().strip()

    if not content:
        return default

    if "|" not in content:
        return default

    created_date, unique_key = content.split("|", maxsplit=1)

    return created_date, unique_key

def save_checkpoint(
    last_created_date: str,
    last_unique_key: str,
) -> None:
    CHECKPOINT_PATH.write_text(
        f"{last_created_date}|{last_unique_key}"
    )

def build_params(
    last_created_date: str,
    last_unique_key: str,
) -> dict[str, str]:

    where_clause = (
        f"closed_date IS NOT NULL "
        f"AND latitude IS NOT NULL "
        f"AND longitude IS NOT NULL "
        f"AND complaint_type IS NOT NULL "
        f"AND descriptor IS NOT NULL "
        f"AND closed_date >= created_date "
        f"AND created_date >= '2020-01-01T00:00:00.000' "
        f"AND ("
        f"created_date < '{last_created_date}' "
        f"OR ("
        f"created_date = '{last_created_date}' "
        f"AND unique_key < '{last_unique_key}'"
        f")"
        f")"
    )

    return {
        "$select": ",".join(SELECT_COLUMNS),
        "$where": where_clause,
        "$order": "created_date DESC, unique_key DESC",
        "$limit": str(PAGE_SIZE),
    }

def fetch_page(
    session: requests.Session,
    last_created_date: str,
    last_unique_key: str,
) -> requests.Response:

    params = build_params(
        last_created_date=last_created_date,
        last_unique_key=last_unique_key,
    )

    last_error: Optional[Exception] = None

    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = session.get(
                BASE_URL,
                params=params,
                stream=True,
                timeout=(30, 300),
            )

            response.raise_for_status()

            return response

        except requests.RequestException as exc:
            last_error = exc

            wait = BASE_SLEEP_SECONDS * attempt

            print(
                f"[warn] "
                f"created_date={last_created_date} "
                f"unique_key={last_unique_key} "
                f"attempt={attempt}/{MAX_RETRIES} failed: {exc} "
                f"retrying in {wait:.1f}s...",
                file=sys.stderr,
            )

            time.sleep(wait)

    raise RuntimeError(
        f"Failed after {MAX_RETRIES} retries: {last_error}"
    )

def append_csv_from_response(
    response: requests.Response,
    output_path: Path,
    write_header: bool,
) -> tuple[int, Optional[str], Optional[str]]:

    rows_written = 0

    last_created_date = None
    last_unique_key = None

    response.encoding = "utf-8"

    lines = response.iter_lines(decode_unicode=True)

    reader = csv.DictReader(lines) # type: ignore

    if reader.fieldnames is None:
        return 0, None, None

    with output_path.open(
        "a",
        newline="",
        encoding="utf-8",
    ) as f:

        writer = csv.DictWriter(
            f,
            fieldnames=reader.fieldnames,
        )

        if write_header:
            writer.writeheader()

        for row in reader:
            writer.writerow(row)

            rows_written += 1

            last_created_date = row["created_date"]
            last_unique_key = row["unique_key"]

    return (
        rows_written,
        last_created_date,
        last_unique_key,
    )

def count_existing_rows(path: Path) -> int:
    if not path.exists():
        return 0

    with path.open("r", encoding="utf-8") as f:
        return max(sum(1 for _ in f) - 1, 0)

def main() -> int:
    ensure_parent_dir(OUTPUT_PATH)

    session = requests.Session()

    session.headers.update({
        "Accept": "text/csv",
        "Accept-Encoding": "gzip",
        "User-Agent": "nyc311-downloader/3.0",
    })

    if APP_TOKEN:
        session.headers["X-App-Token"] = APP_TOKEN

    total_written = count_existing_rows(OUTPUT_PATH)

    last_created_date, last_unique_key = load_checkpoint()

    print(f"[info] resuming from created_date={last_created_date} unique_key={last_unique_key}")

    print(f"[info] existing rows={total_written:,}")

    write_header = not OUTPUT_PATH.exists()

    start_time = time.time()

    while total_written < TARGET_ROWS:
        response = fetch_page(
            session=session,
            last_created_date=last_created_date,
            last_unique_key=last_unique_key,
        )

        (
            rows_written,
            newest_created_date,
            newest_unique_key,
        ) = append_csv_from_response(
            response=response,
            output_path=OUTPUT_PATH,
            write_header=write_header,
        )

        response.close()

        write_header = False

        if rows_written == 0:
            print("[info] no more rows returned")
            break

        total_written += rows_written

        if (
            newest_created_date is not None
            and newest_unique_key is not None
        ):
            last_created_date = newest_created_date
            last_unique_key = newest_unique_key

            save_checkpoint(
                last_created_date,
                last_unique_key,
            )

        elapsed = time.time() - start_time

        rate = total_written / elapsed if elapsed > 0 else 0

        print(
            f"[info] downloaded={rows_written:,} "
            f"total={total_written:,}/{TARGET_ROWS:,} "
            f"rate={rate:,.0f} rows/sec "
            f"created_date={last_created_date} "
            f"unique_key={last_unique_key}"
        )

        if rows_written < PAGE_SIZE:
            print("[info] reached dataset end")
            break

    print(f"[done] wrote {total_written:,} rows")
    print(f"[done] output={OUTPUT_PATH}")

    return 0

if __name__ == "__main__":
    raise SystemExit(main())
