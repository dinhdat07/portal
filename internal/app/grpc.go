package app

import (
	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"
	"portal-system/internal/domain"
	"portal-system/internal/handler/interceptor"

	"google.golang.org/grpc"
)

func (a *App) NewGRPCServer() *grpc.Server {
	publicMethods := buildGRPCPublicMethods()
	methodPermissions := buildGRPCMethodPermissions()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.PreAuthRateLimitInterceptor(a.RateLimiter, a.RateLimitKeyBuilder, a.RateLimitConfig),
			interceptor.ValidationInterceptor(a.Validator),
			interceptor.AuthenticationInterceptor(a.Authenticator, publicMethods),
			interceptor.PostAuthRateLimitInterceptor(a.RateLimiter, a.RateLimitKeyBuilder, a.RateLimitConfig),
			interceptor.PermissionInterceptor(a.Authorizer, methodPermissions),
		),
	)

	authv1.RegisterAuthServiceServer(s, a.AuthGRPC)
	adminv1.RegisterAdminServiceServer(s, a.AdminGRPC)
	userv1.RegisterUserServiceServer(s, a.UserGRPC)

	return s
}

func buildGRPCPublicMethods() map[string]bool {
	return map[string]bool{
		// auth public
		authv1.AuthService_Register_FullMethodName:           true,
		authv1.AuthService_Login_FullMethodName:              true,
		authv1.AuthService_VerifyEmail_FullMethodName:        true,
		authv1.AuthService_ResendVerification_FullMethodName: true,
		authv1.AuthService_SetPassword_FullMethodName:        true,
		authv1.AuthService_ResetPassword_FullMethodName:      true,
		authv1.AuthService_ForgotPassword_FullMethodName:     true,
		authv1.AuthService_RefreshToken_FullMethodName:       true,
	}
}

// auth-only:
//   logout, logout-all

func buildGRPCMethodPermissions() map[string]domain.PermissionCode {
	return map[string]domain.PermissionCode{
		// user self-service
		userv1.UserService_GetMyProfile_FullMethodName:     domain.PermProfileReadSelf,
		userv1.UserService_UpdateMyProfile_FullMethodName:  domain.PermProfileUpdateSelf,
		userv1.UserService_ChangeMyPassword_FullMethodName: domain.PermProfileChangePassword,

		// admin user management
		adminv1.AdminService_ListUsers_FullMethodName:                domain.PermUserList,
		adminv1.AdminService_CreateUser_FullMethodName:               domain.PermUserCreate,
		adminv1.AdminService_GetUserDetail_FullMethodName:            domain.PermUserReadDetail,
		adminv1.AdminService_UpdateUser_FullMethodName:               domain.PermUserUpdate,
		adminv1.AdminService_DeleteUser_FullMethodName:               domain.PermUserDelete,
		adminv1.AdminService_RestoreUser_FullMethodName:              domain.PermUserRestore,
		adminv1.AdminService_UpdateUserRole_FullMethodName:           domain.PermUserRoleUpdate,
		adminv1.AdminService_ListRoles_FullMethodName:                domain.PermRoleList,
		adminv1.AdminService_CreateRole_FullMethodName:               domain.PermRoleCreate,
		adminv1.AdminService_DeleteRole_FullMethodName:               domain.PermRoleDelete,
		adminv1.AdminService_AssignPermissionToRole_FullMethodName:   domain.PermRoleAssignPermission,
		adminv1.AdminService_RemovePermissionFromRole_FullMethodName: domain.PermRoleRemovePermission,
		adminv1.AdminService_ListPermissions_FullMethodName:          domain.PermPermissionList,
	}
}
