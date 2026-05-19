package service

import (
	"context"
	"strings"
	"time"

	"portal-system/internal/model"
	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/google/uuid"
)

type NotificationPublisher interface {
	PublishNotificationRequested(ctx context.Context, event notificationv1.NotificationRequestedEvent) error
}

func newEmailNotificationEvent(notificationType string, template string, user model.User, url string) notificationv1.NotificationRequestedEvent {
	data := map[string]any{
		"username": user.Username,
		"url":      url,
	}
	return notificationv1.NotificationRequestedEvent{
		EventID:          uuid.NewString(),
		OccurredAt:       time.Now().UTC(),
		NotificationType: notificationType,
		Recipient: notificationv1.Recipient{
			UserID: user.ID.String(),
			Email:  user.Email,
			Name:   strings.TrimSpace(user.FirstName + " " + user.LastName),
		},
		Template: template,
		Data:     data,
	}
}
