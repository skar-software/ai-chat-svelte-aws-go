package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port             string
	AppEnv           string
	OpenAIAPIKey     string
	CORSAllowOrigins string
}

func Load() (*Config, error) {
	c := &Config{
		Port:             getEnv("PORT", "8080"),
		AppEnv:           getEnv("APP_ENV", "development"),
		OpenAIAPIKey:     os.Getenv("OPENAI_API_KEY"),
		CORSAllowOrigins: getEnv("CORS_ALLOW_ORIGINS", "http://localhost:5173"),
	}

	if c.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	return c, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return strings.TrimSpace(v)
	}
	return defaultVal
}
