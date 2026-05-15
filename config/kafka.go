package config

import (
	"os"
	"strings"
)

type KafkaConfig struct {
	Brokers       []string
	UserTopic     string
	AuditTopic    string
	ConsumerGroup string
}

func LoadKafkaConfig() KafkaConfig {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	userTopic := os.Getenv("KAFKA_USER_TOPIC")
	if userTopic == "" {
		userTopic = "portal.public.users"
	}

	auditTopic := os.Getenv("KAFKA_AUDIT_LOG_TOPIC")
	if userTopic == "" {
		userTopic = "portal.public.action_logs"
	}

	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if consumerGroup == "" {
		consumerGroup = "portal-user-search-indexer"
	}

	return KafkaConfig{
		Brokers:       strings.Split(brokers, ","),
		UserTopic:     userTopic,
		AuditTopic:    auditTopic,
		ConsumerGroup: consumerGroup,
	}
}
