package services_test

import (
	"context"
	"errors"
	"portal-system/internal/domain"
	"portal-system/internal/models"
	"portal-system/internal/repositories"
	repositoriesmocks "portal-system/internal/repositories/mocks"
	. "portal-system/internal/services"
	servicesmocks "portal-system/internal/services/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_UpdateProfile_Table(t *testing.T) {
	adminRoleID := uuid.New()
	userRoleID := uuid.New()
	userID := uuid.New()
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	baseUser := &models.User{
		ID:        userID,
		Username:  "alice",
		FirstName: "Alice",
		LastName:  "Nguyen",
		Email:     "alice@example.com",
		RoleID:    userRoleID,
		Role: models.Role{
			Code: domain.RoleCodeUser,
		},
	}

	newUsername := "alice2"
	newFirstName := "Alicia"
	newLastName := "Tran"
	dob := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		actor           *AuditUser
		input           UpdateUserInput
		findByIDErr     error
		roleFindErr     error
		foundUser       *models.User
		duplicateUser   *models.User
		findUsernameErr error
		updateErr       error
		auditErr        error
		expectedErr     error
		expectTx        bool
		expectUpdate    bool
		expectAudit     bool
		expectAction    domain.ActionName
	}{
		{
			name:        "user not found",
			actor:       &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:       UpdateUserInput{},
			findByIDErr: repositories.ErrNotFound,
			expectedErr: ErrUserNotFound,
		},
		{
			name:        "role repo error",
			actor:       &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:       UpdateUserInput{},
			foundUser:   cloneUser(baseUser),
			roleFindErr: errors.New("role repo failed"),
			expectedErr: ErrInternalServer,
		},
		{
			name:  "cannot update other admin",
			actor: &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeUser},
			input: UpdateUserInput{},
			foundUser: &models.User{
				ID:        uuid.New(),
				Username:  "boss",
				Email:     "boss@example.com",
				FirstName: "Boss",
				LastName:  "Admin",
				RoleID:    adminRoleID,
				Role: models.Role{
					Code: domain.RoleCodeAdmin,
				},
			},
			expectedErr: ErrForbidden,
		},
		{
			name:      "duplicate username",
			actor:     &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:     UpdateUserInput{Username: &newUsername},
			foundUser: cloneUser(baseUser),
			duplicateUser: &models.User{
				ID:       uuid.New(),
				Username: newUsername,
			},
			expectedErr: ErrUsernameExists,
		},
		{
			name:            "find username repository error",
			actor:           &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:           UpdateUserInput{Username: &newUsername},
			foundUser:       cloneUser(baseUser),
			findUsernameErr: errors.New("query failed"),
			expectedErr:     errors.New("query failed"),
		},
		{
			name:         "update repository error in tx",
			actor:        &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:        UpdateUserInput{FirstName: &newFirstName},
			foundUser:    cloneUser(baseUser),
			updateErr:    errors.New("save failed"),
			expectedErr:  ErrInternalServer,
			expectTx:     true,
			expectUpdate: true,
		},
		{
			name:         "audit error in tx",
			actor:        &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser},
			input:        UpdateUserInput{LastName: &newLastName},
			foundUser:    cloneUser(baseUser),
			auditErr:     errors.New("audit failed"),
			expectedErr:  ErrAuditLogger,
			expectTx:     true,
			expectUpdate: true,
			expectAudit:  true,
			expectAction: domain.ActionUpdateProfile,
		},
		{
			name:  "success user updates own profile",
			actor: &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser, Username: "alice", Email: "alice@example.com"},
			input: UpdateUserInput{
				Username:  &newUsername,
				FirstName: &newFirstName,
				LastName:  &newLastName,
				DOB:       &dob,
			},
			foundUser:    cloneUser(baseUser),
			expectTx:     true,
			expectUpdate: true,
			expectAudit:  true,
			expectAction: domain.ActionUpdateProfile,
		},
		{
			name:         "success admin updates user uses admin action",
			actor:        &AuditUser{ID: userID, RoleCode: domain.RoleCodeAdmin, Username: "admin", Email: "admin@example.com"},
			input:        UpdateUserInput{FirstName: &newFirstName},
			foundUser:    cloneUser(baseUser),
			expectTx:     true,
			expectUpdate: true,
			expectAudit:  true,
			expectAction: domain.ActionAdminUpdateUser,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			adminRole := &models.Role{ID: adminRoleID, Code: domain.RoleCodeAdmin}
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().LogWithMetadata(mock.Anything, meta, mock.Anything, tc.actor, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, meta *AuditMeta, action domain.ActionName, actor *AuditUser, target *AuditUser, data map[string]any) error {
				if tc.expectAction != "" && action != tc.expectAction {
					t.Fatalf("expected audit action %s, got %s", tc.expectAction, action)
				}
				return tc.auditErr
			}).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByID(mock.Anything, userID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				if tc.findByIDErr != nil {
					return nil, tc.findByIDErr
				}
				return cloneUser(tc.foundUser), nil
			})
			userRepo.EXPECT().FindByUsername(mock.Anything, newUsername).RunAndReturn(func(ctx context.Context, username string) (*models.User, error) {
				if tc.findUsernameErr != nil {
					return nil, tc.findUsernameErr
				}
				return tc.duplicateUser, nil
			}).Maybe()
			userRepo.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, user *models.User) error {
				return tc.updateErr
			}).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			roleRepo.EXPECT().FindByCode(mock.Anything, domain.RoleCodeAdmin).RunAndReturn(func(ctx context.Context, code domain.RoleCode) (*models.Role, error) {
				if tc.roleFindErr != nil {
					return nil, tc.roleFindErr
				}
				return adminRole, nil
			}).Maybe()
			tx := repositoriesmocks.NewTxManager(t)
			tx.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			svc := NewUserService(UserServiceDeps{
				TxManager:   tx,
				AuditLogger: auditLogger,
				RoleRepo:    roleRepo,
				UserRepo:    userRepo,
			})

			updated, err := svc.UpdateProfile(context.Background(), meta, tc.actor, userID, tc.input)

			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				if updated == nil {
					t.Fatal("expected updated user")
				}
			} else {
				if tc.name == "find username repository error" {
					if err == nil || err.Error() != tc.expectedErr.Error() {
						t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
					}
				} else if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			}

			if tc.expectTx {
				tx.AssertNumberOfCalls(t, "WithTx", 1)
			} else {
				tx.AssertNotCalled(t, "WithTx", mock.Anything, mock.Anything)
			}
			if tc.expectUpdate {
				userRepo.AssertNumberOfCalls(t, "Update", 1)
			} else {
				userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			}
			if tc.expectAudit {
				auditLogger.AssertNumberOfCalls(t, "LogWithMetadata", 1)
			} else {
				auditLogger.AssertNotCalled(t, "LogWithMetadata", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestUserService_UpdateProfile_NormalizesCaseOnlyUsernameChange(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: userID, RoleCode: domain.RoleCodeUser, Username: "alice", Email: "alice@example.com"}
	inputUsername := " ALICE "

	userRepo := repositoriesmocks.NewUserRepository(t)
	userRepo.EXPECT().FindByID(mock.Anything, userID).Return(&models.User{
		ID:       userID,
		Username: "Alice",
		Email:    "alice@example.com",
		RoleID:   roleID,
		Role:     models.Role{Code: domain.RoleCodeUser},
	}, nil)

	var updatedUsername string
	userRepo.EXPECT().Update(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, user *models.User) error {
		updatedUsername = user.Username
		return nil
	})

	roleRepo := repositoriesmocks.NewRoleRepository(t)
	roleRepo.EXPECT().FindByCode(mock.Anything, domain.RoleCodeAdmin).Return(&models.Role{ID: uuid.New(), Code: domain.RoleCodeAdmin}, nil)

	svc := NewUserService(UserServiceDeps{
		TxManager:   newPassthroughTxManager(),
		AuditLogger: newAuditLoggerMock(),
		RoleRepo:    roleRepo,
		UserRepo:    userRepo,
	})

	user, err := svc.UpdateProfile(context.Background(), meta, actor, userID, UpdateUserInput{Username: &inputUsername})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil || user.Username != "alice" || updatedUsername != "alice" {
		t.Fatalf("expected username to be normalized to alice, got user=%#v updated=%q", user, updatedUsername)
	}
	userRepo.AssertNotCalled(t, "FindByUsername", mock.Anything, mock.Anything)
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	userRepo := repositoriesmocks.NewUserRepository(t)
	userRepo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(nil, repositories.ErrNotFound)
	roleRepo := repositoriesmocks.NewRoleRepository(t)
	tx := repositoriesmocks.NewTxManager(t)
	auditLogger := servicesmocks.NewAuditLogger(t)
	svc := NewUserService(UserServiceDeps{
		UserRepo:    userRepo,
		RoleRepo:    roleRepo,
		TxManager:   tx,
		AuditLogger: auditLogger,
	})

	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeUser}
	_, err := svc.GetProfile(context.Background(), &AuditMeta{}, actor, uuid.New())

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetProfile_AdminWritesAuditLog(t *testing.T) {
	userID := uuid.New()
	var capturedAction domain.ActionName
	auditLogger := servicesmocks.NewAuditLogger(t)
	auditLogger.EXPECT().Log(mock.Anything, mock.Anything, domain.ActionAdminViewUser, mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, meta *AuditMeta, action domain.ActionName, actor *AuditUser, target *AuditUser) error {
		capturedAction = action
		return nil
	}).Once()
	userRepo := repositoriesmocks.NewUserRepository(t)
	userRepo.EXPECT().FindByID(mock.Anything, userID).Return(&models.User{
		ID:       userID,
		Username: "target",
		Email:    "target@example.com",
		Role: models.Role{
			Code: domain.RoleCodeUser,
		},
	}, nil)
	roleRepo := repositoriesmocks.NewRoleRepository(t)
	tx := repositoriesmocks.NewTxManager(t)
	svc := NewUserService(UserServiceDeps{
		UserRepo:    userRepo,
		RoleRepo:    roleRepo,
		TxManager:   tx,
		AuditLogger: auditLogger,
	})

	actor := &AuditUser{ID: uuid.New(), RoleCode: domain.RoleCodeAdmin, Username: "admin", Email: "admin@example.com"}
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	user, err := svc.GetProfile(context.Background(), meta, actor, userID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if user == nil || user.ID != userID {
		t.Fatalf("expected user with id %s", userID)
	}
	if !auditLogger.AssertNumberOfCalls(t, "Log", 1) {
		t.Fatal("expected exactly one audit log create call")
	}
	if capturedAction != domain.ActionAdminViewUser {
		t.Fatalf("expected action %s, got %s", domain.ActionAdminViewUser, capturedAction)
	}
}

