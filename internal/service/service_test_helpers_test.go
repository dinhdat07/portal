package service_test

import (
	"context"
	"time"

	repositorymocks "portal-system/internal/repository/mocks"
	. "portal-system/internal/service"
	servicemocks "portal-system/internal/service/mocks"

	"github.com/stretchr/testify/mock"
)

type authServiceTestDeps struct {
	tx          *repositorymocks.TxManager
	auditLogger *servicemocks.AuditLogger
	userRepo    *repositorymocks.UserRepository
	tokenRepo   *repositorymocks.UserTokenRepository
	roleRepo    *repositorymocks.RoleRepository
	refreshRepo *repositorymocks.RefreshTokenRepository
	sessionRepo *repositorymocks.AuthSessionRepository
	revoStore   *servicemocks.SessionRevocationStore
	tokenMgr    *servicemocks.TokenIssuer
	email       *servicemocks.EmailSender
}

func newAdminServiceForTest(
	tx *repositorymocks.TxManager,
	auditLogger *servicemocks.AuditLogger,
	userRepo *repositorymocks.UserRepository,
	tokenRepo *repositorymocks.UserTokenRepository,
	roleRepo *repositorymocks.RoleRepository,
	tokenMgr *servicemocks.TokenIssuer,
	email *servicemocks.EmailSender,
) AdminService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditLogger == nil {
		auditLogger = newAuditLoggerMock()
	}
	if tokenRepo == nil {
		tokenRepo = &repositorymocks.UserTokenRepository{}
	}
	if roleRepo == nil {
		roleRepo = &repositorymocks.RoleRepository{}
	}
	if tokenMgr == nil {
		tokenMgr = newTokenIssuerMock()
	}
	if email == nil {
		email = newEmailSenderMock()
	}
	return NewAdminService(AdminServiceDeps{
		TxManager:    tx,
		AuditLogger:  auditLogger,
		UserRepo:     userRepo,
		TokenManager: tokenMgr,
		TokenRepo:    tokenRepo,
		RoleRepo:     roleRepo,
		EmailSvc:     email,
		FrontendURL:  "http://frontend.local",
	})
}

func newAuthServiceForTest(deps authServiceTestDeps) AuthService {
	if deps.tx == nil {
		deps.tx = newPassthroughTxManager()
	}
	if deps.auditLogger == nil {
		deps.auditLogger = newAuditLoggerMock()
	}
	if deps.revoStore == nil {
		deps.revoStore = newSessionRevocationStore()
	}
	if deps.tokenMgr == nil {
		deps.tokenMgr = newTokenIssuerMock()
	}
	if deps.email == nil {
		deps.email = newEmailSenderMock()
	}
	return NewAuthService(AuthServiceDeps{
		TxManager:        deps.tx,
		AuditLogger:      deps.auditLogger,
		UserRepo:         deps.userRepo,
		RefreshTokenRepo: deps.refreshRepo,
		TokenRepo:        deps.tokenRepo,
		RoleRepo:         deps.roleRepo,
		SessionRepo:      deps.sessionRepo,
		RevoStore:        deps.revoStore,
		TokenManager:     deps.tokenMgr,
		EmailService:     deps.email,
		FrontendBaseURL:  "http://frontend.local",
		RefreshTTL:       24 * time.Hour,
	})
}

func newPermissionServiceForTest(
	repo *repositorymocks.PermissionRepository,
) PermissionService {
	return NewPermissionService(repo)
}

func newRoleServiceForTest(
	tx *repositorymocks.TxManager,
	auditLogger *servicemocks.AuditLogger,
	roleRepo *repositorymocks.RoleRepository,
	permRepo *repositorymocks.PermissionRepository,
	userRepo *repositorymocks.UserRepository,
) RoleService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditLogger == nil {
		auditLogger = newAuditLoggerMock()
	}
	if roleRepo == nil {
		roleRepo = &repositorymocks.RoleRepository{}
	}
	if permRepo == nil {
		permRepo = &repositorymocks.PermissionRepository{}
	}
	if userRepo == nil {
		userRepo = &repositorymocks.UserRepository{}
	}
	return NewRoleService(RoleServiceDeps{
		RoleRepo:       roleRepo,
		PermissionRepo: permRepo,
		UserRepo:       userRepo,
		TxManager:      tx,
		AuditLogger:    auditLogger,
	})
}

func newPassthroughTxManager() *repositorymocks.TxManager {
	tx := &repositorymocks.TxManager{}
	tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}).Maybe()
	return tx
}

func newAuditLoggerMock() *servicemocks.AuditLogger {
	logger := &servicemocks.AuditLogger{}
	logger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	logger.EXPECT().LogWithMetadata(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return logger
}

func newSessionRevocationStore() *servicemocks.SessionRevocationStore {
	store := &servicemocks.SessionRevocationStore{}
	store.EXPECT().IsRevoked(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	store.EXPECT().MarkRevoked(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}

func newTokenIssuerMock() *servicemocks.TokenIssuer {
	tokenMgr := &servicemocks.TokenIssuer{}
	tokenMgr.EXPECT().GenerateAccessToken(mock.Anything).Return("access-token", nil).Maybe()
	tokenMgr.EXPECT().GenerateRefreshToken().Return("refresh-token", nil).Maybe()
	tokenMgr.EXPECT().ExpiresInSeconds().Return(3600).Maybe()
	tokenMgr.EXPECT().Parse(mock.Anything).Return(nil, nil).Maybe()
	tokenMgr.EXPECT().HashToken(mock.Anything).RunAndReturn(func(raw string) string {
		return "hashed-" + raw
	}).Maybe()
	tokenMgr.EXPECT().GenerateHashToken().Return("token-hash", "raw-token", nil).Maybe()
	return tokenMgr
}

func newEmailSenderMock() *servicemocks.EmailSender {
	email := &servicemocks.EmailSender{}
	email.EXPECT().SendVerificationEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendResetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendSetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return email
}
