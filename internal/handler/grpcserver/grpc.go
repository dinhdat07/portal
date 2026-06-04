package grpcserver

import (
	"portal-system/config"
	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"
	"portal-system/internal/domain"
	"portal-system/internal/handler/grpcserver/interceptor"
	"portal-system/internal/infrastructure/ratelimit"
	"portal-system/internal/infrastructure/security"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
)

type GRPCServerDeps struct {
	Validator     protovalidate.Validator
	Authenticator *security.Authenticator
	Authorizer    *security.Authorizer
	CSRFManager   *security.CSRFManager

	Auth  *AuthServer
	User  *UserServer
	Admin *AdminServer

	RateLimiter         ratelimit.Limiter
	RateLimitKeyBuilder ratelimit.KeyBuilder
	RateLimitConfig     *config.RateLimitConfig
}

func NewGRPCServer(deps GRPCServerDeps) *grpc.Server {
	publicMethods := buildGRPCPublicMethods()
	methodPermissions := buildGRPCMethodPermissions()

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.PreAuthRateLimitInterceptor(deps.RateLimiter, deps.RateLimitKeyBuilder, deps.RateLimitConfig),
			interceptor.ValidationInterceptor(deps.Validator),
			interceptor.CSRFInterceptor(deps.CSRFManager),
			interceptor.AuthenticationInterceptor(deps.Authenticator, publicMethods),
			interceptor.PostAuthRateLimitInterceptor(deps.RateLimiter, deps.RateLimitKeyBuilder, deps.RateLimitConfig),
			interceptor.PermissionInterceptor(deps.Authorizer, methodPermissions),
		),
	)

	authv1.RegisterAuthServiceServer(s, deps.Auth)
	adminv1.RegisterAdminServiceServer(s, deps.Admin)
	userv1.RegisterUserServiceServer(s, deps.User)

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

func buildGRPCMethodPermissions() map[string]domain.PermissionCode {
	return map[string]domain.PermissionCode{
		// user self-service
		userv1.UserService_GetMyProfile_FullMethodName:               domain.PermProfileReadSelf,
		userv1.UserService_UpdateMyProfile_FullMethodName:            domain.PermProfileUpdateSelf,
		userv1.UserService_ChangeMyPassword_FullMethodName:           domain.PermProfileChangePassword,
		userv1.UserService_ListMyNotifications_FullMethodName:        domain.PermNotificationReadSelf,
		userv1.UserService_GetMyNotificationDetail_FullMethodName:    domain.PermNotificationReadSelf,
		userv1.UserService_MarkNotificationAsRead_FullMethodName:     domain.PermNotificationReadSelf,
		userv1.UserService_MarkAllNotificationsAsRead_FullMethodName: domain.PermNotificationReadSelf,
		userv1.UserService_GetUnreadNotificationCount_FullMethodName: domain.PermNotificationReadSelf,

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

		adminv1.AdminService_CreateAnnouncement_FullMethodName:    domain.PermAnnouncementCreate,
		adminv1.AdminService_ListAnnouncements_FullMethodName:     domain.PermAnnouncementList,
		adminv1.AdminService_GetAnnouncementDetail_FullMethodName: domain.PermAnnouncementReadDetail,
	}
}
