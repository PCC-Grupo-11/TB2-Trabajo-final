package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

const (
	DefaultAPIPort          = "8080"
	DefaultMappingsDir      = "data/artifacts/mappings"
	DefaultRedisTTL         = 4 * time.Hour
	DefaultJWTExpiration    = 3 * time.Hour
	DefaultInferenceTimeout = 5 * time.Second
)

type APIConfig struct {
	Port             string
	InferenceAddrs   []string
	MongoURI         string
	RedisAddr        string
	JWTSecret        string
	MappingsDir      string
	RedisTTL         time.Duration
	JWTExpiration    time.Duration
	InferenceTimeout time.Duration
}

func LoadAPIConfig() (*APIConfig, error) {
	jwtSecret := env.GetEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	inferenceAddrsStr := env.GetEnv("INFERENCE_ADDRS", "")
	if inferenceAddrsStr == "" {
		return nil, fmt.Errorf("INFERENCE_ADDRS is required")
	}

	mongoURI := env.GetEnv("MONGO_URI", "")
	if mongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required")
	}

	redisAddr := env.GetEnv("REDIS_ADDR", "")
	if redisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}

	return &APIConfig{
		Port:             env.GetEnv("API_PORT", DefaultAPIPort),
		InferenceAddrs:   splitAndTrim(inferenceAddrsStr, ","),
		MongoURI:         mongoURI,
		RedisAddr:        redisAddr,
		JWTSecret:        jwtSecret,
		MappingsDir:      env.GetEnv("MAPPINGS_DIR", DefaultMappingsDir),
		RedisTTL:         DefaultRedisTTL,
		JWTExpiration:    DefaultJWTExpiration,
		InferenceTimeout: DefaultInferenceTimeout,
	}, nil
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for p := range strings.SplitSeq(s, sep) {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
