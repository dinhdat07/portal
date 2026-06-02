package repository

import (
	"context"
	"portal-system/internal/model"

	"github.com/google/uuid"
)

type AnnouncementRepository interface {
	Create(ctx context.Context, announcement *model.Announcement) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Announcement, error)
	FindAll(ctx context.Context, page, pageSize int, announcementType string) ([]model.Announcement, int64, error)
	UpdateDispatchStatus(ctx context.Context, id uuid.UUID, status model.AnnouncementDispatchStatus) error
	FindPending(ctx context.Context, limit int) ([]model.Announcement, error)
}
