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
		Database: getEnvAsString("POSTGRES_DB", "testdb"),
		User:     getEnvAsString("POSTGRES_USER", "testuser"),
		Password: getEnvAsString("POSTGRES_PASSWORD", "testpass"),
	}
}

func getEnvAsString(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
