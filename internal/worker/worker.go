package outbox

import (
	"context"
	"log"
	"time"

	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type Worker struct {
	txManager repository.TxManager
	repo      repository.OutboxRepository
	publisher Publisher

	workerID  string
	interval  time.Duration
	batchSize int
}

func NewWorker(txManager repository.TxManager, repo repository.OutboxRepository, publisher Publisher, interval time.Duration, batchSize int) *Worker {
	return &Worker{
		txManager: txManager,
		repo:      repo,
		publisher: publisher,
		workerID:  uuid.NewString(),
		interval:  interval,
		batchSize: batchSize,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			if err := w.publishPending(ctx); err != nil {
				log.Printf("outbox worker failed: %v", err)
			}
		}
	}
}

func (w *Worker) publishPending(ctx context.Context) error {
	events, err := w.claimPendingEvents(ctx)
	if err != nil {
		return err
	}

	for _, event := range events {
		if err := w.publisher.Publish(ctx, event.Topic, event.MessageKey, []byte(event.Payload)); err != nil {
			w.handlePublishFailed(ctx, event, err)
			continue
		}

		if err := w.repo.MarkPublished(ctx, event.ID); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) claimPendingEvents(ctx context.Context) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent

	err := w.txManager.WithTx(ctx, func(ctx context.Context) error {
		pendingEvents, err := w.repo.ListPendingForUpdate(ctx, w.batchSize)
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
		return nil, err
	}

	return events, nil
}

func (w *Worker) handlePublishFailed(ctx context.Context, event model.OutboxEvent, publishErr error) {
	retryCount := event.RetryCount + 1

	if retryCount >= event.MaxRetry {
		if err := w.repo.MarkDeadLetter(ctx, event.ID, publishErr.Error()); err != nil {
			log.Printf("mark outbox event dead_letter failed event_id=%s error=%v", event.ID, err)
		}
		return
	}

	if err := w.repo.MarkRetryScheduled(
		ctx,
		event.ID,
		retryCount,
		publishErr.Error(),
		nextRetryAt(retryCount),
	); err != nil {
		log.Printf("mark outbox event retry_scheduled failed event_id=%s error=%v", event.ID, err)
	}
}
