package impl

import (
	"context"
	"errors"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormPermissionRepository struct {
	db *gorm.DB
}

func NewGormPermissionRepository(db *gorm.DB) *GormPermissionRepository {
	return &GormPermissionRepository{db: db}
}

func (r *GormPermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Permission, error) {
	var perm model.Permission
	if err := r.getDB(ctx).First(&perm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (r *GormPermissionRepository) List(ctx context.Context) ([]model.Permission, error) {
	var perms []model.Permission
	if err := r.getDB(ctx).Order("name ASC").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *GormPermissionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
