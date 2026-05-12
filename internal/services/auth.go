package services

import (
	"context"
	"errors"
	"fmt"
	appLogger "log"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, meta *AuditMeta, email, username, password, firstName, lastName string, dob time.Time) error
	LogIn(ctx context.Context, meta *AuditMeta, identifier, password string) (*LoginResult, error)
	Logout(ctx context.Context, meta *AuditMeta, actor *AuditUser, sessionID uuid.UUID) error
	LogoutAll(ctx context.Context, meta *AuditMeta, actor *AuditUser) error

	VerifyEmail(ctx context.Context, meta *AuditMeta, rawToken string, tokenType domain.TokenType) error
	ResendVerification(ctx context.Context, meta *AuditMeta, email string, tokenType domain.TokenType) error

	ForgotPassword(ctx context.Context, meta *AuditMeta, email string) error
	ResetPassword(ctx context.Context, meta *AuditMeta, in *SetPasswordInput, tokenType domain.TokenType) error
	SetPassword(ctx context.Context, meta *AuditMeta, in *SetPasswordInput, tokenType domain.TokenType) error

	Refresh(ctx context.Context, meta *AuditMeta, refreshToken string) (*RefreshResult, error)
}

type authService struct {
	txManager       repositories.TxManager
	auditLogger     AuditLogger
	userRepo        repositories.UserRepository
	refreshRepo     repositories.RefreshTokenRepository
	tokenRepo       repositories.UserTokenRepository
	roleRepo        repositories.RoleRepository
	sessionRepo     repositories.AuthSessionRepository
	revoStore       SessionRevocationStore
	tokenManager    TokenIssuer
	emailService    EmailSender
	frontendBaseURL string
	refreshTTL      time.Duration
}

type AuthServiceDeps struct {
	TxManager        repositories.TxManager
	AuditLogger      AuditLogger
	UserRepo         repositories.UserRepository
	RefreshTokenRepo repositories.RefreshTokenRepository
	TokenRepo        repositories.UserTokenRepository
	RoleRepo         repositories.RoleRepository
	SessionRepo      repositories.AuthSessionRepository
	RevoStore        SessionRevocationStore

	TokenManager    TokenIssuer
	EmailService    EmailSender
	FrontendBaseURL string
	RefreshTTL      time.Duration
}

func NewAuthService(deps AuthServiceDeps) *authService {
	return &authService{
		txManager:       deps.TxManager,
		auditLogger:     deps.AuditLogger,
		userRepo:        deps.UserRepo,
		refreshRepo:     deps.RefreshTokenRepo,
		tokenRepo:       deps.TokenRepo,
		roleRepo:        deps.RoleRepo,
		sessionRepo:     deps.SessionRepo,
		revoStore:       deps.RevoStore,
		tokenManager:    deps.TokenManager,
		emailService:    deps.EmailService,
		frontendBaseURL: deps.FrontendBaseURL,
		refreshTTL:      deps.RefreshTTL,
	}
}

