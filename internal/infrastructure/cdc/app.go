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
	kafkaCfg := config.LoadKafkaConfig()
	esCfg := config.LoadElasticsearchConfig()

	esClient, err := esinfra.Connect(ctx, esCfg.URL)
	if err != nil {
		return nil, fmt.Errorf("connect elasticsearch: %w", err)
	}

	userIndexer := esinfra.NewUserIndexer(esClient, esCfg.UserIndex)

	if err := userIndexer.EnsureIndex(ctx); err != nil {
		return nil, fmt.Errorf("ensure user index: %w", err)
	}

	reader := kafkainfra.NewReader(
		kafkaCfg.Brokers,
		kafkaCfg.UserTopic,
		kafkaCfg.ConsumerGroup,
	)

	handler := NewUserEventHandler(userIndexer)

	log.Printf(
		"[CDC Consumer] Configured: brokers=%v topic=%s group=%s elasticsearch=%s index=%s",
		kafkaCfg.Brokers,
		kafkaCfg.UserTopic,
		kafkaCfg.ConsumerGroup,
		esCfg.URL,
		esCfg.UserIndex,
	)

	return NewConsumer(reader, handler), nil
}
