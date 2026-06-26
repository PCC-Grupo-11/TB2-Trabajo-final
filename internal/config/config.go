package config

import (
	"runtime"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type Config struct {
	DataPath   string
	MongoURI   string
	NumWorkers int
	Sequential bool
}

func Load() (*Config, error) {
	/*
		mongoURI, err := env.RequiredEnv("MONGO_URI")
		if err != nil {
			return nil, err
		}
	*/

	mongoURI := env.GetEnv("MONGO_URI", "")

	workers := runtime.NumCPU()
	if env.GetEnv("SEQUENTIAL", "false") == "true" {
		workers = 1
	}

	return &Config{
		DataPath:   env.GetEnv("DATA_PATH", "data/training/nyc_311_features.csv"),
		MongoURI:   mongoURI,
		NumWorkers: workers,
		Sequential: workers == 1,
	}, nil
}
