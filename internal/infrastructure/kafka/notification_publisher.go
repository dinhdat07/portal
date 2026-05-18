package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/segmentio/kafka-go"
)

type NotificationPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewNotificationPublisher(writer *kafka.Writer, topic string) *NotificationPublisher {
	return &NotificationPublisher{
		writer: writer,
		topic:  topic,
	}
}

func (p *NotificationPublisher) PublishNotificationRequested(ctx context.Context, event notificationv1.NotificationRequestedEvent) error {
	if event.EventID == "" {
		return fmt.Errorf("notificaiton event_id is required")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification requets event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: p.topic,
		Key:   []byte(event.EventID),
		Value: payload,
		Time:  time.Now().UTC(),
	})

	if err != nil {
		return fmt.Errorf("publish notification requested event: %w", err)
	}

	return nil

}
