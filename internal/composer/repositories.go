package composer

import (
	repositories "portal-system/internal/repositories/impl"

	"gorm.io/gorm"
)

type Repositories struct {
	UserRepo       *repositories.GormUserRepository
	AuditLog       *repositories.GormAuditLogRepository
	TokenRepo      *repositories.GormUserTokenRepository
	RoleRepo       *repositories.GormRoleRepository
	PermissionRepo *repositories.GormPermissionRepository
	SessionRepo    *repositories.GormAuthSessionRepository
	RefreshRepo    *repositories.GormRefreshTokenRepository
	TxManager      *repositories.GormTxManager
}

func newRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:       repositories.NewGormUserRepository(db),
		AuditLog:       repositories.NewGormAuditLogRepository(db),
		TokenRepo:      repositories.NewGormUserTokenRepository(db),
		RoleRepo:       repositories.NewGormRoleRepository(db),
		PermissionRepo: repositories.NewGormPermissionRepository(db),
		SessionRepo:    repositories.NewGormAuthSessionRepository(db),
		RefreshRepo:    repositories.NewGormRefreshTokenRepository(db),
		TxManager:      repositories.NewGormTxManager(db),
	}
}
