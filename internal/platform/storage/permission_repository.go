package storage

import (
	"context"
	"errors"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormPermissionRepository struct {
	db *gorm.DB
}

func NewGormPermissionRepository(db *gorm.DB) *GormPermissionRepository {
	return &GormPermissionRepository{db: db}
}

func (r *GormPermissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var perm models.Permission
	if err := r.getDB(ctx).First(&perm, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}
	return &perm, nil
}

func (r *GormPermissionRepository) List(ctx context.Context) ([]models.Permission, error) {
	var perms []models.Permission
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
