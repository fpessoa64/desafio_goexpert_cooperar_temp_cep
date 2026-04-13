package config

import (
	"errors"
	"os"
)

// Config holds all configuration values for service_b, loaded from environment variables.
type Config struct {
	Port          string
	WeatherAPIKey string
	OTELEndpoint  string
	ServiceName   string
}

// Load reads environment variables and returns a Config with defaults applied.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		return nil, errors.New("WEATHER_API_KEY is required")
	}

	return &Config{
		Port:          getEnv("PORT", "8081"),
		WeatherAPIKey: apiKey,
		OTELEndpoint:  getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		ServiceName:   getEnv("OTEL_SERVICE_NAME", "service_b"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
