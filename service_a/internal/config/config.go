package config

import "os"

// Config holds all configuration values for service_a, loaded from environment variables.
type Config struct {
	Port         string
	ServiceBURL  string
	OTELEndpoint string
	ServiceName  string
}

// Load reads environment variables and returns a Config with defaults applied.
func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		ServiceBURL:  getEnv("SERVICE_B_URL", "http://service_b:8081"),
		OTELEndpoint: getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317"),
		ServiceName:  getEnv("OTEL_SERVICE_NAME", "service_a"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
