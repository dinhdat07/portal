package services

import (
	"context"
	"errors"
	appLogger "log"

	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*models.User, error)
	UpdateProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, input UpdateUserInput) (*models.User, error)
	ChangePassword(ctx context.Context, meta *AuditMeta, actor *AuditUser, current, newPassword, confirm string) error
}

type userService struct {
	txManager   repositories.TxManager
	auditLogger AuditLogger
	roleRepo    repositories.RoleRepository
	userRepo    repositories.UserRepository
}

type UserServiceDeps struct {
	TxManager   repositories.TxManager
	AuditLogger AuditLogger
	RoleRepo    repositories.RoleRepository
	UserRepo    repositories.UserRepository
}

func NewUserService(deps UserServiceDeps) *userService {
	return &userService{
		txManager:   deps.TxManager,
		userRepo:    deps.UserRepo,
		roleRepo:    deps.RoleRepo,
		auditLogger: deps.AuditLogger,
	}
}

func (svc *userService) GetProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if actor.RoleCode == domain.RoleCodeAdmin {
		target := MapUserToAuditUser(user)
		if err := svc.auditLogger.Log(ctx, meta, domain.ActionAdminViewUser, actor, target); err != nil {
			appLogger.Println("failed to log admin view user action", "error", err)
		}
	}

	return user, nil
}

func (svc *userService) ChangePassword(ctx context.Context, meta *AuditMeta, actor *AuditUser, current, newPassword, confirm string) error {
	user, err := svc.userRepo.FindByID(ctx, actor.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrUnauthorized
		}
		return err
	}

	if strings.TrimSpace(newPassword) == "" ||
		strings.TrimSpace(confirm) == "" {
		return ErrInvalidInput
	}

	// check nil before compare to avoid panic
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return ErrUnauthorized
	}

	if newPassword != confirm {
		return ErrPasswordConfirmationMismatch
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(current)); err != nil {
		return ErrIncorrectPassword
	}

	if current == newPassword {
		return ErrNewPasswordMustBeDifferent
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.UpdatePassword(ctx, actor.ID, string(hashed)); err != nil {
			return ErrInternalServer
		}

		if err := svc.auditLogger.Log(ctx, meta, domain.ActionChangePassword, actor, actor); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	return err

}

func (svc *userService) UpdateProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, input UpdateUserInput) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
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

	changes := map[string]any{}

	// update allowed fields
	if input.FirstName != nil {
		changes["first_name"] = map[string]any{
			"old": user.FirstName,
			"new": *input.FirstName,
		}
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		changes["last_name"] = map[string]any{
			"old": user.LastName,
			"new": *input.LastName,
		}
		user.LastName = *input.LastName
	}
	if input.DOB != nil {
		changes["dob"] = map[string]any{
			"old": user.DOB,
			"new": input.DOB,
		}
		user.DOB = input.DOB
	}

	// check duplicate username
	if input.Username != nil {
		username := normalizeUsername(*input.Username)
		if err := validateNormalizedUsername(username); err != nil {
			return nil, err
		}
		if username != normalizeUsername(user.Username) {
			existing, err := svc.userRepo.FindByUsername(ctx, username)
			if err != nil {
				return nil, err
			}
			if existing != nil && existing.ID != user.ID {
				return nil, ErrUsernameExists
			}
		}
		if username != user.Username {
			changes["username"] = map[string]any{
				"old": user.Username,
				"new": username,
			}
			user.Username = username
		}
	}

	err = svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := svc.userRepo.Update(ctx, user); err != nil {
			return ErrInternalServer
		}

		action := domain.ActionUpdateProfile
		if actor.RoleCode == domain.RoleCodeAdmin {
			action = domain.ActionAdminUpdateUser
		}

		target := MapUserToAuditUser(user)
		err := svc.auditLogger.LogWithMetadata(ctx, meta, action, actor, target, map[string]any{
			"changes": changes,
		})
		if err != nil {
			return ErrAuditLogger
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
