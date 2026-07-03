package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/env"
)

const (
	DefaultMappingsDir      = "data/artifacts/mappings"
	DefaultRedisTTL         = 24 * time.Hour
	DefaultJWTExpiration    = 16 * time.Hour
	DefaultInferenceTimeout = 5 * time.Second

	tcpPort     = "9001"
	metricsPort = "9002"
)

type APIConfig struct {
	InferenceHosts   []string
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

	hosts := splitAndTrim(inferenceAddrsStr, ",")
	if len(hosts) == 0 {
		return nil, fmt.Errorf("INFERENCE_ADDRS must contain at least one address")
	}

	return &APIConfig{
		InferenceHosts:   hosts,
		MongoURI:         mongoURI,
		RedisAddr:        redisAddr,
		JWTSecret:        jwtSecret,
		MappingsDir:      env.GetEnv("MAPPINGS_DIR", DefaultMappingsDir),
		RedisTTL:         DefaultRedisTTL,
		JWTExpiration:    DefaultJWTExpiration,
		InferenceTimeout: DefaultInferenceTimeout,
	}, nil
}

func (c *APIConfig) InferenceTCPAddrs() []string {
	addrs := make([]string, len(c.InferenceHosts))
	for i, host := range c.InferenceHosts {
		addrs[i] = host + ":" + tcpPort
	}
	return addrs
}

func (c *APIConfig) InferenceMetricsAddrs() []string {
	addrs := make([]string, len(c.InferenceHosts))
	for i, host := range c.InferenceHosts {
		addrs[i] = host + ":" + metricsPort
	}
	return addrs
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
