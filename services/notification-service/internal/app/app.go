package app

import (
	emailworker "portal-system/services/notification-service/internal/worker"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)

type App struct {
	EmailWorker *emailworker.Worker
	KafkaReader *kafka.Reader
	DB          *gorm.DB
}
