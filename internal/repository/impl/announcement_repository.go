package impl

import (
	"context"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormAnnouncementRepository struct {
	db *gorm.DB
}

func NewGormAnnouncementRepository(db *gorm.DB) repository.AnnouncementRepository {
	return &GormAnnouncementRepository{db: db}
}

func (r *GormAnnouncementRepository) Create(ctx context.Context, announcement *model.Announcement) error {
	return r.getDB(ctx).Create(announcement).Error
}

func (r *GormAnnouncementRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.getDB(ctx).Preload("Creator").First(&announcement, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &announcement, nil
}

func (r *GormAnnouncementRepository) FindAll(ctx context.Context, page, pageSize int, announcementType string) ([]model.Announcement, int64, error) {
	var announcements []model.Announcement
	var total int64

	query := r.getDB(ctx).Model(&model.Announcement{})

	if announcementType != "" {
		query = query.Where("type = ?", announcementType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Preload("Creator").Find(&announcements).Error
	if err != nil {
		return nil, 0, err
	}

	return announcements, total, nil
}

func (r *GormAnnouncementRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.AnnouncementStatus) error {
	return r.getDB(ctx).Model(&model.Announcement{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GormAnnouncementRepository) FindPending(ctx context.Context, limit int) ([]model.Announcement, error) {
	var announcements []model.Announcement
	err := r.getDB(ctx).Where("status = ?", model.AnnouncementStatusPending).Order("created_at ASC").Limit(limit).Find(&announcements).Error
	return announcements, err
}

func (r *GormAnnouncementRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return r.db.WithContext(ctx)
}
