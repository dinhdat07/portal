package service

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"

	"github.com/google/uuid"
)

type RoleService interface {
	CreateRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, in CreateRoleInput) (*model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	AssignPermission(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, permID uuid.UUID) error
	RemovePermission(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, permID uuid.UUID) error
	DeleteRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, replaceRoleID *uuid.UUID) error
}

type RoleServiceDeps struct {
	RoleRepo       repository.RoleRepository
	PermissionRepo repository.PermissionRepository
	UserRepo       repository.UserRepository
	TxManager      repository.TxManager
	AuditLogger    AuditLogger
}

type roleService struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
	txManager      repository.TxManager
	auditLogger    AuditLogger
}

func NewRoleService(deps RoleServiceDeps) *roleService {
	return &roleService{
		roleRepo:       deps.RoleRepo,
		permissionRepo: deps.PermissionRepo,
		userRepo:       deps.UserRepo,
		txManager:      deps.TxManager,
		auditLogger:    deps.AuditLogger,
	}
}

func (s *roleService) CreateRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, in CreateRoleInput) (*model.Role, error) {
	if in.Code == "" || in.Name == "" {
		return nil, ErrInvalidInput
	}

	existing, err := s.roleRepo.FindByCode(ctx, in.Code)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, ErrInternalServer
	}
	if existing != nil {
		return nil, ErrRoleCodeExists
	}

	role := &model.Role{
		ID:       uuid.New(),
		Code:     in.Code,
		Name:     in.Name,
		IsSystem: false,
	}

	err = s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if err := s.roleRepo.Create(ctx, role); err != nil {
			return ErrInternalServer
		}

		return s.auditLogger.LogWithMetadata(
			ctx,
			meta,
			domain.ActionCreateRole,
			actor,
			nil,
			map[string]any{
				"role_id":   role.ID,
				"role_code": role.Code,
				"role_name": role.Name,
			},
		)
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) ListRoles(ctx context.Context) ([]model.Role, error) {
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, ErrInternalServer
	}
	return roles, nil
}

func (s *roleService) AssignPermission(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, permID uuid.UUID) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context) error {
		role, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrRoleNotFound
			}
			return ErrInternalServer
		}

		perm, err := s.permissionRepo.FindByID(ctx, permID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrPermissionNotFound
			}
			return ErrInternalServer
		}

		if role.IsSystem {
			return ErrCannotModifySystemRole
		}

		if err := s.roleRepo.AssignPermission(ctx, roleID, permID); err != nil {
			return ErrInternalServer
		}

		return s.auditLogger.LogWithMetadata(
			ctx,
			meta,
			domain.ActionAssignPermission,
			actor,
			nil,
			map[string]any{
				"role_id":         role.ID,
				"role_code":       role.Code,
				"role_name":       role.Name,
				"permission_id":   perm.ID,
				"permission_code": perm.Code,
			},
		)
	})
}

func (s *roleService) RemovePermission(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, permID uuid.UUID) error {
	return s.txManager.WithTx(ctx, func(ctx context.Context) error {
		role, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrRoleNotFound
			}
			return ErrInternalServer
		}

		perm, err := s.permissionRepo.FindByID(ctx, permID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrPermissionNotFound
			}
			return ErrInternalServer
		}

		if role.IsSystem {
			return ErrCannotModifySystemRole
		}

		if err := s.roleRepo.RemovePermission(ctx, roleID, permID); err != nil {
			return ErrInternalServer
		}

		return s.auditLogger.LogWithMetadata(
			ctx,
			meta,
			domain.ActionRemovePermission,
			actor,
			nil,
			map[string]any{
				"role_id":         role.ID,
				"role_code":       role.Code,
				"role_name":       role.Name,
				"permission_id":   perm.ID,
				"permission_code": perm.Code,
			},
		)
	})
}

func (s *roleService) DeleteRole(ctx context.Context, meta *AuditMeta, actor *AuditUser, roleID uuid.UUID, replaceRoleID *uuid.UUID) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRoleNotFound
		}
		return ErrInternalServer
	}

	if role.IsSystem {
		return ErrCannotModifySystemRole
	}

	used, err := s.userRepo.ExistsByRoleIDUnscoped(ctx, roleID)
	if err != nil {
		return ErrInternalServer
	}

	if used && replaceRoleID == nil {
		return ErrRoleInUse
	}

	return s.txManager.WithTx(ctx, func(ctx context.Context) error {
		if replaceRoleID != nil {
			if *replaceRoleID == roleID {
				return ErrInvalidReplacementRole
			}

			_, err := s.roleRepo.FindByID(ctx, *replaceRoleID)
			if err != nil {
				if errors.Is(err, repository.ErrNotFound) {
					return ErrInvalidReplacementRole
				}
				return ErrInternalServer
			}

			if err := s.userRepo.UpdateRoleByRoleIDUnscoped(ctx, roleID, *replaceRoleID); err != nil {
				return ErrInternalServer
			}
		}

		if err := s.roleRepo.Delete(ctx, roleID); err != nil {
			return ErrInternalServer
		}

		data := map[string]any{
			"role_id":   role.ID,
			"role_code": role.Code,
			"role_name": role.Name,
		}

		if replaceRoleID != nil {
			data["replacement_role_id"] = *replaceRoleID
		}

		return s.auditLogger.LogWithMetadata(
			ctx,
			meta,
			domain.ActionDeleteRole,
			actor,
			nil,
			data,
		)
	})
}
