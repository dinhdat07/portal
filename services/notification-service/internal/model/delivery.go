package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	DeliveryStatusProcessing = "processing"
	DeliveryStatusSent       = "sent"
	DeliveryStatusFailed     = "failed"

	DeliveryChannelEmail = "email"
)

type NotificationDelivery struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	EventID          string    `gorm:"not null;uniqueIndex"`
	NotificationType string    `gorm:"not null"`
	Channel          string    `gorm:"not null"`
	RecipientEmail   string    `gorm:"not null"`
	Template         string    `gorm:"not null"`

	Status    string `gorm:"not null;index"`
	LastError *string
	SentAt    *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
