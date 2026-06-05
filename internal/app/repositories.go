package app

import (
	"portal-system/internal/repository"
	impl "portal-system/internal/repository/impl"

	"gorm.io/gorm"
)

type Repositories struct {
	UserRepo             *impl.GormUserRepository
	AuditLog             *impl.GormAuditLogRepository
	TokenRepo            *impl.GormUserTokenRepository
	RoleRepo             *impl.GormRoleRepository
	PermissionRepo       *impl.GormPermissionRepository
	SessionRepo          *impl.GormAuthSessionRepository
	RefreshRepo          *impl.GormRefreshTokenRepository
	OutboxRepo           *impl.GormOutboxRepository
	AnnouncementRepo     repository.AnnouncementRepository
	UserNotificationRepo repository.UserNotificationRepository
	TxManager            *impl.GormTxManager
}

func newRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UserRepo:             impl.NewGormUserRepository(db),
		AuditLog:             impl.NewGormAuditLogRepository(db),
		TokenRepo:            impl.NewGormUserTokenRepository(db),
		RoleRepo:             impl.NewGormRoleRepository(db),
		PermissionRepo:       impl.NewGormPermissionRepository(db),
		SessionRepo:          impl.NewGormAuthSessionRepository(db),
		RefreshRepo:          impl.NewGormRefreshTokenRepository(db),
		OutboxRepo:           impl.NewGormOutboxRepository(db),
		AnnouncementRepo:     impl.NewGormAnnouncementRepository(db),
		UserNotificationRepo: impl.NewGormUserNotificationRepository(db),
		TxManager:            impl.NewGormTxManager(db),
	}
}
