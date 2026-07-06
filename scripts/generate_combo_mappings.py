import csv
from collections import defaultdict
from pathlib import Path

SCHEMA_PATH = Path("data/artifacts/complaint_schema.csv")

def main() -> None:
    complaint_descriptors: dict[str, set[str]] = defaultdict(set)
    descriptor_locations: dict[str, set[str]] = defaultdict(set)
    complaint_locations: dict[str, set[str]] = defaultdict(set)

    with SCHEMA_PATH.open() as f:
        reader = csv.DictReader(f)
        for row in reader:
            ct = row["complaint_type"]
            desc = row["descriptor"]
            loc = row["location_type"]

            complaint_descriptors[ct].add(desc)
            complaint_locations[ct].add(loc)
            descriptor_locations[f"{ct}|{desc}"].add(loc)

    print("// Auto-generated from complaint_schema.csv — do not edit manually.")
    print()

    print("export const COMPLAINT_DESCRIPTORS: Record<string, string[]> = {")
    for ct in sorted(complaint_descriptors):
        descs = sorted(complaint_descriptors[ct])
        escaped_descs = ", ".join(f"'{d}'" for d in descs)
        print(f"\t'{ct}': [{escaped_descs}],")
    print("};")
    print()

    print("export const DESCRIPTOR_LOCATIONS: Record<string, string[]> = {")
    for key in sorted(descriptor_locations):
        locs = sorted(descriptor_locations[key])
        escaped_locs = ", ".join(f"'{l}'" for l in locs)
        print(f"\t'{key}': [{escaped_locs}],")
    print("};")

    print("export const COMPLAINT_LOCATIONS: Record<string, string[]> = {")
    for ct in sorted(complaint_locations):
        locs = sorted(complaint_locations[ct])
        escaped_locs = ", ".join(f"'{l}'" for l in locs)
        print(f"\t'{ct}': [{escaped_locs}],")
    print("};")

if __name__ == "__main__":
    main()
