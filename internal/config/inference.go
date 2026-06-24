package config

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type InferenceConfig struct {
	Port      string
	ModelPath string
}

func LoadInferenceConfig() (*InferenceConfig, error) {
	return &InferenceConfig{
		Port:      env.GetEnv("INFERENCE_PORT", "9001"),
		ModelPath: env.GetEnv("MODEL_PATH", "/models/model.json"),
	}, nil
}
