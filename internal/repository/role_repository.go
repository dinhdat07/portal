package repository

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(ctx context.Context, role *model.Role) error
	FindByCode(ctx context.Context, code domain.RoleCode) (*model.Role, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error)
	List(ctx context.Context) ([]model.Role, error)
	GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*model.Role, error)
	AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
