package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

const (
	OutboxStatusPending        = "pending"
	OutboxStatusPublishing     = "publishing"
	OutboxStatusRetryScheduled = "retry_scheduled"
	OutboxStatusPublished      = "published"
	OutboxStatusDeadLetter     = "dead_letter"
)

type OutboxEvent struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Topic      string         `gorm:"not null"`
	MessageKey string         `gorm:"not null"`
	Payload    datatypes.JSON `gorm:"type:jsonb;not null"`

	Status     string `gorm:"not null;index"`
	RetryCount int    `gorm:"not null;default:0"`
	MaxRetry   int    `gorm:"not null;default:10"`

	LastError   *string
	NextRetryAt *time.Time `gorm:"index"`
	ClaimedAt   *time.Time
	ClaimedBy   *string
	PublishedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
