package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"
	notificationv1 "portal-system/shared/events/notification/v1"

	appLogger "log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService interface {
	ListUsers(ctx context.Context, meta *AuditMeta, actor *AuditUser, in UsersFilter) (*ListUsersResult, error)
	CreateUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, in CreateUserInput) (*model.User, error)
	DeleteUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, userID uuid.UUID) (*model.User, error)
	RestoreUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, userID uuid.UUID) (*model.User, error)
	UpdateRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, roleCode domain.RoleCode) (*model.User, error)
}

type adminService struct {
	txManager         repository.TxManager
	auditLogger       AuditLogger
	userRepo          repository.UserRepository
	tokenRepo         repository.UserTokenRepository
	tokenManager      TokenIssuer
	roleRepo          repository.RoleRepository
	outboxRepo        repository.OutboxRepository
	notificationTopic string
	frontendURL       string
}

type AdminServiceDeps struct {
	TxManager         repository.TxManager
	AuditLogger       AuditLogger
	UserRepo          repository.UserRepository
	TokenManager      TokenIssuer
	TokenRepo         repository.UserTokenRepository
	RoleRepo          repository.RoleRepository
	OutboxRepo        repository.OutboxRepository
	NotificationTopic string
	FrontendURL       string
}

func NewAdminService(deps AdminServiceDeps) *adminService {
	return &adminService{
		txManager:         deps.TxManager,
		userRepo:          deps.UserRepo,
		tokenRepo:         deps.TokenRepo,
		roleRepo:          deps.RoleRepo,
		tokenManager:      deps.TokenManager,
		auditLogger:       deps.AuditLogger,
		outboxRepo:        deps.OutboxRepo,
		notificationTopic: deps.NotificationTopic,
		frontendURL:       deps.FrontendURL,
	}
}

func (svc *adminService) ListUsers(ctx context.Context, meta *AuditMeta, actor *AuditUser, in UsersFilter) (*ListUsersResult, error) {
	repoFilter := repository.UserListFilter{
		Page:           in.Page,
		PageSize:       in.PageSize,
		Username:       in.Username,
		Email:          in.Email,
		FullName:       in.FullName,
		Dob:            in.Dob,
		Status:         in.Status,
		IncludeDeleted: in.IncludeDeleted,
	}

	if in.RoleCode != nil {
		role, err := svc.roleRepo.FindByCode(ctx, *in.RoleCode)
		if err != nil {
			return nil, ErrInvalidInput
		}
		repoFilter.RoleID = &role.ID
	}

	if in.Status != "" && !in.Status.IsValid() {
		return nil, ErrInvalidInput
	}

	users, total, err := svc.userRepo.ListUsers(ctx, repoFilter)
	if err != nil {
		return nil, ErrInternalServer
	}

	logMeta := map[string]any{
		"filters": map[string]any{
			"username":        in.Username,
			"email":           in.Email,
			"full_name":       in.FullName,
			"dob":             in.Dob,
			"role_code":       in.RoleCode,
			"role_id":         repoFilter.RoleID,
			"status":          in.Status,
			"include_deleted": in.IncludeDeleted,
		},
		"pagination": map[string]any{
			"page":      in.Page,
			"page_size": in.PageSize,
		},
		"result_count": len(users),
		"total":        total,
	}

	if err := svc.auditLogger.LogWithMetadata(
		ctx,
		meta,
		domain.ActionAdminSearchUser,
		actor,
		nil,
		logMeta,
	); err != nil {
		appLogger.Println("failed to log admin search user action", "error", err)
	}

	return &ListUsersResult{
		Users:    users,
		Total:    total,
		Page:     in.Page,
		PageSize: in.PageSize,
	}, nil

}

