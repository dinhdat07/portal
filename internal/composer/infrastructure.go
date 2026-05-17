package composer

import (
	"portal-system/config"
	"portal-system/internal/infrastructure/email"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/service"

	"portal-system/internal/infrastructure/security"

	"github.com/redis/go-redis/v9"
)

type Infra struct {
	EmailService    service.EmailSender
	TokenManager    service.TokenIssuer
	RevocationStore service.SessionRevocationStore
}

func newInfra(cfg *config.Config, smtpCfg *config.SMTPConfig, rdb redis.UniversalClient) *Infra {
	return &Infra{
		EmailService:    email.NewSMTPEmailService(*smtpCfg),
		TokenManager:    security.New(cfg.JWTSecret, cfg.JWTAccessTTL),
		RevocationStore: redisx.NewRedisSessionRevocationStore(rdb),
	}
}
