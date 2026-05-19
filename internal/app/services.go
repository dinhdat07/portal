package app

import (
	"time"

	"portal-system/config"
	"portal-system/internal/infrastructure/security"
	"portal-system/internal/service"
)

type Services struct {
	// services
	AuditLog   service.AuditLogger
	Auth       service.AuthService
	User       service.UserService
	Admin      service.AdminService
	Role       service.RoleService
	Permission service.PermissionService

	// auth layer
	Authenticator *security.Authenticator
	Authorizer    *security.Authorizer
}

func newServices(
	cfg *config.Config,
	infra *Infra,
	repos *Repositories,
) *Services {

	// audit
	auditLogService := service.NewAuditLogService(repos.AuditLog)

	// auth service
	authService := service.NewAuthService(service.AuthServiceDeps{
		TxManager:             repos.TxManager,
		AuditLogger:           auditLogService,
		UserRepo:              repos.UserRepo,
		TokenRepo:             repos.TokenRepo,
		RoleRepo:              repos.RoleRepo,
		RefreshTokenRepo:      repos.RefreshRepo,
		SessionRepo:           repos.SessionRepo,
		RevoStore:             infra.RevocationStore,
		TokenManager:          infra.TokenManager,
		NotificationPublisher: infra.NotificationPublisher,
		FrontendBaseURL:       cfg.FrontEndUrl,
		RefreshTTL:            time.Duration(cfg.RefreshTTL) * time.Second,
	})

	// user service
	userService := service.NewUserService(service.UserServiceDeps{
		TxManager:   repos.TxManager,
		AuditLogger: auditLogService,
		UserRepo:    repos.UserRepo,
		RoleRepo:    repos.RoleRepo,
	})

	// admin service
	adminService := service.NewAdminService(service.AdminServiceDeps{
		TxManager:             repos.TxManager,
		AuditLogger:           auditLogService,
		UserRepo:              repos.UserRepo,
		TokenManager:          infra.TokenManager,
		TokenRepo:             repos.TokenRepo,
		RoleRepo:              repos.RoleRepo,
		NotificationPublisher: infra.NotificationPublisher,
		FrontendURL:           cfg.FrontEndUrl,
	})

	roleService := service.NewRoleService(service.RoleServiceDeps{
		RoleRepo:       repos.RoleRepo,
		PermissionRepo: repos.PermissionRepo,
		UserRepo:       repos.UserRepo,
		TxManager:      repos.TxManager,
		AuditLogger:    auditLogService,
	})

	permissionService := service.NewPermissionService(repos.PermissionRepo)

	// auth layer
	authenticator := security.NewAuthenticator(
		infra.TokenManager,
		repos.RoleRepo,
		repos.SessionRepo,
		infra.RevocationStore,
	)

	authorizer := security.NewAuthorizer()

	return &Services{
		AuditLog:   auditLogService,
		Auth:       authService,
		User:       userService,
		Admin:      adminService,
		Role:       roleService,
		Permission: permissionService,

		Authenticator: authenticator,
		Authorizer:    authorizer,
	}
}
