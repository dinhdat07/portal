package app

import (
	"context"
	"portal-system/config"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/service"

	kafkainfra "portal-system/internal/infrastructure/kafka"
	"portal-system/internal/infrastructure/security"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type KafkaPublisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type Infra struct {
	KafkaPublisher    KafkaPublisher
	NotificationTopic string
	TokenManager      service.TokenIssuer
	RevocationStore   service.SessionRevocationStore
	RedisClient       redis.UniversalClient
}

func newInfra(cfg *config.Config, kafkaWriter *kafka.Writer, kafkaCfg config.KafkaConfig, rdb redis.UniversalClient) *Infra {
	return &Infra{
		KafkaPublisher:    kafkainfra.NewPublisher(kafkaWriter),
		NotificationTopic: kafkaCfg.NotificationRequestedTopic,
		TokenManager:      security.New(cfg.JWTSecret, cfg.JWTAccessTTL),
		RevocationStore:   redisx.NewRedisSessionRevocationStore(rdb),
		RedisClient:       rdb,
	}
}
