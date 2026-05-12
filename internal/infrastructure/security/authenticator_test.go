package security_test

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/infrastructure/security"
	"portal-system/internal/model"
	repositorymocks "portal-system/internal/repository/mocks"
	"portal-system/internal/service"
	servicemocks "portal-system/internal/service/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticator_Authenticate_Table(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	roleID := uuid.New()

	role := &model.Role{
		ID:   roleID,
		Code: domain.RoleCodeUser,
		Permissions: []model.Permission{
			{ID: uuid.New(), Code: "user.read", Name: "User Read"},
			{ID: uuid.New(), Code: "user.write", Name: "User Write"},
		},
	}

	tests := []struct {
		name        string
		tokenString string
		revoked     bool
		revokeErr   error
		session     *model.AuthSession
		sessionErr  error
		roleErr     error
		expectedErr string
	}{
		{name: "invalid token", tokenString: "bad-token", expectedErr: "invalid token"},
		{name: "revocation store error ignored", tokenString: "valid-token", revokeErr: errors.New("redis down"), session: &model.AuthSession{ID: sessionID, UserID: userID}},
		{name: "revoked session", tokenString: "valid-token", revoked: true, expectedErr: "session is already revoked"},
		{name: "session lookup error", tokenString: "valid-token", sessionErr: errors.New("not found"), expectedErr: "not found"},
		{name: "session user mismatch", tokenString: "valid-token", session: &model.AuthSession{ID: sessionID, UserID: uuid.New()}, expectedErr: "session does not belong to user"},
		{name: "role lookup error", tokenString: "valid-token", session: &model.AuthSession{ID: sessionID, UserID: userID}, roleErr: errors.New("role missing"), expectedErr: "role missing"},
		{name: "success", tokenString: "valid-token", session: &model.AuthSession{ID: sessionID, UserID: userID}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manager := servicemocks.NewTokenIssuer(t)
			manager.EXPECT().Parse(tc.tokenString).RunAndReturn(func(tokenString string) (*service.TokenClaims, error) {
				if tokenString == "bad-token" {
					return nil, errors.New("invalid token")
				}
				return &service.TokenClaims{
					UserID:    userID,
					SessionID: sessionID,
					Username:  "john",
					Email:     "john@example.com",
					RoleID:    roleID,
					RoleCode:  string(domain.RoleCodeUser),
				}, nil
			}).Maybe()

			roleRepo := repositorymocks.NewRoleRepository(t)
			roleRepo.EXPECT().GetWithPermissions(mock.Anything, roleID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*model.Role, error) {
				if tc.roleErr != nil {
					return nil, tc.roleErr
				}
				return role, nil
			}).Maybe()

			sessionRepo := repositorymocks.NewAuthSessionRepository(t)
			sessionRepo.EXPECT().FindActiveByID(mock.Anything, sessionID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*model.AuthSession, error) {
				if tc.sessionErr != nil {
					return nil, tc.sessionErr
				}
				return tc.session, nil
			}).Maybe()

			revoStore := servicemocks.NewSessionRevocationStore(t)
			revoStore.EXPECT().IsRevoked(mock.Anything, sessionID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (bool, error) {
				return tc.revoked, tc.revokeErr
			}).Maybe()

			authenticator := security.NewAuthenticator(manager, roleRepo, sessionRepo, revoStore)

			principal, err := authenticator.Authenticate(context.Background(), tc.tokenString)
			if tc.expectedErr != "" {
				if err == nil || err.Error() != tc.expectedErr {
					t.Fatalf("expected error %q, got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if principal == nil {
				t.Fatal("expected principal, got nil")
			}
			if principal.UserID != userID || principal.SessionID != sessionID || principal.RoleID != roleID {
				t.Fatalf("unexpected principal identity: %#v", principal)
			}
			if len(principal.Permissions) != 2 || principal.Permissions[0] != "user.read" || principal.Permissions[1] != "user.write" {
				t.Fatalf("unexpected permissions: %#v", principal.Permissions)
			}
		})
	}
}
