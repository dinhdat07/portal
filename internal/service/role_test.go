package service_test

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"portal-system/internal/repository"
	repositorymock "portal-system/internal/repository/mock"
	. "portal-system/internal/service"
	servicemock "portal-system/internal/service/mock"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRoleService_CreateRole(t *testing.T) {
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeAdmin}
	input := CreateRoleInput{Code: domain.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}

	t.Run("invalid input", func(t *testing.T) {
		svc := newRoleServiceForTest(nil, nil, nil, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, CreateRoleInput{})
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInvalidInput)
	})

	t.Run("unexpected find error", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, errors.New("lookup failed"))

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInternalServer)
	})

	t.Run("role code exists", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(&model.Role{ID: uuid.New(), Code: input.Code}, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrRoleCodeExists)
	})

	t.Run("create error", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, repository.ErrNotFound)
		roleRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("insert failed")).Once()

		tx := repositorymock.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, nil, roleRepo, nil, nil)
		role, err := svc.CreateRole(context.Background(), meta, actor, input)
		require.Nil(t, role)
		require.ErrorIs(t, err, ErrInternalServer)
	})

	t.Run("success", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByCode(mock.Anything, input.Code).Return(nil, repository.ErrNotFound)
		roleRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		auditLogger := servicemock.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, domain.ActionCreateRole, actor, (*AuditUser)(nil), mock.Anything).Return(nil).Once()
		tx := repositorymock.NewTxManager(t)
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
		roleRepo := repositorymock.NewRoleRepository(t)
		expected := []model.Role{{ID: uuid.New(), Name: "Admin"}}
		roleRepo.EXPECT().List(context.Background()).Return(expected, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		roles, err := svc.ListRoles(context.Background())
		require.NoError(t, err)
		require.Equal(t, expected, roles)
	})

	t.Run("repository error", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().List(context.Background()).Return(nil, errors.New("db error"))

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		roles, err := svc.ListRoles(context.Background())
		require.Nil(t, roles)
		require.ErrorIs(t, err, ErrInternalServer)
	})
}

func TestRoleService_AssignPermission(t *testing.T) {
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeAdmin}
	roleID := uuid.New()
	permID := uuid.New()

	t.Run("role not found", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(nil, repository.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrRoleNotFound)
	})

	t.Run("permission not found", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(&model.Role{ID: roleID, Code: domain.RoleCode("ROLE_CODE_MANAGER")}, nil)
		permRepo := repositorymock.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(nil, repository.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrPermissionNotFound)
	})

	t.Run("system role cannot be modified", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(&model.Role{ID: roleID, IsSystem: true}, nil)
		permRepo := repositorymock.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(&model.Permission{ID: permID}, nil)

		svc := newRoleServiceForTest(nil, nil, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.ErrorIs(t, err, ErrCannotModifySystemRole)
	})

	t.Run("success", func(t *testing.T) {
		role := &model.Role{ID: roleID, Code: domain.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}
		perm := &model.Permission{ID: permID, Code: "roles:list"}

		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().AssignPermission(mock.Anything, roleID, permID).Return(nil)
		permRepo := repositorymock.NewPermissionRepository(t)
		permRepo.EXPECT().FindByID(mock.Anything, permID).Return(perm, nil)
		auditLogger := servicemock.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, domain.ActionAssignPermission, actor, (*AuditUser)(nil), mock.Anything).Return(nil).Once()

		svc := newRoleServiceForTest(nil, auditLogger, roleRepo, permRepo, nil)
		err := svc.AssignPermission(context.Background(), meta, actor, roleID, permID)
		require.NoError(t, err)
	})
}

func TestRoleService_RemovePermission(t *testing.T) {
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeAdmin}
	roleID := uuid.New()
	permID := uuid.New()
	role := &model.Role{ID: roleID, Code: domain.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}
	perm := &model.Permission{ID: permID, Code: "roles:list"}

	roleRepo := repositorymock.NewRoleRepository(t)
	roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
	roleRepo.EXPECT().RemovePermission(mock.Anything, roleID, permID).Return(nil)
	permRepo := repositorymock.NewPermissionRepository(t)
	permRepo.EXPECT().FindByID(mock.Anything, permID).Return(perm, nil)
	auditLogger := servicemock.NewAuditLogger(t)
	auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, domain.ActionRemovePermission, actor, (*AuditUser)(nil), mock.Anything).Return(nil).Once()

	svc := newRoleServiceForTest(nil, auditLogger, roleRepo, permRepo, nil)
	err := svc.RemovePermission(context.Background(), meta, actor, roleID, permID)
	require.NoError(t, err)
}

func TestRoleService_DeleteRole(t *testing.T) {
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeAdmin}
	roleID := uuid.New()
	replacementID := uuid.New()
	role := &model.Role{ID: roleID, Code: domain.RoleCode("ROLE_CODE_MANAGER"), Name: "Manager"}

	t.Run("role not found", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(nil, repository.ErrNotFound)

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, nil)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, nil)
		require.ErrorIs(t, err, ErrRoleNotFound)
	})

	t.Run("role in use without replacement", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		userRepo := repositorymock.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()

		svc := newRoleServiceForTest(nil, nil, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, nil)
		require.ErrorIs(t, err, ErrRoleInUse)
	})

	t.Run("invalid replacement role", func(t *testing.T) {
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().FindByID(mock.Anything, replacementID).Return(nil, repository.ErrNotFound)
		userRepo := repositorymock.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()

		tx := repositorymock.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, nil, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, &replacementID)
		require.ErrorIs(t, err, ErrInvalidReplacementRole)
	})

	t.Run("success with replacement", func(t *testing.T) {
		replacementRole := &model.Role{ID: replacementID, Code: domain.RoleCodeUser, Name: "User"}
		roleRepo := repositorymock.NewRoleRepository(t)
		roleRepo.EXPECT().FindByID(mock.Anything, roleID).Return(role, nil)
		roleRepo.EXPECT().FindByID(mock.Anything, replacementID).Return(replacementRole, nil)
		roleRepo.On("Delete", mock.Anything, roleID).Return(nil).Once()
		userRepo := repositorymock.NewUserRepository(t)
		userRepo.On("ExistsByRoleIDUnscoped", mock.Anything, roleID).Return(true, nil).Once()
		userRepo.On("UpdateRoleByRoleIDUnscoped", mock.Anything, roleID, replacementID).Return(nil).Once()
		auditLogger := servicemock.NewAuditLogger(t)
		auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, domain.ActionDeleteRole, actor, (*AuditUser)(nil), mock.Anything).Return(nil).Once()
		tx := repositorymock.NewTxManager(t)
		tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

		svc := newRoleServiceForTest(tx, auditLogger, roleRepo, nil, userRepo)
		err := svc.DeleteRole(context.Background(), meta, actor, roleID, &replacementID)
		require.NoError(t, err)
	})
}
