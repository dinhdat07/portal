package repositories

import (
	"context"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type PermissionRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	List(ctx context.Context) ([]models.Permission, error)
}
