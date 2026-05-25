package app

import (
	repository "portal-system/internal/repository/impl"

	"gorm.io/gorm"
)

type Repositories struct {
	UserRepo       *repository.GormUserRepository
	AuditLog       *repository.GormAuditLogRepository
	TokenRepo      *repository.GormUserTokenRepository
	RoleRepo       *repository.GormRoleRepository
	PermissionRepo *repository.GormPermissionRepository
	SessionRepo    *repository.GormAuthSessionRepository
	RefreshRepo    *repository.GormRefreshTokenRepository
	OutboxRepo     *repository.GormOutboxRepository
	TxManager      *repository.GormTxManager
}

func newRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:       repository.NewGormUserRepository(db),
		AuditLog:       repository.NewGormAuditLogRepository(db),
		TokenRepo:      repository.NewGormUserTokenRepository(db),
		RoleRepo:       repository.NewGormRoleRepository(db),
		PermissionRepo: repository.NewGormPermissionRepository(db),
		SessionRepo:    repository.NewGormAuthSessionRepository(db),
		RefreshRepo:    repository.NewGormRefreshTokenRepository(db),
		OutboxRepo:     repository.NewGormOutboxRepository(db),
		TxManager:      repository.NewGormTxManager(db),
	}
}
