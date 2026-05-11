package services_test

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/domain/constants"
	"portal-system/internal/domain/enum"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	repositoriesmocks "portal-system/internal/repositories/mocks"
	. "portal-system/internal/services"
	servicesmocks "portal-system/internal/services/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoleService_CreateRole(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	input := domain.CreateRoleInput{Code: constants.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}

	t.Run("invalid input", func(t *testing.T) {
		svc := newRoleServiceForTest(nil, nil, nil, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, domain.CreateRoleInput{})
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("unexpected find error", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, errors.New("lookup failed"))

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInternalServer)
	})

	t.Run("role code exists", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(&models.Role{ID: uuid.New(), Code: input.Code}, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrRoleCodeExists)
	})

	t.Run("create error", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, repositories.ErrNotFound)
		roleRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed")).Once()

		tx := repositoriesmocks.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInternalServer)
	})

	t.Run("success", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, repositories.ErrNotFound)
		roleRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		auditLogger := servicesmocks.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionCreateRole, actor, (*domain.AuditUser)(nil), mock.Anything).Return(nil).Once()
		tx := repositoriesmocks.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, auditLogger, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.NoError(t, err)
		require.NotNil(t, role)
		require.Equal(t, input.Code, role.Code)
		require.Equal(t, input.Name, role.Name)
		require.False(t, role.IsSystem)
		require.NotEqual(t, uuid.Nil, role.ID)
	})
}

func TestRoleService_ListRoles(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		expected := []models.Role{{ID: uuid.New(), Name: "Admin"}}
		roleRepo.EXPECT().List(context.Background()).Return(expected, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		roles, err := svc.ListRoles(context.Background())
		require.NoError(t, err)
		require.Equal(t, expected, roles)
	})

	t.Run("repository error", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().List(context.Background()).Return(nil, errors.New("db error"))

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		roles, err := svc.ListRoles(context.Background())
		require.Nil(t, roles)
		require.ErrorIs(t, err, ErrInternalServer)
	})
}

func TestRoleService_AssignPermission(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	roleID := uuid.New()
	permID := uuid.New()

	t.Run("role not found", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(nil, repositories.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrRoleNotFound)
	})

	t.Run("permission not found", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(&models.Role{ID: roleID, Code: constants.RoleCode("ROLE_CODE_MANAGER")}, nil)
		permRepo := repositoriesmocks.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(nil, repositories.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrPermissionNotFound)
	})

	t.Run("system role cannot be modified", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(&models.Role{ID: roleID, IsSystem: true}, nil)
		permRepo := repositoriesmocks.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(&models.Permission{ID: permID}, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrCannotModifySystemRole)
	})

	t.Run("success", func(t *testing.T) {
		role := &models.Role{ID: roleID, Code: constants.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}
		perm := &models.Permission{ID: permID, Code: "roles:list"}

		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().AssignPermission(mock.Anything, roleID, permID).Return(nil)
		permRepo := repositoriesmocks.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(perm, nil)
		auditLogger := servicesmocks.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionAssignPermission, actor, (*domain.AuditUser)(nil), mock.Anything).Return(nil).Once()

		svc := newRoleServiceForTest(nil, auditLogger, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.NoError(t, err)
	})
}

func TestRoleService_RemovePermission(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	roleID := uuid.New()
	permID := uuid.New()
	role := &models.Role{ID: roleID, Code: constants.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}
	perm := &models.Permission{ID: permID, Code: "roles:list"}

	roleRepo := repositoriesmocks.NewRoleRepository(t)
	roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
	roleRepo.EXPECT().RemovePermission(mock.Anything, roleID, permID).Return(nil)
	permRepo := repositoriesmocks.NewPermissionRepository(t)
	permRepo.EXPECT().FindByID(mock.Anything, permID).Return(perm, nil)
	auditLogger := servicesmocks.NewAuditLogger(t)
	auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionRemovePermission, actor, (*domain.AuditUser)(nil), mock.Anything).Return(nil).Once()

	svc := newRoleServiceForTest(nil, auditLogger, roleRepo, permRepo, nil)
	err := svc.RemovePermission(context.Background(), meta, actor, roleID, permID)
	require.NoError(t, err)
}

func TestRoleService_DeleteRole(t *testing.T) {
	meta := &domain.AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &domain.AuditUser{ID: uuid.New(), RoleCode: constants.RoleCodeAdmin}
	roleID := uuid.New()
	replacementID := uuid.New()
	role := &models.Role{ID: roleID, Code: constants.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}

	t.Run("role not found", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(nil, repositories.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, nil)
		require.ErrorIs(t, err, ErrRoleNotFound)
	})

	t.Run("role in use without replacement", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		userRepo := repositoriesmocks.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, nil)
		require.ErrorIs(t, err, ErrRoleInUse)
	})

	t.Run("invalid replacement role", func(t *testing.T) {
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().FindByID(mock.Anything, replacementID).Return(nil, repositories.ErrNotFound)
		userRepo := repositoriesmocks.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()

		tx := repositoriesmocks.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, nil, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, &replacementID)
		require.ErrorIs(t, err, ErrInvalidReplacementRole)
	})

	t.Run("success with replacement", func(t *testing.T) {
		replacementRole := &models.Role{ID: replacementID, Code: constants.RoleCodeUser, Name: "User"}
		roleRepo := repositoriesmocks.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().FindByID(mock.Anything, replacementID).Return(replacementRole, nil)
		roleRepo.On("Delete", mock.Anything, roleID).Return(nil).Once()
		userRepo := repositoriesmocks.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()
		userRepo.On("UpdateRoleByRoleIDUnscoped", mock.Anything, roleID, replacementID).Return(nil).Once()
		auditLogger := servicesmocks.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, enum.ActionDeleteRole, actor, (*domain.AuditUser)(nil), mock.Anything).Return(nil).Once()
		tx := repositoriesmocks.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, auditLogger, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, &replacementID)
		require.NoError(t, err)
	})
}
