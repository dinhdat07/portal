package domain

import (
	"portal-system/internal/models"
)

type LoginResult struct {
	AccessToken string
	ExpiresIn   int
	User        *models.User
}

type SetPasswordInput struct {
	Token           string
	Password        string
	ConfirmPassword string
}
