import json
from pathlib import Path

MAPPINGS_DIR = Path("data/artifacts/mappings")

def load_json(name: str) -> dict:
    path = MAPPINGS_DIR / name
    with path.open() as f:
        return json.load(f)

def sorted_values(data: dict) -> list[str]:
    return [data[str(i)] for i in range(len(data))]

def print_const_array(name: str, values: list[str]) -> None:
    print(f"export const {name} = [")
    for v in values:
        escaped = v.replace("'", "\\'")
        print(f"\t'{escaped}',")
    print("] as const;\n")

def print_const_record(name: str, keys: list[str], values: list[str]) -> None:
    print(f"export const {name}: Record<string, string> = {{")
    for k, v in zip(keys, values):
        escaped_k = k.replace("'", "\\'")
        escaped_v = v.replace("'", "\\'")
        print(f"\t'{escaped_k}': '{escaped_v}',")
    print("};\n")

def main() -> None:
    agency_map = load_json("agency_map.json")
    agency_name_map = load_json("agency_name_map.json")
    borough_map = load_json("borough_map.json")
    location_map = load_json("location_map.json")
    complaint_map = load_json("complaint_map.json")
    descriptor_map = load_json("descriptor_map.json")

    agencies = sorted_values(agency_map)
    agency_names = sorted_values(agency_name_map)
    boroughs = sorted_values(borough_map)
    locations = sorted_values(location_map)
    complaints = sorted_values(complaint_map)
    descriptors = sorted_values(descriptor_map)

    print_const_array("AGENCIES", agencies)
    print_const_record("AGENCY_NAMES", agencies, agency_names)
    print_const_array("BOROUGHS", boroughs)
    print_const_array("LOCATION_TYPES", locations)
    print_const_array("COMPLAINT_TYPES", complaints)
    print_const_array("DESCRIPTORS", descriptors)

if __name__ == "__main__":
    main()
