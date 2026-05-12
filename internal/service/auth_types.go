package service

import (
	"portal-system/internal/model"
	"time"

	"github.com/google/uuid"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         *model.User
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type SetPasswordInput struct {
	Token           string
	Password        string
	ConfirmPassword string
}

type RefreshInput struct {
	SessionID    uuid.UUID
	NewTokenHash string
	NewExpiresAt time.Time
	RotatedAt    time.Time
}
