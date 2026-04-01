package services

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db          *gorm.DB
	auditLogger *AuditLogService
	userRepo    *repositories.UserRepository
}

func NewUserService(db *gorm.DB, repo *repositories.UserRepository, logger *AuditLogService) *UserService {
	return &UserService{db: db, userRepo: repo, auditLogger: logger}
}

func (svc *UserService) GetProfile(ctx context.Context, meta *domain.AuditMeta, actor *models.User, id uuid.UUID) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if actor.Role == enum.RoleAdmin {
		svc.auditLogger.Log(ctx, meta, enum.ActionAdminViewUser, actor, user)
	}

	return user, nil
}

func (svc *UserService) ChangePassword(ctx context.Context, meta *domain.AuditMeta, id uuid.UUID, current, newPassword, confirm string) error {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnauthorized
		}
		return err
	}

	//check nil before compare to avoid panic
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

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.UpdatePassword(ctx, id, string(hashed)); err != nil {
			return ErrInternalServer
		}
		if err := svc.auditLogger.Log(ctx, meta, enum.ActionChangePassword, user, user); err != nil {
			return ErrAuditLogger
		}
		return nil
	})

	return err

}

func (svc *UserService) UpdateProfile(ctx context.Context, meta *domain.AuditMeta, actor *models.User, id uuid.UUID, input domain.UpdateUserInput) (*models.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
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
	if input.Username != nil && *input.Username != user.Username {
		existing, err := svc.userRepo.FindByUsername(ctx, *input.Username)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil {
			return nil, ErrUsernameExists
		}
		changes["username"] = map[string]any{
			"old": user.Username,
			"new": *input.Username,
		}
		user.Username = *input.Username
	}

	err = svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := svc.userRepo.Update(ctx, user); err != nil {
			return ErrInternalServer
		}

		action := enum.ActionUpdateProfile
		if actor.Role == enum.RoleAdmin {
			action = enum.ActionAdminUpdateUser
		}

		err := svc.auditLogger.LogWithMetadata(ctx, meta, action, actor, user, map[string]any{
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
