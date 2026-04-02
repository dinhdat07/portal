package repositories

import (
	"context"
	"time"

	"portal-system/internal/domain/enum"
	"portal-system/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserTokenRepository struct {
	db *gorm.DB
}

func NewUserTokenRepository(db *gorm.DB) *UserTokenRepository {
	return &UserTokenRepository{db: db}
}

func (r *UserTokenRepository) WithTx(tx any) *UserTokenRepository {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return r
	}
	return &UserTokenRepository{db: gormTx}
}

func (r *UserTokenRepository) Create(ctx context.Context, token *models.UserToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *UserTokenRepository) FindValidToken(ctx context.Context, tokenHash string, tokenType enum.TokenType) (*models.UserToken, error) {
	var token models.UserToken

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ?", tokenHash).
		Where("token_type = ?", tokenType).
		Where("used_at IS NULL").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", time.Now().UTC()).
		First(&token).Error
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *UserTokenRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.UserToken{}).
		Where("id = ?", id).
		Where("used_at IS NULL").
		Update("used_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.UserToken{}).
		Where("id = ?", id).
		Where("revoked_at IS NULL").
		Update("revoked_at", &now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *UserTokenRepository) RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType enum.TokenType) error {
	now := time.Now().UTC()

	return r.db.WithContext(ctx).
		Model(&models.UserToken{}).
		Where("user_id = ?", userID).
		Where("token_type = ?", tokenType).
		Where("used_at IS NULL").
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		Update("revoked_at", &now).Error
}
