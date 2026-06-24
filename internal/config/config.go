package config

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type Config struct {
	DataPath string
	MongoURI string
}

func Load() (*Config, error) {
	mongoURI, err := env.RequiredEnv("MONGO_URI")
	if err != nil {
		return nil, err
	}

	return &Config{
		DataPath: env.GetEnv("DATA_PATH", "data/training/nyc_311_features.csv"),
		MongoURI: mongoURI,
	}, nil
}
