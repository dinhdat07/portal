package v1

import "time"

const (
	EventTypeNotificationRequested = "notification.requested"

	ChannelEmail = "email"

	NotificationTypeVerifyEmail   = "verify_email"
	NotificationTypeResetPassword = "reset_password"
	NotificationTypeSetPassword   = "set_password"
)

type NotificationRequestedEvent struct {
	EventID          string         `json:"event_id"`
	EventType        string         `json:"event_type"`
	OccurredAt       time.Time      `json:"occurred_at"`
	NotificationType string         `json:"notification_type"`
	Recipient        Recipient      `json:"recipient"`
	Channels         []string       `json:"channels"`
	Template         string         `json:"template"`
	Data             map[string]any `json:"data"`
	Metadata         Metadata       `json:"metadata"`
}

type Recipient struct {
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
}

type Metadata struct {
	Source        string `json:"source"`
	CorrelationID string `json:"correlation_id,omitempty"`
}
