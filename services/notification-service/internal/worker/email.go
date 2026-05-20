package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/segmentio/kafka-go"
)

type EmailSender interface {
	Send(ctx context.Context, template string, to string, name string, data map[string]any) error
}

type Worker struct {
	reader      *kafka.Reader
	emailSender EmailSender
}

func NewWorker(reader *kafka.Reader, emailSender EmailSender) *Worker {
	return &Worker{
		reader:      reader,
		emailSender: emailSender,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	log.Println("email notification worker started; waiting for kafka messages")

	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Println("email notification worker stopped")
				return nil
			}

			return fmt.Errorf("fetch kafka message: %w", err)
		}
		log.Printf(
			"fetched kafka message topic=%s partition=%d offset=%d key=%s bytes=%d",
			msg.Topic,
			msg.Partition,
			msg.Offset,
			string(msg.Key),
			len(msg.Value),
		)

		err = w.handleMessage(ctx, msg)
		if err != nil && !errors.Is(err, ErrNonRetryable) {
			log.Printf(
				"failed to handle email notification topic=%s partition=%d offset=%d error=%v",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err,
			)

			continue
		}

		if err := w.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
		log.Printf(
			"committed kafka message topic=%s partition=%d offset=%d",
			msg.Topic,
			msg.Partition,
			msg.Offset,
		)
	}
}

func (w *Worker) handleMessage(ctx context.Context, msg kafka.Message) error {
	var event notificationv1.NotificationRequestedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal notification event: %w", err)
	}

	if err := validateEmailEvent(event); err != nil {
		return err
	}

	log.Printf(
		"handling notification event event_id=%s notification_type=%s template=%s topic=%s partition=%d offset=%d",
		event.EventID,
		event.NotificationType,
		event.Template,
		msg.Topic,
		msg.Partition,
		msg.Offset,
	)

	if err := w.emailSender.Send(
		ctx,
		event.Template,
		event.Recipient.Email,
		event.Recipient.Name,
		event.Data,
	); err != nil {
		return err
	}

	log.Printf(
		"handled notification event successfully event_id=%s notification_type=%s template=%s topic=%s partition=%d offset=%d",
		event.EventID,
		event.NotificationType,
		event.Template,
		msg.Topic,
		msg.Partition,
		msg.Offset,
	)

	return nil
}

func validateEmailEvent(event notificationv1.NotificationRequestedEvent) error {
	if event.EventID == "" {
		return fmt.Errorf("event_id is required")
	}
	if event.NotificationType == "" {
		return fmt.Errorf("notification_type is required")
	}
	if event.Template == "" {
		return fmt.Errorf("template is required")
	}
	if event.Recipient.Email == "" {
		return fmt.Errorf("recipient email is required")
	}

	return nil
}
