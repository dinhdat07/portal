package services

import "strings"

const (
	minUsernameLength = 3
	maxUsernameLength = 50
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func normalizeLoginIdentifier(identifier string) string {
	return strings.ToLower(strings.TrimSpace(identifier))
}

func validateNormalizedEmail(email string) error {
	if email == "" {
		return ErrInvalidInput
	}
	return nil
}

func validateNormalizedUsername(username string) error {
	if len(username) < minUsernameLength || len(username) > maxUsernameLength {
		return ErrInvalidInput
	}
	return nil
}
