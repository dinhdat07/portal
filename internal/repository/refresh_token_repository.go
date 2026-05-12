package repository

import (
	"context"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID) error
	RevokeBySessionID(ctx context.Context, sessionID uuid.UUID) error
	MarkReplacement(ctx context.Context, id uuid.UUID, replacementID uuid.UUID) error
}
