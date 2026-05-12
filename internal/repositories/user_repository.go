package repositories

import (
	"context"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/models"

	"github.com/google/uuid"
)

type UserListFilter struct {
	Page     int
	PageSize int
	Username string
	Email    string
	FullName string
	Dob      *time.Time
	RoleID   *uuid.UUID

	Status         domain.UserStatus
	IncludeDeleted bool
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error

	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	ListUsers(ctx context.Context, filter UserListFilter) ([]models.User, int64, error)

	Update(ctx context.Context, user *models.User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error

	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error
	Restore(ctx context.Context, id uuid.UUID) error

	ExistsByRoleIDUnscoped(ctx context.Context, roleID uuid.UUID) (bool, error)
	UpdateRoleByRoleIDUnscoped(ctx context.Context, oldRoleID uuid.UUID, newRoleID uuid.UUID) error
}
