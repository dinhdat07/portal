package repository

import (
	"context"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type UserNotificationRepository interface {
	BatchCreate(ctx context.Context, notifications []model.UserNotification) error
	FindByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int, unreadOnly bool) ([]model.UserNotification, int64, error)
	FindByIDAndUserID(ctx context.Context, id, userID uuid.UUID) (*model.UserNotification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int64, error)
}
