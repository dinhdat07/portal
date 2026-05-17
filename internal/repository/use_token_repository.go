package repository

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type UserTokenRepository interface {
	Create(ctx context.Context, token *model.UserToken) error
	FindValidToken(ctx context.Context, tokenHash string, tokenType domain.TokenType) (*model.UserToken, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByUserAndType(ctx context.Context, userID uuid.UUID, tokenType domain.TokenType) error
}
