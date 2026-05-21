package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"portal-system/internal/model"
	"portal-system/internal/repository"
	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

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

func createNotificationOutboxEvent(
	ctx context.Context,
	repo repository.OutboxRepository,
	topic string,
	event notificationv1.NotificationRequestedEvent,
) error {
	if repo == nil {
		return fmt.Errorf("outbox repository is required")
	}
	if topic == "" {
		return fmt.Errorf("notification topic is required")
	}
	if event.EventID == "" {
		return fmt.Errorf("notification event_id is required")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification requested event: %w", err)
	}

	return repo.Create(ctx, &model.OutboxEvent{
		ID:         uuid.New(),
		Topic:      topic,
		MessageKey: event.EventID,
		Payload:    datatypes.JSON(payload),
		Status:     model.OutboxStatusPending,
		MaxRetry:   getMaxRetryFromEnv(),
	})
}

func getMaxRetryFromEnv() int {
	val := os.Getenv("OUTBOX_WORKER_MAX_RETRY")
	if val == "" {
		return 10
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 10
	}
	return parsed
}

