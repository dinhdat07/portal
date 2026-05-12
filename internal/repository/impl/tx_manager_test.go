package impl

import (
	"context"
	"errors"
	"testing"

	"portal-system/internal/domain"
	"portal-system/internal/model"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGormTxManager_WithTx(t *testing.T) {
	t.Run("commits on success", func(t *testing.T) {
		manager := NewGormTxManager(testDB)
		repo := NewGormUserRepository(testDB)
		email := "commit@example.com"
		roleCode := domain.RoleCode("tx_commit_user")

		err := manager.WithTx(context.Background(), func(ctx context.Context) error {
			tx, ok := ctx.Value(txKey{}).(*gorm.DB)
			require.True(t, ok)
			role := mustCreateRole(t, tx, roleCode)
			user := &model.User{
				Email:     email,
				Username:  "commit-user",
				FirstName: "Commit",
				LastName:  "User",
				RoleID:    role.ID,
				Status:    domain.StatusActive,
			}
			return repo.Create(ctx, user)
		})
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = testDB.Unscoped().Where("email = ?", email).Delete(&model.User{}).Error
			_ = testDB.Where("code = ?", string(roleCode)).Delete(&model.Role{}).Error
		})

		found, err := repo.FindByEmail(context.Background(), email)
		require.NoError(t, err)
		require.NotNil(t, found)
	})

	t.Run("rolls back on error", func(t *testing.T) {
		manager := NewGormTxManager(testDB)
		repo := NewGormUserRepository(testDB)
		email := "rollback@example.com"
		roleCode := domain.RoleCode("tx_rollback_admin")

		err := manager.WithTx(context.Background(), func(ctx context.Context) error {
			tx, ok := ctx.Value(txKey{}).(*gorm.DB)
			require.True(t, ok)
			role := mustCreateRole(t, tx, roleCode)
			user := &model.User{
				Email:     email,
				Username:  "rollback-user",
				FirstName: "Rollback",
				LastName:  "User",
				RoleID:    role.ID,
				Status:    domain.StatusActive,
			}
			require.NoError(t, repo.Create(ctx, user))
			return errors.New("boom")
		})
		require.EqualError(t, err, "boom")

		found, findErr := repo.FindByEmail(context.Background(), email)
		require.NoError(t, findErr)
		require.Nil(t, found)
	})
}
