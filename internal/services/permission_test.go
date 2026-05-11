package services_test

import (
	"context"
	"errors"
	"portal-system/internal/models"
	repositoriesmocks "portal-system/internal/repositories/mocks"
	. "portal-system/internal/services"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPermissionService_ListPermission(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := repositoriesmocks.NewPermissionRepository(t)
		expected := []models.Permission{
			{Code: "permissions:list", Name: "permissions:list"},
			{Code: "roles:create", Name: "roles:create"},
		}
		repo.EXPECT().List(context.Background()).Return(expected, nil)

		svc := newPermissionServiceForTest(repo)
		got, err := svc.ListPermission(context.Background())
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := repositoriesmocks.NewPermissionRepository(t)
		repo.EXPECT().List(context.Background()).Return(nil, errors.New("db error"))

		svc := newPermissionServiceForTest(repo)
		got, err := svc.ListPermission(context.Background())
		require.Nil(t, got)
		require.ErrorIs(t, err, ErrInternalServer)
	})
}
