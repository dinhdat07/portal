package impl

import (
	"context"
	"errors"
	"portal-system/internal/model"
	"portal-system/internal/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormAuthSessionRepository struct {
	db *gorm.DB
}

func NewGormAuthSessionRepository(db *gorm.DB) *GormAuthSessionRepository {
	return &GormAuthSessionRepository{db: db}
}

func (r *GormAuthSessionRepository) Create(ctx context.Context, session *model.AuthSession) error {
	return r.getDB(ctx).Create(session).Error
}

func (r *GormAuthSessionRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (*model.AuthSession, error) {
	var AuthSession model.AuthSession

	err := r.getDB(ctx).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&AuthSession).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &AuthSession, nil
}

func (r *GormAuthSessionRepository) RevokeByID(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.AuthSession{}).
		Where("id = ?", sessionID).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *GormAuthSessionRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.AuthSession{}).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *GormAuthSessionRepository) ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.AuthSession, error) {
	now := time.Now().UTC()

	var sessions []model.AuthSession
	result := r.getDB(ctx).
		Model(&model.AuthSession{}).
		Where("user_id = ?", userID).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Find(&sessions)
	if result.Error != nil {
		return nil, result.Error
	}

	return sessions, nil
}

func (r *GormAuthSessionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