func (svc *adminService) CreateUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, in CreateUserInput) (*model.User, error) {
	if in.RoleCode == "" {
		return nil, ErrInvalidInput
	}

	in.Email = normalizeEmail(in.Email)
	in.Username = normalizeUsername(in.Username)
	if err := validateNormalizedEmail(in.Email); err != nil {
		return nil, err
	}
	if err := validateNormalizedUsername(in.Username); err != nil {
		return nil, err
	}

	existingByEmail, err := svc.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		return nil, ErrInternalServer
	}

	if existingByEmail != nil && existingByEmail.ID != uuid.Nil {
		return nil, ErrEmailExists
	}

	existingByUsername, err := svc.userRepo.FindByUsername(ctx, in.Username)
	if err != nil {
		return nil, ErrInternalServer
	}
	if existingByUsername != nil && existingByUsername.ID != uuid.Nil {
		return nil, ErrUsernameExists
	}

	tokenHash, rawToken, err := svc.tokenManager.GenerateHashToken()
	if err != nil {
		return nil, err
	}

	role, err := svc.roleRepo.FindByCode(ctx, in.RoleCode)
	if role == nil || err != nil {
		return nil, ErrInternalServer
	}

	user := &model.User{
		Email:     in.Email,
		Username:  in.Username,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		DOB:       in.DOB,
		RoleID:    role.ID,
		Role:      *role,
		Status:    domain.StatusPending,
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.Create(txCtx, user); err != nil {
			return ErrInternalServer
		}

		if err := svc.tokenRepo.
			RevokeByUserAndType(txCtx, user.ID, domain.TokenTypePasswordSet); err != nil {
			return ErrInternalServer
		}

		setPasswordToken := &model.UserToken{
			UserID:    user.ID,
			TokenType: domain.TokenTypePasswordSet,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := svc.tokenRepo.Create(txCtx, setPasswordToken); err != nil {
			return ErrInternalServer
		}

		setPasswordURL := fmt.Sprintf("%s/set-password?token=%s", svc.frontendURL, url.QueryEscape(rawToken))
		event := newEmailNotificationEvent(notificationv1.NotificationTypeSetPassword, notificationv1.TemplateSetPassword, *user, setPasswordURL)

		if err := createNotificationOutboxEvent(txCtx, svc.outboxRepo, svc.notificationTopic, event); err != nil {
			return ErrSendSetPasswordEmail
		}

		target := MapUserToAuditUser(user)
		if err := svc.auditLogger.Log(txCtx, meta, domain.ActionAdminCreateUser, actor, target); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *adminService) DeleteUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, userID uuid.UUID) (*model.User, error) {
	user, err := svc.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	roleAdmin, err := svc.roleRepo.FindByCode(ctx, domain.RoleCodeAdmin)
	if err != nil {
		return nil, ErrInternalServer
	}

	if actor.ID != user.ID && user.RoleID == roleAdmin.ID {
		return nil, ErrForbidden
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.Delete(ctx, userID, actor.ID); err != nil {
			return ErrInternalServer
		}

		now := time.Now()
		user.DeletedAt = gorm.DeletedAt{Time: now, Valid: true}
		user.DeletedBy = &actor.ID
		user.Status = domain.StatusDeleted

		target := MapUserToAuditUser(user)
		if err := svc.auditLogger.Log(ctx, meta, domain.ActionAdminDeleteUser, actor, target); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *adminService) RestoreUser(ctx context.Context, meta *AuditMeta, actor *AuditUser, userID uuid.UUID) (*model.User, error) {
	user, err := svc.userRepo.FindByIDUnscoped(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.DeletedAt.Time.IsZero() {
		return nil, ErrUserNotDeleted
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.Restore(ctx, userID); err != nil {
			return ErrInternalServer
		}

		user.DeletedAt = gorm.DeletedAt{}
		user.DeletedBy = nil
		user.Status = domain.StatusActive

		target := MapUserToAuditUser(user)
		if err := svc.auditLogger.Log(ctx, meta, domain.ActionAdminRestoreUser, actor, target); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (svc *adminService) UpdateRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, roleCode domain.RoleCode) (*model.User, error) {
	if roleCode == "" {
		return nil, ErrInvalidInput
	}

	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	role, err := svc.roleRepo.FindByCode(ctx, roleCode)
	if err != nil {
		return nil, ErrInternalServer
	}

	roleAdmin, err := svc.roleRepo.FindByCode(ctx, domain.RoleCodeAdmin)
	if err != nil {
		return nil, ErrInternalServer
	}

	if actor.ID != user.ID && user.RoleID == roleAdmin.ID {
		return nil, ErrForbidden
	}

	if user.RoleID == role.ID {
		return user, nil
	}

	changes := map[string]any{}
	changes["role"] = map[string]any{
		"old": user.Role,
		"new": role,
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.UpdateRole(ctx, id, role.ID); err != nil {
			return ErrInternalServer
		}
		user.Role = *role

		target := MapUserToAuditUser(user)

		if err := svc.auditLogger.LogWithMetadata(ctx, meta, domain.ActionAdminAssignRole, actor, target, changes); err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
