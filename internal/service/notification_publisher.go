package service

import (
	"context"

	notificationv1 "portal-system/shared/events/notification/v1"
)

type NotificationPublisher interface {
	PublishNotificationRequested(ctx context.Context, event notificationv1.NotificationRequestedEvent) error
}
