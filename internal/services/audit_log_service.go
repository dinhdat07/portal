package services

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
)

type AuditLogService struct {
	repo *repositories.AuditLogRepository
}

func NewAuditLogService(repo *repositories.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

func (s *AuditLogService) Create(ctx context.Context, log *models.AuditLog) error {
	if !log.Action.IsValid() {
		return ErrInvalidAction
	}
	return s.repo.Create(ctx, log)
}

func (s *AuditLogService) List(ctx context.Context, filter domain.AuditLogFilter) ([]models.AuditLog, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}

	// validate time range
	if filter.From != nil && filter.To != nil {
		if filter.From.After(*filter.To) {
			return nil, 0, ErrInvalidTimeRange
		}
	}

	logs, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
