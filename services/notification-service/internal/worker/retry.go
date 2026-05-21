package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"portal-system/services/notification-service/internal/model"
	"portal-system/services/notification-service/internal/repository"
)

type RetryWorkerConfig struct {
	Interval                    time.Duration
	BatchSize                   int
	MaxRetry                    int
	DeliveryRetryInitialBackoff time.Duration
	DeliveryRetryMaxBackoff     time.Duration
	DeliveryRetryJitterRatio    float64
}

type RetryWorker struct {
	emailSender  EmailSender
	deliveryRepo repository.DeliveryRepository
	cfg          RetryWorkerConfig
}

func NewRetryWorker(emailSender EmailSender, deliveryRepo repository.DeliveryRepository, cfg RetryWorkerConfig) *RetryWorker {
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 10
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

	return &RetryWorker{
		emailSender:  emailSender,
		deliveryRepo: deliveryRepo,
		cfg:          cfg,
	}
}

func (w *RetryWorker) Run(ctx context.Context) error {
	log.Printf("retry worker started; checking for retryable deliveries every %s (batch size: %d)", w.cfg.Interval, w.cfg.BatchSize)

	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("retry worker stopped by context cancellation")
			return nil
		case <-ticker.C:
			if err := w.processRetryQueue(ctx); err != nil {
				log.Printf("failed to process retry queue: %v", err)
			}
		}
	}
}

func (w *RetryWorker) processRetryQueue(ctx context.Context) error {
	deliveries, err := w.deliveryRepo.ClaimRetryDue(ctx, w.cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("claim retry due: %w", err)
	}

	if len(deliveries) == 0 {
		return nil
	}

	log.Printf("claimed %d deliveries for retry", len(deliveries))

	for _, delivery := range deliveries {
		if err := w.processDelivery(ctx, delivery); err != nil {
			log.Printf("failed to retry delivery event_id=%s error=%v", delivery.EventID, err)
		}
	}

	return nil
}

func (w *RetryWorker) processDelivery(ctx context.Context, delivery model.NotificationDelivery) error {
	if isExpired(delivery.ValidUntil) {
		log.Printf("delivery event_id=%s has expired; marking expired", delivery.EventID)
		return w.deliveryRepo.MarkExpired(ctx, delivery.EventID, "delivery expired before retry attempt")
	}

	var data map[string]any
	if err := json.Unmarshal(delivery.Data, &data); err != nil {
		log.Printf("failed to unmarshal delivery data for event_id=%s: %v; moving to dead letter", delivery.EventID, err)
		return w.deliveryRepo.MarkDeadLetter(ctx, delivery.EventID, fmt.Sprintf("unmarshal data: %v", err))
	}

	log.Printf("retrying delivery event_id=%s count=%d recipient=%s", delivery.EventID, delivery.RetryCount, delivery.RecipientEmail)

	if err := w.emailSender.Send(
		ctx,
		delivery.Template,
		delivery.RecipientEmail,
		delivery.RecipientName,
		data,
	); err != nil {
		if updateErr := w.handleRetryFailure(ctx, delivery, err.Error()); updateErr != nil {
			return fmt.Errorf("send email retry failed: %w; update delivery retry state failed: %v", err, updateErr)
		}
		return err
	}

	log.Printf("retry delivery successful event_id=%s", delivery.EventID)
	return w.deliveryRepo.MarkSent(ctx, delivery.EventID)
}

func (w *RetryWorker) handleRetryFailure(ctx context.Context, delivery model.NotificationDelivery, lastError string) error {
	if isExpired(delivery.ValidUntil) {
		return w.deliveryRepo.MarkExpired(ctx, delivery.EventID, lastError)
	}

	nextRetryCount := delivery.RetryCount + 1
	if nextRetryCount >= delivery.MaxRetry {
		return w.deliveryRepo.MarkDeadLetter(ctx, delivery.EventID, lastError)
	}

	nextAt := nextRetryAt(
		nextRetryCount,
		w.cfg.DeliveryRetryInitialBackoff,
		w.cfg.DeliveryRetryMaxBackoff,
		w.cfg.DeliveryRetryJitterRatio,
	)

	if delivery.ValidUntil != nil && !nextAt.Before(*delivery.ValidUntil) {
		return w.deliveryRepo.MarkExpired(ctx, delivery.EventID, lastError)
	}

	return w.deliveryRepo.ScheduleRetry(ctx, delivery.EventID, nextRetryCount, lastError, nextAt)
}
