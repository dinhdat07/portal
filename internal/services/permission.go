package services

import (
	"context"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
)

type PermissionService interface {
	ListPermission(ctx context.Context) ([]models.Permission, error)
}

type permissionService struct {
	repo repositories.PermissionRepository
}

func NewPermissionService(repo repositories.PermissionRepository) *permissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) ListPermission(ctx context.Context) ([]models.Permission, error) {
	perms, err := s.repo.List(ctx)
	if err != nil {
		return nil, ErrInternalServer
	}
	return perms, nil
}
