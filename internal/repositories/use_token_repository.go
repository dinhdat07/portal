package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type UserTokenRepository interface {
	Create(ctx context.Context, token *models.UserToken) error
	FindValidToken(ctx context.Context, tokenHash string, tokenType domain.TokenType) (*models.UserToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error
}
