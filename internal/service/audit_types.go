package service

import (
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"time"

	"github.com/google/uuid"
)

type AuditLogFilter struct {
	Action       string
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	From         *time.Time
	To           *time.Time
	Page         int
	PageSize     int
}

type AuditMeta struct {
	IPAddress string
	UserAgent string
}

type AuditUser struct {
	ID       uuid.UUID
	Username string
	Email    string
	RoleCode domain.RoleCode
}

func MapUserToAuditUser(u *model.User) *AuditUser {
	if u == nil {
		return nil
	}

	return &AuditUser{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		RoleCode: u.Role.Code,
	}
}
