package composer

import (
	"portal-system/config"
	"portal-system/internal/platform/email"
	redisx "portal-system/internal/platform/redis"
	"portal-system/internal/services"

	token "portal-system/internal/platform/security"

	"github.com/redis/go-redis/v9"
)

type Infra struct {
	EmailService    services.EmailSender
	TokenManager    services.TokenIssuer
	RevocationStore services.SessionRevocationStore
}

func newInfra(cfg *config.Config, smtpCfg *config.SMTPConfig, rdb redis.UniversalClient) *Infra {
	return &Infra{
		EmailService:    email.NewSMTPEmailService(*smtpCfg),
		TokenManager:    token.New(cfg.JWTSecret, cfg.JWTAccessTTL),
		RevocationStore: redisx.NewRedisSessionRevocationStore(rdb),
	}
}
