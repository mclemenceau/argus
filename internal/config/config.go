package config

import (
	"fmt"
	"os"
)

type Config struct {
	AnthropicAPIKey  string
	OpenRouterAPIKey string
	BuildAPIURL      string
	TemporalHost     string
	Port             string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	return &Config{
		AnthropicAPIKey:  apiKey,
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		BuildAPIURL:      envOrDefault("BUILD_API_URL", "http://localhost:8000"),
		TemporalHost:     envOrDefault("TEMPORAL_HOST", "localhost:7233"),
		Port:             envOrDefault("PORT", "8080"),
	}, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
