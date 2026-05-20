package service_test

import (
	"context"
	"encoding/json"
	"portal-system/internal/model"
	"time"

	repositorymock "portal-system/internal/repository/mock"
	. "portal-system/internal/service"
	servicemock "portal-system/internal/service/mock"
	notificationv1 "portal-system/shared/events/notification/v1"

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
	outboxRepo  *repositorymock.OutboxRepository
}

func newAdminServiceForTest(
	tx *repositorymock.TxManager,
	auditLogger *servicemock.AuditLogger,
	userRepo *repositorymock.UserRepository,
	tokenRepo *repositorymock.UserTokenRepository,
	roleRepo *repositorymock.RoleRepository,
	tokenMgr *servicemock.TokenIssuer,
	outboxRepo *repositorymock.OutboxRepository,
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
	if outboxRepo == nil {
		outboxRepo = newOutboxRepositoryMock()
	}
	return NewAdminService(AdminServiceDeps{
		TxManager:         tx,
		AuditLogger:       auditLogger,
		UserRepo:          userRepo,
		TokenManager:      tokenMgr,
		TokenRepo:         tokenRepo,
		RoleRepo:          roleRepo,
		OutboxRepo:        outboxRepo,
		NotificationTopic: "notification.requested",
		FrontendURL:       "http://frontend.local",
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
	if deps.outboxRepo == nil {
		deps.outboxRepo = newOutboxRepositoryMock()
	}
	return NewAuthService(AuthServiceDeps{
		TxManager:         deps.tx,
		AuditLogger:       deps.auditLogger,
		UserRepo:          deps.userRepo,
		RefreshTokenRepo:  deps.refreshRepo,
		TokenRepo:         deps.tokenRepo,
		RoleRepo:          deps.roleRepo,
		SessionRepo:       deps.sessionRepo,
		RevoStore:         deps.revoStore,
		TokenManager:      deps.tokenMgr,
		OutboxRepo:        deps.outboxRepo,
		NotificationTopic: "notification.requested",
		FrontendBaseURL:   "http://frontend.local",
		RefreshTTL:        24 * time.Hour,
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

func newOutboxRepositoryMock() *repositorymock.OutboxRepository {
	outboxRepo := &repositorymock.OutboxRepository{}
	outboxRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil).Maybe()
	return outboxRepo
}

func matchOutboxNotification(match func(notificationv1.NotificationRequestedEvent) bool) interface{} {
	return mock.MatchedBy(func(event *model.OutboxEvent) bool {
		if event == nil {
			return false
		}
		if event.Topic != "notification.requested" || event.MessageKey == "" || event.Status != model.OutboxStatusPending {
			return false
		}

		var payload notificationv1.NotificationRequestedEvent
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return false
		}
		if payload.EventID == "" || payload.EventID != event.MessageKey {
			return false
		}

		return match(payload)
	})
}
