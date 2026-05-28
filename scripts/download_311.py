import csv
import io
import os
import sys
import time
from typing import Dict, List, Sequence

import requests

BASE_URL = "https://data.cityofnewyork.us/resource/erm2-nwe9.csv"
OUTPUT_PATH = "data/raw/nyc_311_4M.csv"
TARGET_ROWS = 4_000_000
PAGE_SIZE = 5_000
ORDER_BY = "created_date DESC"
MAX_RETRIES = 5
SLEEP_SECONDS = 2.0

def ensure_parent_dir(path: str) -> None:
    parent = os.path.dirname(path)
    if parent:
        os.makedirs(parent, exist_ok=True)

def fetch_page(session: requests.Session, offset: int, limit: int) -> List[Dict[str, str]]:
    params = {
        "$limit": str(limit),
        "$offset": str(offset),
        "$order": ORDER_BY,
    }

    last_error: Exception | None = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            response = session.get(BASE_URL, params=params, timeout=120)
            response.raise_for_status()

            return list(csv.DictReader(io.StringIO(response.text)))

        except requests.RequestException as exc:
            last_error = exc
            wait = SLEEP_SECONDS * attempt
            print(
                f"[warn] offset={offset} attempt={attempt}/{MAX_RETRIES} failed: {exc}. "
                f"Retrying in {wait:.1f}s...",
                file=sys.stderr,
            )
            time.sleep(wait)

    raise RuntimeError(f"Failed to fetch offset={offset} after {MAX_RETRIES} retries: {last_error}")

def write_rows(path: str, fieldnames: Sequence[str], rows: List[Dict[str, str]], append: bool) -> None:
    mode = "a" if append else "w"
    with open(path, mode, newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        if not append:
            writer.writeheader()
        writer.writerows(rows)

def main() -> int:
    ensure_parent_dir(OUTPUT_PATH)

    session = requests.Session()
    session.headers.update({"Accept": "text/csv", "User-Agent": "tb2-nyc311-downloader/1.0"})

    total_written = 0
    offset = 0
    append = False
    fieldnames: List[str] | None = None

    while total_written < TARGET_ROWS:
        remaining = TARGET_ROWS - total_written
        limit = min(PAGE_SIZE, remaining)

        rows = fetch_page(session, offset=offset, limit=limit)
        if not rows:
            print("[info] API returned no more rows. Stopping early.")
            break

        if fieldnames is None:
            fieldnames = list(rows[0].keys())

        write_rows(OUTPUT_PATH, fieldnames, rows, append=append)
        append = True

        downloaded = len(rows)
        total_written += downloaded
        offset += downloaded

        print(f"[info] downloaded {downloaded} rows | total={total_written}/{TARGET_ROWS} | offset={offset}")

        if downloaded < limit:
            break

    print(f"[done] wrote {total_written} rows to {OUTPUT_PATH}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
