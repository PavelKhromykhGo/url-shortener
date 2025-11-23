package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Env              string
	HTTPAddr         string
	PostgresDSN      string
	RedisAddr        string
	RedisDB          int
	RedisPassword    string
	KafkaBrokers     []string
	KafkaClicksTopic string
	BaseURL          string
}

func Load() (*Config, error) {
	cfg := &Config{
		Env:              getEnv("APP_ENV", "dev"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		PostgresDSN:      getEnv("POSTGRES_DSN", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisDB:          getEnvInt("REDIS_DB", 0),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:     splitComma(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaClicksTopic: getEnv("KAFKA_CLICKS_TOPIC", "clicks"),
		BaseURL:          getEnv("BASE_URL", "http://localhost:8080"),
	}

	if cfg.PostgresDSN == "" {
		return nil, fmt.Errorf("POSTGRES_DSN is required")
	}
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func getEnvInt(key string, def int) int {
	valStr, ok := os.LookupEnv(key)
	if !ok || valStr == "" {
		return def
	}

	v, err := strconv.Atoi(valStr)
	if err != nil {
		return def
	}
	return v
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))

	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
