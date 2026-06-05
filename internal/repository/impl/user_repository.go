package impl

import (
	"context"
	"errors"
	"strings"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *model.User) error {
	user.Email = normalizeEmail(user.Email)
	user.Username = normalizeUsername(user.Username)

	return r.getDB(ctx).Create(user).Error
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.getDB(ctx).Preload("Role").First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) ExistsByRoleIDUnscoped(ctx context.Context, roleID uuid.UUID) (bool, error) {
	var user model.User
	err := r.getDB(ctx).Unscoped().First(&user, "role_id = ?", roleID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *GormUserRepository) FindByIDUnscoped(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.getDB(ctx).Preload("Role").Unscoped().First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	username = normalizeUsername(username)

	err := r.getDB(ctx).Preload("Role").
		Where("LOWER(TRIM(username)) = ?", username).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	email = normalizeEmail(email)

	err := r.getDB(ctx).Preload("Role").
		Where("LOWER(TRIM(email)) = ?", email).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, err
}

func (r *GormUserRepository) Update(ctx context.Context, user *model.User) error {
	user.Username = normalizeUsername(user.Username)

	return r.getDB(ctx).
		Model(user).
		Updates(map[string]interface{}{
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"dob":        user.DOB,
			"username":   user.Username,
		}).Error
}

func (r *GormUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	var user model.User
	result := r.getDB(ctx).
		Model(&user).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateRole(ctx context.Context, id uuid.UUID, roleID uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Update("role_id", roleID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) UpdateRoleByRoleIDUnscoped(ctx context.Context, oldRoleID uuid.UUID, newRoleID uuid.UUID) error {
	result := r.getDB(ctx).
		Unscoped().
		Model(&model.User{}).
		Where("role_id = ?", oldRoleID).
		Update("role_id", newRoleID)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	result := r.getDB(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"email_verified_at": &now,
			"status":            domain.StatusActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": time.Now(),
			"deleted_by": deletedBy,
			"status":     domain.StatusDeleted,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result := r.getDB(ctx).
		Model(&model.User{}).
		Unscoped().
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"deleted_at": nil,
			"deleted_by": nil,
			"status":     domain.StatusActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) ListUsers(ctx context.Context, filter repository.UserListFilter) ([]model.User, int64, error) {
	var user model.User

	db := r.getDB(ctx).Model(&user)

	// build dynamic query
	if filter.IncludeDeleted {
		db = db.Unscoped()
	}

	if username := strings.TrimSpace(filter.Username); username != "" {
		db = db.Where("username ILIKE ?", "%"+username+"%")
	}

	if email := strings.TrimSpace(filter.Email); email != "" {
		db = db.Where("email ILIKE ?", "%"+email+"%")
	}

	if filter.FullName != "" {
		db = db.Where(
			"CONCAT(first_name, ' ', last_name) ILIKE ?",
			"%"+filter.FullName+"%",
		)
	}

	if filter.Dob != nil {
		db = db.Where("dob = ?", *filter.Dob)
	}

	if filter.RoleID != nil {
		db = db.Where("role_id = ?", *filter.RoleID)
	}

	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	var users []model.User

	if err := db.Preload("Role").Offset(offset).Limit(filter.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *GormUserRepository) FindUsersByRoleCodes(ctx context.Context, roleCodes []string, page, pageSize int) ([]model.User, error) {
	var users []model.User

	db := r.getDB(ctx).Model(&model.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.code IN ?", roleCodes).
		Where("users.status = ?", domain.StatusActive)

	offset := (page - 1) * pageSize
	err := db.Preload("Role").Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *GormUserRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}
