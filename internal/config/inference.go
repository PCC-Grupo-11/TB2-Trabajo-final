package config

import (
	"fmt"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type InferenceConfig struct {
	Port     string
	MongoURI string
}

func LoadInferenceConfig() (*InferenceConfig, error) {
	mongoURI := env.GetEnv("MONGO_URI", "")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	return &InferenceConfig{
		Port:     env.GetEnv("INFERENCE_PORT", "9001"),
		MongoURI: mongoURI,
	}, nil
}
