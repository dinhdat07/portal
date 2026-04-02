package config

import (
	"errors"
	"os"
	"strconv"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	FromName string

	UseAuth bool
	UseTLS  bool
}

func LoadSMTPConfig() (*SMTPConfig, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if host == "" || port == "" {
		return nil, errors.New("missing SMTP_HOST or SMTP_PORT")
	}

	useAuth := parseBool(os.Getenv("SMTP_USE_AUTH"))
	useTLS := parseBool(os.Getenv("SMTP_USE_TLS"))

	return &SMTPConfig{
		Host:     host,
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     getEnvDefault("SMTP_FROM", "noreply@portal.local"),
		FromName: getEnvDefault("SMTP_FROM_NAME", "Portal System"),
		UseAuth:  useAuth,
		UseTLS:   useTLS,
	}, nil
}

func getEnvDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}

func parseBool(val string) bool {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return b
}
