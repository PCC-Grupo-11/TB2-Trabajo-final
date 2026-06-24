package config

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type InferenceConfig struct {
	Port     string
	MongoURI string
}

func LoadInferenceConfig() (*InferenceConfig, error) {
	mongoURI, err := env.RequiredEnv("MONGO_URI")
	if err != nil {
		return nil, err
	}

	return &InferenceConfig{
		Port:     env.GetEnv("INFERENCE_PORT", "9001"),
		MongoURI: mongoURI,
	}, nil
}
