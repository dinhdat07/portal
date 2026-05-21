package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"portal-system/services/notification-service/internal/model"
	"portal-system/services/notification-service/internal/repository"
	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type EmailSender interface {
	Send(ctx context.Context, template string, to string, name string, data map[string]any) error
}

type Config struct {
	FetchRetryInitialBackoff time.Duration
	FetchRetryMaxBackoff     time.Duration
}

type Worker struct {
	reader       *kafka.Reader
	emailSender  EmailSender
	deliveryRepo repository.DeliveryRepository
	cfg          Config
}

func NewWorker(reader *kafka.Reader, emailSender EmailSender, deliveryRepo repository.DeliveryRepository, cfg Config) *Worker {
	if cfg.FetchRetryInitialBackoff <= 0 {
		cfg.FetchRetryInitialBackoff = time.Second
	}
	if cfg.FetchRetryMaxBackoff <= 0 {
		cfg.FetchRetryMaxBackoff = 30 * time.Second
	}

	return &Worker{
		reader:       reader,
		emailSender:  emailSender,
		deliveryRepo: deliveryRepo,
		cfg:          cfg,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	log.Println("email notification worker started; waiting for kafka messages")

	backoff := w.cfg.FetchRetryInitialBackoff
	maxBackoff := w.cfg.FetchRetryMaxBackoff

	for {
		msg, err := w.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				log.Println("email notification worker stopped")
				return nil
			}

			log.Printf("fetch kafka message failed: %v; retrying in %s", err, backoff)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(backoff):
			}

			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			continue
		}

		backoff = w.cfg.FetchRetryInitialBackoff

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
		return fmt.Errorf("%w: unmarshal notification event: %v", ErrNonRetryable, err)
	}

	if err := validateEmailEvent(event); err != nil {
		return fmt.Errorf("%w: validate notification event: %v", ErrNonRetryable, err)
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

	shouldSend, err := w.ensureDelivery(ctx, event)
	if err != nil {
		return err
	}

	if !shouldSend {
		log.Printf("skip duplicate email notification event_id=%s", event.EventID)
		return nil
	}

	if err := w.emailSender.Send(
		ctx,
		event.Template,
		event.Recipient.Email,
		event.Recipient.Name,
		event.Data,
	); err != nil {
		_ = w.deliveryRepo.MarkFailed(ctx, event.EventID, err.Error())
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

	return w.deliveryRepo.MarkSent(ctx, event.EventID)
}

func (w *Worker) ensureDelivery(ctx context.Context, event notificationv1.NotificationRequestedEvent) (bool, error) {
	delivery := &model.NotificationDelivery{
		ID:               uuid.New(),
		EventID:          event.EventID,
		NotificationType: event.NotificationType,
		Channel:          model.DeliveryChannelEmail,
		RecipientEmail:   event.Recipient.Email,
		Template:         event.Template,
		Status:           model.DeliveryStatusProcessing,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	err := w.deliveryRepo.CreateProcessing(ctx, delivery)
	if err == nil {
		return true, nil
	}

	existing, err := w.deliveryRepo.FindByEventID(ctx, event.EventID)
	if err != nil {
		return false, err
	}

	if existing.Status == model.DeliveryStatusSent {
		return false, nil
	}

	return true, nil
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
