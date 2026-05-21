package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"portal-system/services/notification-service/internal/model"
	"portal-system/services/notification-service/internal/repository"
	notificationv1 "portal-system/shared/events/notification/v1"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"gorm.io/datatypes"
)

type EmailSender interface {
	Send(ctx context.Context, template string, to string, name string, data map[string]any) error
}

type Config struct {
	FetchRetryInitialBackoff    time.Duration
	FetchRetryMaxBackoff        time.Duration
	MaxRetry                    int
	DeliveryRetryInitialBackoff time.Duration
	DeliveryRetryMaxBackoff     time.Duration
	DeliveryRetryJitterRatio    float64
}

type Worker struct {
	reader       *kafka.Reader
	emailSender  EmailSender
	deliveryRepo repository.DeliveryRepository
	txManager    repository.TxManager
	cfg          Config
}

func NewWorker(reader *kafka.Reader, emailSender EmailSender, txManager repository.TxManager, deliveryRepo repository.DeliveryRepository, cfg Config) *Worker {
	if cfg.FetchRetryInitialBackoff <= 0 {
		cfg.FetchRetryInitialBackoff = time.Second
	}
	if cfg.FetchRetryMaxBackoff <= 0 {
		cfg.FetchRetryMaxBackoff = 30 * time.Second
	}
	if cfg.MaxRetry <= 0 {
		cfg.MaxRetry = 3
	}
	if cfg.DeliveryRetryInitialBackoff <= 0 {
		cfg.DeliveryRetryInitialBackoff = 30 * time.Second
	}
	if cfg.DeliveryRetryMaxBackoff <= 0 {
		cfg.DeliveryRetryMaxBackoff = 30 * time.Minute
	}
	if cfg.DeliveryRetryJitterRatio < 0 {
		cfg.DeliveryRetryJitterRatio = 0.2
	}

	return &Worker{
		reader:       reader,
		emailSender:  emailSender,
		deliveryRepo: deliveryRepo,
		txManager:    txManager,
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
		log.Printf(
			"skip notification event event_id=%s notification_type=%s reason=duplicate_or_terminal_state",
			event.EventID,
			event.NotificationType,
		)
		return nil
	}

	if isExpired(event.ValidUntil) {
		if err := w.deliveryRepo.MarkExpired(
			ctx,
			event.EventID,
			"delivery expired before first send",
		); err != nil {
			return fmt.Errorf("mark delivery expired: %w", err)
		}

		log.Printf(
			"skip expired notification event event_id=%s notification_type=%s",
			event.EventID,
			event.NotificationType,
		)
		return nil
	}

	if err := w.emailSender.Send(
		ctx,
		event.Template,
		event.Recipient.Email,
		event.Recipient.Name,
		event.Data,
	); err != nil {
		if updateErr := w.handleSendFailure(ctx, event.EventID, err.Error()); updateErr != nil {
			return fmt.Errorf("send email failed: %w; update delivery retry state failed: %v", err, updateErr)
		}

		log.Printf(
			"scheduled retry or finalized failed notification event_id=%s notification_type=%s error=%v",
			event.EventID,
			event.NotificationType,
			err,
		)

		return nil
	}

	if err := w.deliveryRepo.MarkSent(ctx, event.EventID); err != nil {
		return fmt.Errorf("mark delivery sent: %w", err)
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

func (w *Worker) ensureDelivery(ctx context.Context, event notificationv1.NotificationRequestedEvent) (bool, error) {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return false, fmt.Errorf("%w: marshal delivery data: %v", ErrNonRetryable, err)
	}

	now := time.Now().UTC()

	delivery := &model.NotificationDelivery{
		ID:               uuid.New(),
		EventID:          event.EventID,
		BusinessKey:      event.BusinessKey,
		NotificationType: event.NotificationType,
		Channel:          model.DeliveryChannelEmail,
		RecipientEmail:   event.Recipient.Email,
		RecipientName:    event.Recipient.Name,
		Template:         event.Template,
		Data:             datatypes.JSON(data),
		Status:           model.DeliveryStatusProcessing,
		MaxRetry:         w.cfg.MaxRetry,
		ValidUntil:       event.ValidUntil,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	var shouldSend bool

	err = w.txManager.WithTx(ctx, func(ctx context.Context) error {
		err := w.deliveryRepo.CreateProcessing(ctx, delivery)
		if err == nil {
			if err := w.deliveryRepo.SupersedeRetryableByBusinessKey(
				ctx,
				event.BusinessKey,
				event.EventID,
				"superseded by newer notification event",
			); err != nil {
				return fmt.Errorf("supersede previous retryable deliveries: %w", err)
			}

			shouldSend = true
			return nil
		}

		if !errors.Is(err, repository.ErrDuplicateDelivery) {
			return fmt.Errorf("create notification delivery: %w", err)
		}

		existing, err := w.deliveryRepo.FindByEventID(ctx, event.EventID)
		if err != nil {
			return fmt.Errorf("find existing notification delivery: %w", err)
		}

		switch existing.Status {
		case model.DeliveryStatusSent,
			model.DeliveryStatusProcessing,
			model.DeliveryStatusRetryScheduled,
			model.DeliveryStatusDeadLetter,
			model.DeliveryStatusExpired,
			model.DeliveryStatusSuperseded:
			shouldSend = false
			return nil

		default:
			return fmt.Errorf("unknown delivery status: %s", existing.Status)
		}
	})

	if err != nil {
		return false, err
	}

	return shouldSend, nil
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

func (w *Worker) handleSendFailure(ctx context.Context, eventID string, lastError string) error {
	delivery, err := w.deliveryRepo.FindByEventID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("find delivery before retry scheduling: %w", err)
	}

	if isExpired(delivery.ValidUntil) {
		return w.deliveryRepo.MarkExpired(ctx, eventID, lastError)
	}

	nextRetryCount := delivery.RetryCount + 1
	if nextRetryCount >= delivery.MaxRetry {
		return w.deliveryRepo.MarkDeadLetter(ctx, eventID, lastError)
	}

	nextAt := nextRetryAt(
		nextRetryCount,
		w.cfg.DeliveryRetryInitialBackoff,
		w.cfg.DeliveryRetryMaxBackoff,
		w.cfg.DeliveryRetryJitterRatio,
	)

	if delivery.ValidUntil != nil && !nextAt.Before(*delivery.ValidUntil) {
		return w.deliveryRepo.MarkExpired(ctx, eventID, lastError)
	}

	return w.deliveryRepo.ScheduleRetry(ctx, eventID, nextRetryCount, lastError, nextAt)
}

func isExpired(validUntil *time.Time) bool {
	if validUntil == nil {
		return false
	}

	return !time.Now().UTC().Before(validUntil.UTC())
}

func nextRetryAt(
	retryCount int,
	initialBackoff time.Duration,
	maxBackoff time.Duration,
	jitterRatio float64,
) time.Time {
	delay := initialBackoff
	if delay <= 0 {
		delay = 30 * time.Second
	}

	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Minute
	}

	for i := 1; i < retryCount; i++ {
		delay *= 2
		if delay > maxBackoff {
			delay = maxBackoff
			break
		}
	}

	delay = applyJitter(delay, jitterRatio)

	return time.Now().UTC().Add(delay)
}

func applyJitter(delay time.Duration, ratio float64) time.Duration {
	if ratio <= 0 {
		return delay
	}

	if ratio > 1 {
		ratio = 1
	}

	delta := float64(delay) * ratio
	minDelay := float64(delay) - delta
	maxDelay := float64(delay) + delta

	jittered := minDelay + rand.Float64()*(maxDelay-minDelay)
	if jittered < float64(time.Millisecond) {
		return time.Millisecond
	}

	return time.Duration(jittered)
}
