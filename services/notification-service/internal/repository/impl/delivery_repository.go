package impl

import (
	"context"
	"errors"
	"time"

	"portal-system/services/notification-service/internal/model"

	"gorm.io/gorm"
)

type GormDeliveryRepository struct {
	db *gorm.DB
}

func NewGormDeliveryRepository(db *gorm.DB) *GormDeliveryRepository {
	return &GormDeliveryRepository{db: db}
}

func (r *GormDeliveryRepository) CreateProcessing(ctx context.Context, delivery *model.NotificationDelivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *GormDeliveryRepository) FindByEventID(ctx context.Context, eventID string) (*model.NotificationDelivery, error) {
	var delivery model.NotificationDelivery

	err := r.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		First(&delivery).Error

	if err != nil {
		return nil, err
	}

	return &delivery, nil
}

func (r *GormDeliveryRepository) MarkSent(ctx context.Context, eventID string) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&model.NotificationDelivery{}).
		Where("event_id = ?", eventID).
		Updates(map[string]any{
			"status":     model.DeliveryStatusSent,
			"sent_at":    now,
			"last_error": nil,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *GormDeliveryRepository) MarkFailed(ctx context.Context, eventID string, lastError string) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&model.NotificationDelivery{}).
		Where("event_id = ?", eventID).
		Updates(map[string]any{
			"status":     model.DeliveryStatusFailed,
			"last_error": lastError,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
