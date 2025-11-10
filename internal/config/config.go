package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type PoolConfig struct {
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	MaxConnIdleTime       time.Duration
	MaxConns              int32
	MinConns              int32
	MinIdleConns          int32
	HealthCheckPeriod     time.Duration
	PingMaxRetries        int
	PingRetryDelay        time.Duration
}

type DBConfig struct {
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPassword string
	Pool       PoolConfig
}

type Config struct {
	Port string
	DB   DBConfig
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port: os.Getenv("PORT"),
		DB: DBConfig{
			PgHost:     os.Getenv("POSTGRES_HOST"),
			PgPort:     os.Getenv("POSTGRES_PORT"),
			PgDB:       os.Getenv("POSTGRES_DB"),
			PgUser:     os.Getenv("POSTGRES_USER"),
			PgPassword: os.Getenv("POSTGRES_PASSWORD"),
			Pool: PoolConfig{
				MaxConns:          getEnvAsInt32("POSTGRES_MAX_CONNS", 20),
				MinConns:          getEnvAsInt32("POSTGRES_MIN_CONNS", 5),
				MaxConnLifetime:   getEnvAsDuration("POSTGRES_MAX_CONN_LIFETIME", time.Hour),
				MaxConnIdleTime:   getEnvAsDuration("POSTGRES_MAX_CONN_IDLE_TIME", 30*time.Minute),
				HealthCheckPeriod: getEnvAsDuration("POSTGRES_HEALTH_CHECK_PERIOD", time.Minute),

				PingMaxRetries: getEnvAsInt("POSTGRES_MAX_RETRIES", 5),
				PingRetryDelay: getEnvAsDuration("POSTGRES_RETRY_DELAY", time.Second),
			},
		},
	}

	pflag.StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port to listen on")
	pflag.Parse()

	return cfg
}

func getEnvAs[T any](key string, defaultVal T, parse func(string) (T, error)) T {
	if value := os.Getenv(key); value != "" {
		if v, err := parse(value); err == nil {
			return v
		}
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	return getEnvAs[int](key, defaultVal, strconv.Atoi)
}

func getEnvAsInt32(key string, defaultVal int32) int32 {
	return getEnvAs[int32](key, defaultVal, func(s string) (int32, error) {
		v, err := strconv.ParseInt(s, 10, 32)
		return int32(v), err
	})
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	return getEnvAs(key, defaultVal, time.ParseDuration)
}
