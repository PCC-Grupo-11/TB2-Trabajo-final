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
	DefaultRedisTTL         = 24 * time.Hour
	DefaultJWTExpiration    = 16 * time.Hour
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
	jwtSecret, err := env.RequiredEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	inferenceAddrsStr, err := env.RequiredEnv("INFERENCE_ADDRS")
	if err != nil {
		return nil, err
	}

	mongoURI, err := env.RequiredEnv("MONGO_URI")
	if err != nil {
		return nil, err
	}

	redisAddr, err := env.RequiredEnv("REDIS_ADDR")
	if err != nil {
		return nil, err
	}

	addrs := splitAndTrim(inferenceAddrsStr, ",")
	if len(addrs) == 0 {
		return nil, fmt.Errorf("INFERENCE_ADDRS must contain at least one address")
	}

	return &APIConfig{
		Port:             env.GetEnv("API_PORT", DefaultAPIPort),
		InferenceAddrs:   addrs,
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
