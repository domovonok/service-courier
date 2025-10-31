package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type Config struct {
	Port       string
	PgHost     string
	PgPort     string
	PgDB       string
	PgUser     string
	PgPassword string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:       os.Getenv("PORT"),
		PgHost:     os.Getenv("POSTGRES_HOST"),
		PgPort:     os.Getenv("POSTGRES_PORT"),
		PgDB:       os.Getenv("POSTGRES_DB"),
		PgUser:     os.Getenv("POSTGRES_USER"),
		PgPassword: os.Getenv("POSTGRES_PASSWORD"),
	}

	pflag.StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port to listen on")
	pflag.Parse()

	return cfg
}
