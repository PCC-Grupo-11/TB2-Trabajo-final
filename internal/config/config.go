package config

import "os"

const GlobalSeed = 42
const BatchSize = 4096

type Config struct {
	DataPath        string
	ModelOutputPath string
	MetadataPath    string
}

func Load() *Config {
	return &Config{
		DataPath:        getEnv("DATA_PATH", "data/training/nyc_311_features.csv"),
		ModelOutputPath: getEnv("MODEL_OUTPUT_PATH", "artifacts/models/latest.json"),
		MetadataPath:    getEnv("METADATA_OUTPUT_PATH", "artifacts/metadata/training_report.json"),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
