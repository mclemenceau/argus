package config

import (
	"fmt"
	"os"
)

type Config struct {
	OpenRouterAPIKey string
	LLMModel         string
	DefaultRelease   string
	TestObserverURL  string
	TemporalHost     string
	Port             string
	ServerURL        string // base URL the worker uses to push to the HTTP server
}

func Load() (*Config, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is required")
	}

	return &Config{
		OpenRouterAPIKey: apiKey,
		LLMModel:         envOrDefault("LLM_MODEL", "anthropic/claude-sonnet-4-5"),
		DefaultRelease:   os.Getenv("DEFAULT_RELEASE"), // empty = auto-detect from data
		TestObserverURL:  envOrDefault("TEST_OBSERVER_URL", "https://tests-api.ubuntu.com"),
		TemporalHost:     envOrDefault("TEMPORAL_HOST", "localhost:7233"),
		Port:             envOrDefault("PORT", "8080"),
		ServerURL:        envOrDefault("SERVER_URL", "http://localhost:8080"),
	}, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
