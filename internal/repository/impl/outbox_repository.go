package impl

import (
	"context"
	"portal-system/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormOutboxRepository struct {
	db *gorm.DB
}

func NewGormOutboxRepository(db *gorm.DB) *GormOutboxRepository {
	return &GormOutboxRepository{db: db}
}

func (r *GormOutboxRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func (r *GormOutboxRepository) Create(ctx context.Context, event *model.OutboxEvent) error {
	return r.getDB(ctx).Create(event).Error
}

func (r *GormOutboxRepository) ListPendingForUpdate(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	now := time.Now().UTC()

	err := r.getDB(ctx).Where("status IN ?", []string{
		model.OutboxStatusPending,
		model.OutboxStatusRetryScheduled,
	}).Where("next_retry_at IS NULL OR next_retry_at <= ?", now).
		Limit(limit).
		Order("created_at ASC").
		Clauses(clause.Locking{
			Strength: "Update",
			Options:  "SKIP LOCKED",
		}).Find(&events).Error

	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *GormOutboxRepository) MarkPublishing(ctx context.Context, ids []uuid.UUID, workerID string) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.OutboxEvent{}).
		Where("id IN ?", ids).
		Where("status IN ?", []string{
			model.OutboxStatusPending,
			model.OutboxStatusRetryScheduled,
		}).
		Updates(map[string]any{
			"status":     model.OutboxStatusPublishing,
			"claimed_at": now,
			"claimed_by": workerID,
			"updated_at": now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *GormOutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Where("status = ?", model.OutboxStatusPublishing).
		Updates(map[string]any{
			"status":        model.OutboxStatusPublished,
			"updated_at":    now,
			"published_at":  now,
			"last_error":    nil,
			"next_retry_at": nil,
			"claimed_at":    nil,
			"claimed_by":    nil,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *GormOutboxRepository) MarkRetryScheduled(ctx context.Context, id uuid.UUID, retryCount int, lastError string, nextRetryAt time.Time) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Where("status = ?", model.OutboxStatusPublishing).
		Updates(map[string]any{
			"status":        model.OutboxStatusRetryScheduled,
			"updated_at":    now,
			"retry_count":   retryCount,
			"last_error":    lastError,
			"next_retry_at": nextRetryAt,
			"claimed_at":    nil,
			"claimed_by":    nil,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *GormOutboxRepository) MarkDeadLetter(
	ctx context.Context,
	id uuid.UUID,
	lastError string,
) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Where("status = ?", model.OutboxStatusPublishing).
		Updates(map[string]any{
			"status":        model.OutboxStatusDeadLetter,
			"updated_at":    now,
			"last_error":    lastError,
			"next_retry_at": nil,
			"claimed_at":    nil,
			"claimed_by":    nil,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