func (s *authService) Register(ctx context.Context, meta *AuditMeta, email, username, password, firstName, lastName string, dob time.Time) error {
	email = normalizeEmail(email)
	username = normalizeUsername(username)
	if err := validateNormalizedEmail(email); err != nil {
		return err
	}
	if err := validateNormalizedUsername(username); err != nil {
		return err
	}

	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if existing != nil && existing.ID != uuid.Nil {
		return ErrEmailExists
	}

	existing, err = s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return err
	}

	if existing != nil && existing.ID != uuid.Nil {
		return ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}
	hashStr := string(hash)

	tokenHash, rawToken, err := s.tokenManager.GenerateHashToken()
	if err != nil {
		return err
	}

	role, err := s.roleRepo.FindByCode(ctx, domain.RoleCodeUser)
	if role == nil || err != nil {
		return ErrInternalServer
	}

	user := &models.User{
		Email:        email,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		DOB:          &dob,
		PasswordHash: &hashStr,
		RoleID:       role.ID,
		Role:         *role,
		Status:       domain.StatusPending,
	}

	// transaction, critical
	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Create(ctx, user); err != nil {
			return err
		}

		if err := s.tokenRepo.
			RevokeByUserAndType(ctx, user.ID, domain.TokenTypeEmailVerification); err != nil {
			return err
		}

		verifyToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: domain.TokenTypeEmailVerification,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, verifyToken); err != nil {
			return err
		}

		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

		if err := s.emailService.SendVerificationEmail(ctx, email, firstName, verifyURL); err != nil {
			return ErrSendVerificationEmail
		}

		target := MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, domain.ActionRegister, nil, target)
	})

	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *authService) LogIn(ctx context.Context, meta *AuditMeta, identifier, password string) (*LoginResult, error) {
	var user *models.User
	var err error

	identifier = normalizeLoginIdentifier(identifier)

	if isEmail(identifier) {
		user, err = s.userRepo.FindByEmail(ctx, identifier)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, identifier)
	}
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrAccountNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrAccountNotVerified
	}

	if user.DeletedAt.Valid {
		return nil, ErrAccountDeleted
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, ErrInternalServer
	}

	now := time.Now()
	refreshTokenHash := s.tokenManager.HashToken(refreshToken)
	refreshExpiresAt := now.Add(s.refreshTTL)

	session := &models.AuthSession{
		UserID:     user.ID,
		ExpiresAt:  refreshExpiresAt,
		LastUsedAt: &now,
	}

	if meta != nil {
		session.UserAgent = meta.UserAgent
		session.IPAddress = meta.IPAddress
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.sessionRepo.Create(ctx, session); err != nil {
			return err
		}

		refreshTokenModel := &models.RefreshToken{
			SessionID: session.ID,
			UserID:    user.ID,
			TokenHash: refreshTokenHash,
			ExpiresAt: refreshExpiresAt,
		}

		if err := s.refreshRepo.Create(txCtx, refreshTokenModel); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, ErrInternalServer
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(
		GenerateAccessTokenInput{
			UserID:    user.ID,
			SessionID: session.ID,
			RoleID:    user.Role.ID,
			RoleCode:  string(user.Role.Code),
			Email:     user.Email,
			Username:  user.Username,
		})
	if err != nil {
		return nil, err
	}

	actor := MapUserToAuditUser(user)
	if err := s.auditLogger.Log(ctx, meta, domain.ActionLogin, actor, nil); err != nil {
		appLogger.Println("failed to log user login action", "error", err)
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokenManager.ExpiresInSeconds(),
		User:         user,
	}, nil
}

func (s *authService) VerifyEmail(ctx context.Context, meta *AuditMeta, rawToken string, tokenType domain.TokenType) error {
	tokenHash := s.tokenManager.HashToken(rawToken)

	found, err := s.tokenRepo.FindValidToken(ctx, tokenHash, tokenType)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, found.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Status == domain.StatusDeleted {
		return ErrUserAlreadyDeleted
	}

	if user.Status == domain.StatusActive {
		return nil
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
			return err
		}

		if err := s.tokenRepo.MarkUsed(ctx, found.ID); err != nil {
			return err
		}

		actor := MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, domain.ActionVerifyEmail, actor, actor)
	})
	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *authService) ResendVerification(ctx context.Context, meta *AuditMeta, email string, tokenType domain.TokenType) error {
	email = normalizeEmail(email)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrUserNotFound
	}

	// revoke old token
	if err = s.tokenRepo.RevokeByUserAndType(ctx, user.ID, tokenType); err != nil {
		return ErrInternalServer
	}

	// generate new token
	tokenHash, rawToken, err := s.tokenManager.GenerateHashToken()
	if err != nil {
		return err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		verifyToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: domain.TokenTypeEmailVerification,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, verifyToken); err != nil {
			return err
		}

		verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

		if err := s.emailService.SendVerificationEmail(ctx, email, user.FirstName, verifyURL); err != nil {
			return ErrSendVerificationEmail
		}

		actor := MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, domain.ActionResendVerification, actor, actor)
	})

	if err != nil {
		return ErrInternalServer
	}

	return nil

}

