package repository

import (
	"context"
	"portal-system/internal/model"
	"time"

	"github.com/google/uuid"
)

type OutboxRepository interface {
	Create(ctx context.Context, event *model.OutboxEvent) error

	ListPendingForUpdate(ctx context.Context, limit int) ([]model.OutboxEvent, error)

	MarkPublishing(ctx context.Context, ids []uuid.UUID, workerID string) error
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkRetryScheduled(ctx context.Context, id uuid.UUID, retryCount int, lastError string, nextRetryAt time.Time) error
	MarkDeadLetter(ctx context.Context, id uuid.UUID, lastError string) error
}
