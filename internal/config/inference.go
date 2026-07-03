package config

import (
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

type InferenceConfig struct {
	ModelPath string
}

func LoadInferenceConfig() (*InferenceConfig, error) {
	return &InferenceConfig{
		ModelPath: env.GetEnv("MODEL_PATH", "/models/model.json"),
	}, nil
}