func TestUserService_ChangePassword_Table(t *testing.T) {
	actorID := uuid.New()
	meta := &AuditMeta{IPAddress: "127.0.0.1", UserAgent: "unit-test"}
	actor := &AuditUser{ID: actorID, Username: "user1", Email: "u1@example.com", RoleCode: domain.RoleCodeUser}
	hashedCurrent, err := bcrypt.GenerateFromPassword([]byte("current-pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("cannot hash password for test: %v", err)
	}

	tests := []struct {
		name              string
		currentPassword   string
		newPassword       string
		confirmPassword   string
		user              *models.User
		findByIDErr       error
		updatePasswordErr error
		auditErr          error
		expectedErr       error
		expectTx          bool
		expectUpdateCall  bool
		expectAuditCall   bool
	}{
		{
			name:            "user not found maps to unauthorized",
			currentPassword: "current-pass",
			newPassword:     "new-pass-123",
			confirmPassword: "new-pass-123",
			findByIDErr:     repositories.ErrNotFound,
			expectedErr:     ErrUnauthorized,
		},
		{
			name:            "blank new password",
			currentPassword: "current-pass",
			newPassword:     "   ",
			confirmPassword: "new-pass-123",
			user:            &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			expectedErr:     ErrInvalidInput,
		},
		{
			name:            "nil hash unauthorized",
			currentPassword: "current-pass",
			newPassword:     "new-pass-123",
			confirmPassword: "new-pass-123",
			user:            &models.User{ID: actorID, PasswordHash: nil},
			expectedErr:     ErrUnauthorized,
		},
		{
			name:            "confirm mismatch",
			currentPassword: "current-pass",
			newPassword:     "new-pass-123",
			confirmPassword: "another-pass",
			user:            &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			expectedErr:     ErrPasswordConfirmationMismatch,
		},
		{
			name:            "incorrect current password",
			currentPassword: "wrong-pass",
			newPassword:     "new-pass-123",
			confirmPassword: "new-pass-123",
			user:            &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			expectedErr:     ErrIncorrectPassword,
		},
		{
			name:            "new equals current",
			currentPassword: "current-pass",
			newPassword:     "current-pass",
			confirmPassword: "current-pass",
			user:            &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			expectedErr:     ErrNewPasswordMustBeDifferent,
		},
		{
			name:              "update password repository error maps internal",
			currentPassword:   "current-pass",
			newPassword:       "new-pass-123",
			confirmPassword:   "new-pass-123",
			user:              &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			updatePasswordErr: errors.New("db write failed"),
			expectedErr:       ErrInternalServer,
			expectTx:          true,
			expectUpdateCall:  true,
		},
		{
			name:             "audit write error maps audit logger error",
			currentPassword:  "current-pass",
			newPassword:      "new-pass-123",
			confirmPassword:  "new-pass-123",
			user:             &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			auditErr:         errors.New("audit failed"),
			expectedErr:      ErrAuditLogger,
			expectTx:         true,
			expectUpdateCall: true,
			expectAuditCall:  true,
		},
		{
			name:             "success",
			currentPassword:  "current-pass",
			newPassword:      "new-pass-123",
			confirmPassword:  "new-pass-123",
			user:             &models.User{ID: actorID, PasswordHash: ptrString(string(hashedCurrent))},
			expectTx:         true,
			expectUpdateCall: true,
			expectAuditCall:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			auditLogger := servicesmocks.NewAuditLogger(t)
			auditLogger.EXPECT().Log(mock.Anything, meta, domain.ActionChangePassword, actor, actor).RunAndReturn(func(ctx context.Context, meta *AuditMeta, action domain.ActionName, actor *AuditUser, target *AuditUser) error {
				return tc.auditErr
			}).Maybe()
			userRepo := repositoriesmocks.NewUserRepository(t)
			userRepo.EXPECT().FindByID(mock.Anything, actorID).RunAndReturn(func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				if tc.findByIDErr != nil {
					return nil, tc.findByIDErr
				}
				return tc.user, nil
			})
			userRepo.EXPECT().UpdatePassword(mock.Anything, actorID, mock.Anything).RunAndReturn(func(ctx context.Context, id uuid.UUID, passwordHash string) error {
				if id != actorID {
					t.Fatalf("expected update password for actor id %s, got %s", actorID, id)
				}
				if passwordHash == "" {
					t.Fatal("expected non-empty hashed password")
				}
				return tc.updatePasswordErr
			}).Maybe()
			txManager := repositoriesmocks.NewTxManager(t)
			txManager.EXPECT().WithTx(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			}).Maybe()
			roleRepo := repositoriesmocks.NewRoleRepository(t)
			svc := NewUserService(UserServiceDeps{
				TxManager:   txManager,
				AuditLogger: auditLogger,
				RoleRepo:    roleRepo,
				UserRepo:    userRepo,
			})

			err := svc.ChangePassword(context.Background(), meta, actor, tc.currentPassword, tc.newPassword, tc.confirmPassword)

			if tc.expectedErr == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
			} else if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
			}

			if tc.expectTx {
				txManager.AssertNumberOfCalls(t, "WithTx", 1)
			} else {
				txManager.AssertNotCalled(t, "WithTx", mock.Anything, mock.Anything)
			}
			if tc.expectUpdateCall {
				userRepo.AssertNumberOfCalls(t, "UpdatePassword", 1)
			} else {
				userRepo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
			}
			if tc.expectAuditCall {
				auditLogger.AssertNumberOfCalls(t, "Log", 1)
			} else {
				auditLogger.AssertNotCalled(t, "Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
