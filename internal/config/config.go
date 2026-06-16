package config

import (
	"fmt"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type Config struct {
	DataPath string
	MongoURI string
}

func Load() (*Config, error) {
	mongoURI := env.GetEnv("MONGO_URI", "")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	return &Config{
		DataPath: env.GetEnv("DATA_PATH", "data/training/nyc_311_features.csv"),
		MongoURI: mongoURI,
	}, nil
}
