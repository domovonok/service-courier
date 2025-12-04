package testdb

import (
	"os"

	"github.com/joho/godotenv"
)

type TestDBConfig struct {
	Database string
	User     string
	Password string
}

func LoadTestConfig() *TestDBConfig {
	_ = godotenv.Load("tests/.env.test")

	return &TestDBConfig{
		Database: getEnvOrDefault("POSTGRES_DB", "testdb"),
		User:     getEnvOrDefault("POSTGRES_USER", "testuser"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "testpass"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
