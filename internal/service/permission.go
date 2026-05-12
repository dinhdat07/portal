package service

import (
	"context"
	"portal-system/internal/model"
	"portal-system/internal/repository"
)

type PermissionService interface {
	ListPermission(ctx context.Context) ([]model.Permission, error)
}

type permissionService struct {
	repo repository.PermissionRepository
}

func NewPermissionService(repo repository.PermissionRepository) *permissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) ListPermission(ctx context.Context) ([]model.Permission, error) {
	perms, err := s.repo.List(ctx)
	if err != nil {
		return nil, ErrInternalServer
	}
	return perms, nil
}
