package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Port           string
	WebAPICertFile string
	WebAPIKeyFile  string
}

func Load() *Config {
	return &Config{
		Port:           envOrDefault("API_PORT", "8443"),
		WebAPICertFile: envOrDefault("WEB_API_CERT_FILE", "certs/server/server.crt"),
		WebAPIKeyFile:  envOrDefault("WEB_API_KEY_FILE", "certs/server/server.key"),
	}
}

func envOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
