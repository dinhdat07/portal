package services

import "errors"

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailBlacklisted   = errors.New("email is blacklisted")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountNotVerified = errors.New("account not verified")
	ErrAccountDeleted     = errors.New("account deleted")
)
