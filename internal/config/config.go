package config

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/pflag"
)

type Config struct {
	Port string
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port: os.Getenv("PORT"),
	}

	pflag.StringVarP(&cfg.Port, "port", "p", cfg.Port, "Port to listen on")
	pflag.Parse()

	return cfg
}
