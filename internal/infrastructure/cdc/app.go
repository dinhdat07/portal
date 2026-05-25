package cdc

import (
	"context"
	"fmt"
	"log"

	"portal-system/config"
	esinfra "portal-system/internal/infrastructure/elasticsearch"
	kafkainfra "portal-system/internal/infrastructure/kafka"
)

func New() (*Consumer, error) {
	ctx := context.Background()

	kafkaCfg, err := config.LoadKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("load kafka config: %w", err)
	}
	esCfg := config.LoadElasticsearchConfig()

	esClient, err := esinfra.Connect(ctx, esCfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect elasticsearch: %w", err)
	}

	userIndexer := esinfra.NewUserIndexer(esClient, esCfg.UserIndex)
	if err := userIndexer.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure user index: %w", err)
	}

	auditLogIndexer := esinfra.NewAuditLogIndexer(esClient, esCfg.AuditIndex)
	if err := auditLogIndexer.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure audit log index: %w", err)
	}

	topics := []string{
		kafkaCfg.AuditTopic,
		kafkaCfg.UserTopic,
	}

	reader := kafkainfra.NewReader(
		kafkaCfg.Brokers,
		topics,
		kafkaCfg.ConsumerGroup,
	)

	router := NewRouterHandler(map[string]TopicHandler{
		kafkaCfg.UserTopic:  NewUserEventHandler(userIndexer),
		kafkaCfg.AuditTopic: NewAuditLogEventHandler(auditLogIndexer),
	})

	log.Printf(
		"[CDC Consumer] Configured: brokers=%v topic=%s group=%s elasticsearch=%s index=%s",
		kafkaCfg.Brokers,
		kafkaCfg.UserTopic,
		kafkaCfg.ConsumerGroup,
		esCfg.URL,
		esCfg.UserIndex,
	)

	return NewConsumer(reader, router), nil
}
