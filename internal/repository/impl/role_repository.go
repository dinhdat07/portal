package impl

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) *GormRoleRepository {
	return &GormRoleRepository{db: db}
}

func (r *GormRoleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.getDB(ctx).Create(role).Error
}

func (r *GormRoleRepository) FindByCode(ctx context.Context, code domain.RoleCode) (*model.Role, error) {
	var role model.Role

	err := r.getDB(ctx).
		Where("code = ?", string(code)).
		First(&role).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Role, error) {
	var role model.Role

	if err := r.getDB(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role

	err := r.getDB(ctx).Preload("Permissions").Order("name ASC").Find(&roles).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return roles, nil
}

func (r *GormRoleRepository) GetWithPermissions(ctx context.Context, roleID uuid.UUID) (*model.Role, error) {
	var role model.Role

	err := r.getDB(ctx).
		Preload("Permissions").
		First(&role, "id = ?", roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *GormRoleRepository) AssignPermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	rp := &model.RolePermission{
		RoleID:       roleID,
		PermissionID: permID,
	}

	return r.getDB(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(rp).Error
}

func (r *GormRoleRepository) RemovePermission(ctx context.Context, roleID uuid.UUID, permID uuid.UUID) error {
	return r.getDB(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(&model.RolePermission{}).Error
}

func (r *GormRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.getDB(ctx).
		Where("id = ?", id).
		Delete(&model.Role{}).Error
}

func (r *GormRoleRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
