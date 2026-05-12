package models

import (
	"portal-system/internal/domain"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Code        domain.RoleCode `gorm:"size:50;uniqueIndex;not null"`
	Name        string          `gorm:"size:100;not null"`
	IsSystem    bool            `gorm:"not null;default:false"`
	Permissions []Permission    `gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE;"`
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
}
