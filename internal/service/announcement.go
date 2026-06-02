package service

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
)

type AnnouncementService interface {
	CreateAnnouncement(ctx context.Context, meta *AuditMeta, actor *AuditUser, input CreateAnnouncementInput) (*model.Announcement, error)
	GetAnnouncement(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*model.Announcement, error)
	ListAnnouncements(ctx context.Context, meta *AuditMeta, actor *AuditUser, filter AnnouncementListFilter) (*AnnouncementListResult, error)
}

type announcementService struct {
	announcementRepo repository.AnnouncementRepository
	txManager        repository.TxManager
	auditLogger      AuditLogger
}

func NewAnnouncementService(
	announcementRepo repository.AnnouncementRepository,
	txManager repository.TxManager,
	auditLogger AuditLogger,
) AnnouncementService {
	return &announcementService{
		announcementRepo: announcementRepo,
		txManager:        txManager,
		auditLogger:      auditLogger,
	}
}

func (s *announcementService) CreateAnnouncement(ctx context.Context, meta *AuditMeta, actor *AuditUser, input CreateAnnouncementInput) (*model.Announcement, error) {
	announcement := &model.Announcement{
		Title:          input.Title,
		Content:        input.Content,
		Type:           model.AnnouncementType(input.Type),
		TargetRoles:    input.TargetRoles,
		CreatedBy:      actor.ID,
		DispatchStatus: model.AnnouncementDispatchStatusPending,
	}

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.announcementRepo.Create(txCtx, announcement); err != nil {
			return err
		}

		target := &AuditUser{ID: announcement.ID}
		if err := s.auditLogger.Log(txCtx, meta, domain.ActionAdminCreateAnnouncement, actor, target); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return announcement, nil
}

func (s *announcementService) GetAnnouncement(ctx context.Context, meta *AuditMeta, actor *AuditUser, id uuid.UUID) (*model.Announcement, error) {
	announcement, err := s.announcementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if announcement == nil {
		return nil, repository.ErrNotFound
	}

	target := &AuditUser{ID: announcement.ID}
	_ = s.auditLogger.Log(ctx, meta, domain.ActionAdminViewAnnouncement, actor, target)

	return announcement, nil
}

func (s *announcementService) ListAnnouncements(ctx context.Context, meta *AuditMeta, actor *AuditUser, filter AnnouncementListFilter) (*AnnouncementListResult, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	announcements, total, err := s.announcementRepo.FindAll(ctx, filter.Page, filter.PageSize, filter.AnnouncementType)
	if err != nil {
		return nil, err
	}

	logMeta := map[string]any{
		"filters": map[string]any{
			"type": filter.AnnouncementType,
		},
		"pagination": map[string]any{
			"page":      filter.Page,
			"page_size": filter.PageSize,
		},
		"result_count": len(announcements),
		"total":        total,
	}

	_ = s.auditLogger.LogWithMetadata(ctx, meta, domain.ActionAdminSearchAnnouncement, actor, nil, logMeta)

	return &AnnouncementListResult{
		Announcements: announcements,
		Total:         total,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
	}, nil
}
