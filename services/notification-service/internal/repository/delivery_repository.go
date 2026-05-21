package repository

import (
	"context"

	"portal-system/services/notification-service/internal/model"
)

type DeliveryRepository interface {
	CreateProcessing(ctx context.Context, delivery *model.NotificationDelivery) error
	FindByEventID(ctx context.Context, eventID string) (*model.NotificationDelivery, error)
	MarkSent(ctx context.Context, eventID string) error
	MarkFailed(ctx context.Context, eventID string, lastError string) error
}
