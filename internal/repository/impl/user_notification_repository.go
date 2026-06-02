package impl

import (
	"context"
	"portal-system/internal/model"
	"portal-system/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormUserNotificationRepository struct {
	db *gorm.DB
}

func NewGormUserNotificationRepository(db *gorm.DB) repository.UserNotificationRepository {
	return &GormUserNotificationRepository{db: db}
}

func (r *GormUserNotificationRepository) BatchCreate(ctx context.Context, notifications []model.UserNotification) error {
	if len(notifications) == 0 {
		return nil
	}
	// OnConflict DoNothing to avoid constraint violation if already exists
	return r.getDB(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(notifications, 100).Error
}

func (r *GormUserNotificationRepository) FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, unreadOnly bool) ([]model.UserNotification, int64, error) {
	var notifications []model.UserNotification
	var total int64

	query := r.getDB(ctx).Model(&model.UserNotification{}).Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Preload("Announcement").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *GormUserNotificationRepository) FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*model.UserNotification, error) {
	var notification model.UserNotification
	err := r.getDB(ctx).Preload("Announcement").First(&notification, "id = ? AND user_id = ?", id, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &notification, nil
}

func (r *GormUserNotificationRepository) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	now := time.Now()
	return r.getDB(ctx).Model(&model.UserNotification{}).
		Where("id = ? AND user_id = ? AND is_read = ?", id, userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *GormUserNotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.getDB(ctx).Model(&model.UserNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (r *GormUserNotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.getDB(ctx).Model(&model.UserNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *GormUserNotificationRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
