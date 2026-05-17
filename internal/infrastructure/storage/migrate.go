package storage

import (
	"portal-system/internal/model"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.AuditLog{},
		&model.UserToken{},
		&model.Role{},
		&model.Permission{},
		&model.AuthSession{},
		&model.RefreshToken{},
	)
}
