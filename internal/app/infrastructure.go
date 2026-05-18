package app

import (
	"portal-system/config"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/service"

	kafkainfra "portal-system/internal/infrastructure/kafka"
	"portal-system/internal/infrastructure/security"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type Infra struct {
	NotificationPublisher service.NotificationPublisher
	TokenManager          service.TokenIssuer
	RevocationStore       service.SessionRevocationStore
}

func newInfra(cfg *config.Config, kafkaWriter *kafka.Writer, kafkaCfg config.KafkaConfig, rdb redis.UniversalClient) *Infra {
	return &Infra{
		NotificationPublisher: kafkainfra.NewNotificationPublisher(kafkaWriter, kafkaCfg.NotificationRequestedTopic),
		TokenManager:          security.New(cfg.JWTSecret, cfg.JWTAccessTTL),
		RevocationStore:       redisx.NewRedisSessionRevocationStore(rdb),
	}
}
