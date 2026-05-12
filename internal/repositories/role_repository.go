package repositories

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(ctx context.Context, role *models.Role) error
	FindByCode(ctx context.Context, code domain.RoleCode) (*models.Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	List(ctx context.Context) ([]models.Role, error)
	GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
