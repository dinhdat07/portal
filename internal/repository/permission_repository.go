package repository

import (
	"context"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type PermissionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*model.Permission, error)
	List(ctx context.Context) ([]model.Permission, error)
}
