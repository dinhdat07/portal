package service

import (
	"context"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
)

type UserNotificationService interface {
	ListNotifications(ctx context.Context, userID uuid.UUID, filter UserNotificationListFilter) (*UserNotificationListResult, error)
	GetNotification(ctx context.Context, id, userID uuid.UUID) (*model.UserNotification, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

type userNotificationService struct {
	notificationRepo repository.UserNotificationRepository
	txManager        repository.TxManager
}

func NewUserNotificationService(
	notificationRepo repository.UserNotificationRepository,
	txManager repository.TxManager,
) UserNotificationService {
	return &userNotificationService{
		notificationRepo: notificationRepo,
		txManager:        txManager,
	}
}

func (s *userNotificationService) ListNotifications(ctx context.Context, userID uuid.UUID, filter UserNotificationListFilter) (*UserNotificationListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	notifications, total, err := s.notificationRepo.FindByUserID(ctx, userID, filter.Page, filter.PageSize, filter.UnreadOnly)
	if err != nil {
		return nil, err
	}

	return &UserNotificationListResult{
		Notifications: notifications,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
	}, nil
}

func (s *userNotificationService) GetNotification(ctx context.Context, id, userID uuid.UUID) (*model.UserNotification, error) {
	notification, err := s.notificationRepo.FindByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if notification == nil {
		return nil, repository.ErrNotFound
	}
	return notification, nil
}

func (s *userNotificationService) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	// Verify exists and owned
	notification, err := s.GetNotification(ctx, id, userID)
	if err != nil {
		return err
	}
	if notification.IsRead {
		return nil // already read
	}

	return s.notificationRepo.MarkAsRead(ctx, id, userID)
}

func (s *userNotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *userNotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}
