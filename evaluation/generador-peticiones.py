import csv
import json
import random
import time
from pathlib import Path

# Carpeta donde están los mappings
MAPPINGS_DIR = "mappings"

# Cantidad de peticiones a generar
N = 10
SEED = 42
random.seed(SEED)

def load_mapping(filename):
    with open(Path(MAPPINGS_DIR) / filename, "r", encoding="utf-8") as f:
        data = json.load(f)

    if isinstance(data, dict):
        return list(data.values())

    return data


agency = load_mapping("agency_map.json")
complaint = load_mapping("complaint_map.json")
descriptor = load_mapping("descriptor_map.json")
location = load_mapping("location_map.json")
borough = load_mapping("borough_map.json")

start_ts = int(time.mktime(time.strptime("2023-01-01", "%Y-%m-%d")))
end_ts = int(time.mktime(time.strptime("2024-12-31", "%Y-%m-%d")))
output = f"predict_requests_{N}.csv"

with open(output, "w", newline="", encoding="utf-8") as csvfile:

    writer = csv.writer(csvfile)

    # Encabezados (deben coincidir con las variables de Postman)
    writer.writerow([
        "timestamp",
        "latitude",
        "longitude",
        "agency",
        "complaint_type",
        "descriptor",
        "location_type",
        "borough"
    ])

    for _ in range(N):

        writer.writerow([
            random.randint(start_ts, end_ts),
            round(random.uniform(40.50, 40.92), 6),
            round(random.uniform(-74.25, -73.70), 6),
            random.choice(agency),
            random.choice(complaint),
            random.choice(descriptor),
            random.choice(location),
            random.choice(borough)
        ])

print(f"Generadas {N} peticiones.")
print(f"Archivo creado: {output}")