package config

import (
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	APIPort        string
	WebAPICertPath string
	WebAPIKeyPath  string

	JWTPrivateKeyPath string
	JWTPublicKeyPath  string

	DBConnectionURL   string
	DBMaxConnLifetime time.Duration
	DBMaxConnIdleTime time.Duration
	DBHealthCheck     time.Duration
	DBMinConns        int32
	DBMaxConns        int32
}

func Load() *Config {
	return &Config{
		APIPort:           getEnv("API_PORT", "8443"),
		WebAPICertPath:    getEnv("WEB_API_CERT_FILE", "certs/server/server.crt"),
		WebAPIKeyPath:     getEnv("WEB_API_KEY_FILE", "certs/server/server.key"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "certs/jwt_keys/private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "certs/jwt_keys/public.pem"),
		DBConnectionURL:   getEnv("DB_CONNECTION_URL", "postgres://ffuser:ffpass@featureflagsdb:5432/featureflagsdb"),
		DBMaxConnLifetime: getDuration("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		DBMaxConnIdleTime: getDuration("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		DBHealthCheck:     getDuration("DB_HEALTH_CHECK_PERIOD", 1*time.Minute),
		DBMinConns:        getInt32("DB_MIN_CONNS", 1),
		DBMaxConns:        getInt32("DB_MAX_CONNS", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return defaultValue
}

func getInt32(key string, defaultValue int32) int32 {
	if value := os.Getenv(key); value != "" {
		i, err := strconv.Atoi(value)
		if err == nil {
			return int32(i)
		}
	}
	return defaultValue
}
