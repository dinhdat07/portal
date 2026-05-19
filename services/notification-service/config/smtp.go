package config

import (
	"os"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     string
	UseAuth  bool
	UseTLS   bool
	Username string
	Password string
	From     string
	FromName string
}

func LoadSMTPConfig() SMTPConfig {
	return SMTPConfig{
		Host:     getEnv("SMTP_HOST", "localhost"),
		Port:     getEnv("SMTP_PORT", "1025"),
		UseAuth:  getEnv("SMTP_USE_AUTH", "false") == "true",
		UseTLS:   getEnv("SMTP_USE_TLS", "false") == "true",
		Username: getEnv("SMTP_USERNAME", ""),
		Password: getEnv("SMTP_PASSWORD", ""),
		From:     getEnv("SMTP_FROM", "noreply@portal.local"),
		FromName: getEnv("SMTP_FROM_NAME", "Portal System"),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