func (s *authService) ForgotPassword(ctx context.Context, meta *AuditMeta, email string) error {
	email = normalizeEmail(email)

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrInternalServer
	}

	if user == nil {
		return nil
	}

	if user.Status == domain.StatusDeleted {
		return nil
	}

	// generate token
	tokenHash, rawToken, err := s.tokenManager.GenerateHashToken()
	if err != nil {
		return err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {

		// revoke old token
		if err := s.tokenRepo.
			RevokeByUserAndType(ctx, user.ID, domain.TokenTypePasswordReset); err != nil {
			return err
		}

		// create token
		resetToken := &models.UserToken{
			UserID:    user.ID,
			TokenType: domain.TokenTypePasswordReset,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(1 * time.Hour),
		}

		if err := s.tokenRepo.Create(ctx, resetToken); err != nil {
			return err
		}
		actor := MapUserToAuditUser(user)
		return s.auditLogger.
			Log(ctx, meta, domain.ActionForgotPassword, actor, actor)
	})

	if err != nil {
		return ErrInternalServer
	}

	// build link
	resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", s.frontendBaseURL, url.QueryEscape(rawToken))

	// send mail
	if err := s.emailService.SendResetPasswordEmail(ctx, user.Email, user.FirstName, resetURL); err != nil {
		return ErrSendResetPasswordEmail
	}

	return nil
}

func (s *authService) SetPassword(ctx context.Context, meta *AuditMeta, in *SetPasswordInput, tokenType domain.TokenType) error {
	return s.applyPasswordByToken(ctx, meta, in, tokenType)
}

func (s *authService) ResetPassword(ctx context.Context, meta *AuditMeta, in *SetPasswordInput, tokenType domain.TokenType) error {
	return s.applyPasswordByToken(ctx, meta, in, tokenType)
}

func (s *authService) applyPasswordByToken(ctx context.Context, meta *AuditMeta, in *SetPasswordInput, tokenType domain.TokenType) error {
	if in == nil {
		return ErrInvalidInput
	}

	if strings.TrimSpace(in.Token) == "" ||
		strings.TrimSpace(in.Password) == "" ||
		strings.TrimSpace(in.ConfirmPassword) == "" {
		return ErrInvalidInput
	}

	if in.Password != in.ConfirmPassword {
		return ErrPasswordConfirmationMismatch
	}

	tokenHash := s.tokenManager.HashToken(in.Token)

	found, err := s.tokenRepo.FindValidToken(ctx, tokenHash, tokenType)
	if err != nil {
		return ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, found.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Status == domain.StatusDeleted {
		return ErrUserAlreadyDeleted
	}

	if tokenType == domain.TokenTypePasswordSet {
		if user.PasswordHash != nil && *user.PasswordHash != "" {
			return ErrPasswordAlreadySet
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternalServer
	}
	hashStr := string(hash)

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		switch tokenType {
		case domain.TokenTypePasswordSet:
			if err := s.userRepo.UpdatePassword(ctx, user.ID, hashStr); err != nil {
				return err
			}

			if err := s.userRepo.MarkEmailVerified(ctx, user.ID); err != nil {
				return err
			}

		case domain.TokenTypePasswordReset:
			if err := s.userRepo.UpdatePassword(ctx, user.ID, hashStr); err != nil {
				return err
			}

		default:
			return ErrInvalidInput
		}

		if err := s.tokenRepo.MarkUsed(ctx, found.ID); err != nil {
			return err
		}

		action := domain.ActionResetPassword
		if tokenType == domain.TokenTypePasswordSet {
			action = domain.ActionSetPassword
		}

		actor := MapUserToAuditUser(user)
		return s.auditLogger.Log(ctx, meta, action, actor, actor)
	})
	if err != nil {
		if errors.Is(err, ErrPasswordAlreadySet) || errors.Is(err, ErrInvalidInput) {
			return err
		}
		return ErrInternalServer
	}

	return nil
}

