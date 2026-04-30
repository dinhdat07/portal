package bootstrap

import (
	"errors"
	"portal-system/internal/config"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedAdmin(db *gorm.DB, cfg *config.Config) error {
	var existing models.User
	err := db.Where("email = ?", cfg.AdminEmail).First(&existing).Error
	if err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashStr := string(hash)
	now := time.Now()

	var role models.Role
	if err := db.Where("code = ?", constants.RoleCodeAdmin).First(&role).Error; err != nil {
		return err
	}

	admin := &models.User{
		Email:           cfg.AdminEmail,
		Username:        "admin",
		FirstName:       "System",
		LastName:        "Admin",
		PasswordHash:    &hashStr,
		RoleID:          role.ID,
		Status:          enum.StatusActive,
		EmailVerifiedAt: &now,
	}

	return db.Create(admin).Error
}

func SeedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{
			ID:       uuid.New(),
			Code:     constants.RoleCodeAdmin,
			Name:     "Admin",
			IsSystem: true,
		},
		{
			ID:       uuid.New(),
			Code:     constants.RoleCodeUser,
			Name:     "User",
			IsSystem: true,
		},
	}

	for _, r := range roles {
		var existing models.Role
		err := db.Where("code = ?", r.Code).First(&existing).Error
		switch {
		case err == nil:
			if err := db.Model(&existing).Updates(map[string]any{
				"name":      r.Name,
				"is_system": r.IsSystem,
			}).Error; err != nil {
				return err
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		default:
			return err
		}
	}

	return nil
}

func SeedPermissions(db *gorm.DB) error {
	for _, perm := range constants.AllPermissions {
		code := string(perm)

		err := db.
			Where(models.Permission{Code: code}).
			Assign(models.Permission{Name: code}).
			FirstOrCreate(&models.Permission{
				ID:   uuid.New(),
				Code: code,
				Name: code,
			}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func SeedRolePermissions(db *gorm.DB) error {
	for roleCode, perms := range constants.RolePermissions {

		var role models.Role
		if err := db.Where("code = ?", string(roleCode)).First(&role).Error; err != nil {
			return err
		}

		for _, permCode := range perms {
			var perm models.Permission
			if err := db.Where("code = ?", string(permCode)).First(&perm).Error; err != nil {
				return err
			}

			rp := models.RolePermission{
				RoleID:       role.ID,
				PermissionID: perm.ID,
			}

			// avoid duplicate
			err := db.
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(&rp).Error
			if err != nil {
				return err
			}
		}
	}

	return nil
}
