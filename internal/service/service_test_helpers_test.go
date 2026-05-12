package service_test

import (
	"context"
	"time"

	repositorymock "portal-system/internal/repository/mock"
	. "portal-system/internal/service"
	servicemock "portal-system/internal/service/mock"

	"github.com/stretchr/testify/mock"
)

type authServiceTestDeps struct {
	tx          *repositorymock.TxManager
	auditLogger *servicemock.AuditLogger
	userRepo    *repositorymock.UserRepository
	tokenRepo   *repositorymock.UserTokenRepository
	roleRepo    *repositorymock.RoleRepository
	refreshRepo *repositorymock.RefreshTokenRepository
	sessionRepo *repositorymock.AuthSessionRepository
	revoStore   *servicemock.SessionRevocationStore
	tokenMgr    *servicemock.TokenIssuer
	email       *servicemock.EmailSender
}

func newAdminServiceForTest(
	tx *repositorymock.TxManager,
	auditLogger *servicemock.AuditLogger,
	userRepo *repositorymock.UserRepository,
	tokenRepo *repositorymock.UserTokenRepository,
	roleRepo *repositorymock.RoleRepository,
	tokenMgr *servicemock.TokenIssuer,
	email *servicemock.EmailSender,
) AdminService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditLogger == nil {
		auditLogger = newAuditLoggerMock()
	}
	if tokenRepo == nil {
		tokenRepo = &repositorymock.UserTokenRepository{}
	}
	if roleRepo == nil {
		roleRepo = &repositorymock.RoleRepository{}
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
	repo *repositorymock.PermissionRepository,
) PermissionService {
	return NewPermissionService(repo)
}

func newRoleServiceForTest(
	tx *repositorymock.TxManager,
	auditLogger *servicemock.AuditLogger,
	roleRepo *repositorymock.RoleRepository,
	permRepo *repositorymock.PermissionRepository,
	userRepo *repositorymock.UserRepository,
) RoleService {
	if tx == nil {
		tx = newPassthroughTxManager()
	}
	if auditLogger == nil {
		auditLogger = newAuditLoggerMock()
	}
	if roleRepo == nil {
		roleRepo = &repositorymock.RoleRepository{}
	}
	if permRepo == nil {
		permRepo = &repositorymock.PermissionRepository{}
	}
	if userRepo == nil {
		userRepo = &repositorymock.UserRepository{}
	}
	return NewRoleService(RoleServiceDeps{
		RoleRepo:       roleRepo,
		PermissionRepo: permRepo,
		UserRepo:       userRepo,
		TxManager:      tx,
		AuditLogger:    auditLogger,
	})
}

func newPassthroughTxManager() *repositorymock.TxManager {
	tx := &repositorymock.TxManager{}
	tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}).Maybe()
	return tx
}

func newAuditLoggerMock() *servicemock.AuditLogger {
	logger := &servicemock.AuditLogger{}
	logger.EXPECT().Log(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	logger.EXPECT().LogWithMetadata(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return logger
}

func newSessionRevocationStore() *servicemock.SessionRevocationStore {
	store := &servicemock.SessionRevocationStore{}
	store.EXPECT().IsRevoked(mock.Anything, mock.Anything).Return(false, nil).Maybe()
	store.EXPECT().MarkRevoked(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return store
}

func newTokenIssuerMock() *servicemock.TokenIssuer {
	tokenMgr := &servicemock.TokenIssuer{}
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

func newEmailSenderMock() *servicemock.EmailSender {
	email := &servicemock.EmailSender{}
	email.EXPECT().SendVerificationEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendResetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	email.EXPECT().SendSetPasswordEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	return email
}
