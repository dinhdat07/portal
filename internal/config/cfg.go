package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl     string
	JWTSecret string
	Port      string
}

func Load() (*Config, error) {
	// load .env into os env
	if err := godotenv.Load(); err != nil {
		return nil, errors.New("errors: cannot load env variables")
	}

	return &Config{
		DBUrl:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),
	}, nil
}
