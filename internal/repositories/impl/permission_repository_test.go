package impl

import (
	"testing"

	"portal-system/internal/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGormPermissionRepository_FindByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, tx := newTestTx(t)
		repo := NewGormPermissionRepository(testDB)
		perm := mustCreatePermission(t, tx, "roles:list", "roles:list")

		found, err := repo.FindByID(ctx, perm.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		require.Equal(t, perm.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		ctx, _ := newTestTx(t)
		repo := NewGormPermissionRepository(testDB)

		found, err := repo.FindByID(ctx, uuid.New())
		require.Nil(t, found)
		require.ErrorIs(t, err, repositories.ErrNotFound)
	})
}

func TestGormPermissionRepository_List(t *testing.T) {
	ctx, tx := newTestTx(t)
	repo := NewGormPermissionRepository(testDB)
	mustCreatePermission(t, tx, "roles:update", "B Role Update")
	mustCreatePermission(t, tx, "roles:create", "A Role Create")

	perms, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, perms, 2)
	require.Equal(t, "A Role Create", perms[0].Name)
	require.Equal(t, "B Role Update", perms[1].Name)
}
