package worker

import (
	"context"
	"log/slog"
	"time"

	portalmetrics "portal-system/internal/metrics"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
)

type Config struct {
	Interval            time.Duration
	BatchSize           int
	MaxRetry            int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
	RetryJitterRatio    float64
}

type Publisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type OutboxPublisher struct {
	txManager repository.TxManager
	repo      repository.OutboxRepository
	publisher Publisher
	logger    *slog.Logger
	metrics   portalmetrics.OutboxMetrics

	workerID string
	cfg      Config
}

func NewOutboxPublisher(
	txManager repository.TxManager,
	repo repository.OutboxRepository,
	publisher Publisher,
	logger *slog.Logger,
	metrics portalmetrics.OutboxMetrics,
	cfg Config,
) *OutboxPublisher {
	if logger == nil {
		logger = slog.Default()
	}
	if metrics == nil {
		metrics = portalmetrics.NoopOutboxMetrics{}
	}
	return &OutboxPublisher{
		txManager: txManager,
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		metrics:   metrics,
		workerID:  uuid.NewString(),
		cfg:       cfg,
	}
}

func (w *OutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := w.publishPending(ctx); err != nil {
				w.logger.Error("outbox worker failed", slog.Any("error", err))
			}
		}
	}
}

func (w *OutboxPublisher) publishPending(ctx context.Context) error {
	events, err := w.claimPendingEvents(ctx)
	if err != nil {
		return err
	}

	w.metrics.EventsClaimed(len(events))

	for _, event := range events {
		w.logger.Info("outbox_event_claimed",
			slog.String("outbox_event_id", event.ID.String()),
			slog.String("topic", event.Topic),
			slog.String("key", event.MessageKey),
			slog.Int("retry_count", event.RetryCount),
			slog.String("worker_id", w.workerID),
		)

		if err := w.publisher.Publish(ctx, event.Topic, event.MessageKey, []byte(event.Payload)); err != nil {
			w.handlePublishFailed(ctx, event, err)
			continue
		}

		if err := w.repo.MarkPublished(ctx, event.ID); err != nil {
			return err
		}

		w.metrics.EventsPublished()

		w.logger.Info("outbox_event_published",
			slog.String("outbox_event_id", event.ID.String()),
			slog.String("topic", event.Topic),
			slog.String("key", event.MessageKey),
		)
	}

	return nil
}

func (w *OutboxPublisher) claimPendingEvents(ctx context.Context) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent

	err := w.txManager.WithTx(ctx, func(ctx context.Context) error {
		pendingEvents, err := w.repo.ListPendingForUpdate(ctx, w.cfg.BatchSize)
		if err != nil {
			return err
		}

		if len(pendingEvents) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, 0, len(pendingEvents))
		for _, event := range pendingEvents {
			ids = append(ids, event.ID)
		}

		if err := w.repo.MarkPublishing(ctx, ids, w.workerID); err != nil {
			return err
		}

		events = pendingEvents
		return nil
	})

	if err != nil {
		w.logger.Error("outbox_events_claim_failed",
			slog.String("worker_id", w.workerID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return events, nil
}

func (w *OutboxPublisher) handlePublishFailed(ctx context.Context, event model.OutboxEvent, publishErr error) {
	w.metrics.EventsPublishFailed()
	retryCount := event.RetryCount + 1

	maxRetry := event.MaxRetry
	if maxRetry <= 0 {
		maxRetry = w.cfg.MaxRetry
	}

	if retryCount >= maxRetry {
		if err := w.repo.MarkDeadLetter(ctx, event.ID, publishErr.Error()); err != nil {
			w.logger.Error("mark outbox event dead_letter failed",
				slog.String("outbox_event_id", event.ID.String()),
				slog.Any("error", err),
			)
		} else {
			w.metrics.EventsDeadLettered()
		}
		w.logger.Error("outbox_event_dead_lettered",
			slog.String("outbox_event_id", event.ID.String()),
			slog.String("topic", event.Topic),
			slog.String("key", event.MessageKey),
			slog.Int("retry_count", retryCount),
			slog.String("last_error", publishErr.Error()),
		)
		return
	}

	nextAt := nextRetryAt(retryCount, w.cfg)
	if err := w.repo.MarkRetryScheduled(
		ctx,
		event.ID,
		retryCount,
		publishErr.Error(),
		nextAt,
	); err != nil {
		w.logger.Error("mark outbox event retry_scheduled failed",
			slog.String("outbox_event_id", event.ID.String()),
			slog.Any("error", err),
		)
	} else {
		w.metrics.EventsRetryScheduled()
	}
	w.logger.Warn("outbox_event_publish_failed",
		slog.String("outbox_event_id", event.ID.String()),
		slog.String("topic", event.Topic),
		slog.String("key", event.MessageKey),
		slog.Int("retry_count", retryCount),
		slog.Time("next_retry_at", nextAt),
		slog.Any("error", publishErr),
	)
}
