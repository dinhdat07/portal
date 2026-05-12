package repository

import (
	"context"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type AuthSessionRepository interface {
	Create(ctx context.Context, session *model.AuthSession) error
	FindActiveByID(ctx context.Context, id uuid.UUID) (*model.AuthSession, error)
	RevokeByID(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	ListActiveByUserID(ctx context.Context, userID uuid.UUID) ([]model.AuthSession, error)
}
