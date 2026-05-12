package repository

import (
	"context"
	"time"

	"portal-system/internal/model"

	"github.com/google/uuid"
)

type AuditLogListFilter struct {
	Action       string
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	From         *time.Time
	To           *time.Time
	Page         int
	PageSize     int
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, filter AuditLogListFilter) ([]model.AuditLog, int64, error)
}