func (s *authService) Refresh(ctx context.Context, meta *AuditMeta, refreshToken string) (*RefreshResult, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidInput
	}
	refreshTokenHash := s.tokenManager.HashToken(refreshToken)

	foundToken, err := s.refreshRepo.FindByTokenHash(ctx, refreshTokenHash)
	if err != nil || foundToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	if foundToken.RevokedAt != nil {
		if err := s.handleRefreshTokenReuse(ctx, meta, foundToken); err != nil {
			appLogger.Println(err)
		}
		return nil, ErrInvalidRefreshToken
	}

	now := time.Now().UTC()
	if !foundToken.ExpiresAt.After(now) {
		return nil, ErrInvalidRefreshToken
	}

	session, err := s.sessionRepo.FindActiveByID(ctx, foundToken.SessionID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	if user.DeletedAt.Valid {
		return nil, ErrAccountDeleted
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrAccountNotVerified
	}

	newRefreshTokenHash, newRefreshToken, err := s.tokenManager.GenerateHashToken()
	if err != nil {
		return nil, err
	}
	newRefreshExpiresAt := now.Add(s.refreshTTL)
	if newRefreshExpiresAt.After(session.ExpiresAt) {
		newRefreshExpiresAt = session.ExpiresAt
	}

	refreshTokenModel := &models.RefreshToken{
		SessionID: session.ID,
		UserID:    user.ID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: newRefreshExpiresAt,
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(
		GenerateAccessTokenInput{
			UserID:    user.ID,
			SessionID: session.ID,
			RoleID:    user.Role.ID,
			RoleCode:  string(user.Role.Code),
			Email:     user.Email,
			Username:  user.Username,
		},
	)
	if err != nil {
		return nil, ErrInternalServer
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.refreshRepo.RevokeByID(txCtx, foundToken.ID); err != nil {
			return err
		}

		if err := s.refreshRepo.Create(txCtx, refreshTokenModel); err != nil {
			return err
		}

		if err := s.refreshRepo.MarkReplacement(txCtx, foundToken.ID, refreshTokenModel.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, ErrInternalServer
	}

	return &RefreshResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.tokenManager.ExpiresInSeconds(),
	}, nil
}

func (s *authService) handleRefreshTokenReuse(ctx context.Context, meta *AuditMeta, reused *models.RefreshToken) error {
	if reused == nil || reused.UserID == uuid.Nil {
		return nil
	}

	sessions, err := s.sessionRepo.ListActiveByUserID(ctx, reused.UserID)
	if err != nil {
		return ErrInternalServer
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.sessionRepo.RevokeAllByUserID(txCtx, reused.UserID); err != nil && !errors.Is(err, repositories.ErrNotFound) {
			return err
		}

		if err := s.refreshRepo.RevokeByUserID(txCtx, reused.UserID); err != nil && !errors.Is(err, repositories.ErrNotFound) {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := s.revoStore.MarkRevoked(ctx, session.ID, session.ExpiresAt); err != nil {
			appLogger.Println(err)
		}
	}

	actor := &AuditUser{ID: reused.UserID}
	metadata := map[string]any{
		"event":            "refresh_token_reuse_detected",
		"session_id":       reused.SessionID.String(),
		"refresh_token_id": reused.ID.String(),
	}

	if err := s.auditLogger.LogWithMetadata(ctx, meta, domain.ActionRefreshTokenReuseDetected, actor, actor, metadata); err != nil {
		appLogger.Println(err)
	}

	return nil
}

func (s *authService) Logout(ctx context.Context, meta *AuditMeta, actor *AuditUser, sessionID uuid.UUID) error {
	if actor == nil {
		return ErrUnauthorized
	}

	if sessionID == uuid.Nil {
		return ErrInvalidInput
	}

	session, err := s.sessionRepo.FindActiveByID(ctx, sessionID)
	if err != nil || session == nil {
		return ErrInvalidInput
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.sessionRepo.RevokeByID(txCtx, sessionID); err != nil && !errors.Is(err, repositories.ErrNotFound) {
			return err
		}

		if err := s.refreshRepo.RevokeBySessionID(txCtx, sessionID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return ErrInternalServer
	}

	if err := s.revoStore.MarkRevoked(ctx, session.ID, session.ExpiresAt); err != nil {
		appLogger.Println(err)
	}

	if err := s.auditLogger.Log(ctx, meta, domain.ActionLogout, actor, actor); err != nil {
		appLogger.Println(err)
	}

	return nil
}

func (s *authService) LogoutAll(ctx context.Context, meta *AuditMeta, actor *AuditUser) error {
	if actor == nil {
		return ErrUnauthorized
	}

	if actor.ID == uuid.Nil {
		return ErrInvalidInput
	}

	sessions, err := s.sessionRepo.ListActiveByUserID(ctx, actor.ID)
	if err != nil {
		return ErrInternalServer
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.sessionRepo.RevokeAllByUserID(txCtx, actor.ID); err != nil && !errors.Is(err, repositories.ErrNotFound) {
			return err
		}

		if err := s.refreshRepo.RevokeByUserID(txCtx, actor.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return ErrInternalServer
	}

	for _, session := range sessions {
		if err := s.revoStore.MarkRevoked(ctx, session.ID, session.ExpiresAt); err != nil {
			appLogger.Println(err)
		}
	}

	if err := s.auditLogger.Log(ctx, meta, domain.ActionLogout, actor, actor); err != nil {
		appLogger.Println(err)
	}

	return nil
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
