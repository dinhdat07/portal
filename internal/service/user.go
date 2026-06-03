package service

import (
	"context"
	"errors"
	appLogger "log"

	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"
	"strings"
	"time"
	"fmt"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*model.User, error)
	UpdateProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, input UpdateUserInput) (*model.User, error)
	ChangePassword(ctx context.Context, meta *AuditMeta, actor *AuditUser, current, newPassword, confirm string) error
	RegisterFCMToken(ctx context.Context, actor *AuditUser, token string, deviceName *string) error
	GenerateTelegramLinkToken(ctx context.Context, actor *AuditUser) (token string, link string, expiresInSeconds int, err error)
	ProcessTelegramWebhook(ctx context.Context, chatID int64, text string) error
}

type userService struct {
	txManager           repository.TxManager
	auditLogger         AuditLogger
	roleRepo            repository.RoleRepository
	userRepo            repository.UserRepository
	outboxRepo          repository.OutboxRepository
	redisClient         redis.UniversalClient
	frontendURL         string
	telegramBotUsername string
}

type UserServiceDeps struct {
	TxManager           repository.TxManager
	AuditLogger         AuditLogger
	RoleRepo            repository.RoleRepository
	UserRepo            repository.UserRepository
	OutboxRepo          repository.OutboxRepository
	RedisClient         redis.UniversalClient
	FrontendURL         string
	TelegramBotUsername string
}

func NewUserService(deps UserServiceDeps) *userService {
	return &userService{
		txManager:   deps.TxManager,
		userRepo:    deps.UserRepo,
		roleRepo:    deps.RoleRepo,
		auditLogger: deps.AuditLogger,
		outboxRepo:          deps.OutboxRepo,
		redisClient:         deps.RedisClient,
		frontendURL:         deps.FrontendURL,
		telegramBotUsername: deps.TelegramBotUsername,
	}
}

func (svc *userService) GetProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*model.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
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
		if errors.Is(err, repository.ErrNotFound) {
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

func (svc *userService) UpdateProfile(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID, input UpdateUserInput) (*model.User, error) {
	user, err := svc.userRepo.FindByID(ctx, id)
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

type NotificationEndpointRegisteredPayload struct {
	UserID     string  `json:"user_id"`
	Provider   string  `json:"provider"`
	Endpoint   string  `json:"endpoint"`
	DeviceName *string `json:"device_name,omitempty"`
}

func (svc *userService) RegisterFCMToken(ctx context.Context, actor *AuditUser, token string, deviceName *string) error {
	payload := NotificationEndpointRegisteredPayload{
		UserID:     actor.ID.String(),
		Provider:   "FIREBASE",
		Endpoint:   token,
		DeviceName: deviceName,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ErrInternalServer
	}

	outboxEvent := &model.OutboxEvent{
		ID:         uuid.New(),
		Topic:      "notification.endpoint.registered",
		MessageKey: actor.ID.String(),
		Payload:    payloadBytes,
		Status:     model.OutboxStatusPending,
		RetryCount: 0,
		MaxRetry:   10,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		return svc.outboxRepo.Create(txCtx, outboxEvent)
	})
}

func (svc *userService) GenerateTelegramLinkToken(ctx context.Context, actor *AuditUser) (string, string, int, error) {
	token := "link_" + uuid.NewString()[:8]
	expiresIn := 300 // 5 mins

	key := fmt.Sprintf("telegram_link:%s", token)
	if err := svc.redisClient.Set(ctx, key, actor.ID.String(), time.Duration(expiresIn)*time.Second).Err(); err != nil {
		return "", "", 0, ErrInternalServer
	}

	// Generate telegram deep link using bot username from config
	link := fmt.Sprintf("https://t.me/%s?start=%s", svc.telegramBotUsername, token)

	return token, link, expiresIn, nil
}

func (svc *userService) ProcessTelegramWebhook(ctx context.Context, chatID int64, text string) error {
	if !strings.HasPrefix(text, "/start ") {
		return nil // Not a link token command, ignore
	}

	token := strings.TrimPrefix(text, "/start ")
	key := fmt.Sprintf("telegram_link:%s", token)

	userID, err := svc.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return errors.New("invalid or expired link token")
		}
		return err
	}

	// Token is valid, let's delete it so it can't be reused
	svc.redisClient.Del(ctx, key)

	payload := NotificationEndpointRegisteredPayload{
		UserID:   userID,
		Provider: "TELEGRAM",
		Endpoint: fmt.Sprintf("%d", chatID),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ErrInternalServer
	}

	outboxEvent := &model.OutboxEvent{
		ID:         uuid.New(),
		Topic:      "notification.endpoint.registered",
		MessageKey: userID,
		Payload:    payloadBytes,
		Status:     model.OutboxStatusPending,
		RetryCount: 0,
		MaxRetry:   10,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	return svc.txManager.WithTx(ctx, func(txCtx context.Context) error {
		return svc.outboxRepo.Create(txCtx, outboxEvent)
	})
}

