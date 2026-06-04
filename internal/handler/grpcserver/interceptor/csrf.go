package interceptor

import (
	"context"
	"strings"

	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"
	"portal-system/internal/infrastructure/security"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// stateChangingMethods are gRPC methods that require CSRF validation.
func stateChangingMethods() map[string]bool {
	return map[string]bool{
		// Auth
		authv1.AuthService_Login_FullMethodName:     true,
		authv1.AuthService_Logout_FullMethodName:    true,
		authv1.AuthService_LogoutAll_FullMethodName: true,
		// Users
		userv1.UserService_UpdateMyProfile_FullMethodName:  true,
		userv1.UserService_ChangeMyPassword_FullMethodName: true,
		// Admin
		adminv1.AdminService_CreateUser_FullMethodName:               true,
		adminv1.AdminService_UpdateUser_FullMethodName:               true,
		adminv1.AdminService_DeleteUser_FullMethodName:               true,
		adminv1.AdminService_RestoreUser_FullMethodName:              true,
		adminv1.AdminService_UpdateUserRole_FullMethodName:           true,
		adminv1.AdminService_CreateRole_FullMethodName:               true,
		adminv1.AdminService_DeleteRole_FullMethodName:               true,
		adminv1.AdminService_AssignPermissionToRole_FullMethodName:   true,
		adminv1.AdminService_RemovePermissionFromRole_FullMethodName: true,
	}
}

func CSRFInterceptor(csrfManager *security.CSRFManager) grpc.UnaryServerInterceptor {
	protectedMethods := stateChangingMethods()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !protectedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		cookieValues := md.Get("cookie-csrf-token")
		headerValues := md.Get("x-csrf-token")

		cookieToken := ""
		headerToken := ""

		if len(cookieValues) > 0 {
			cookieToken = strings.TrimSpace(cookieValues[0])
		}
		if len(headerValues) > 0 {
			headerToken = strings.TrimSpace(headerValues[0])
		}

		// If neither is present, client may not have cookies yet (e.g., login).
		if cookieToken == "" && headerToken == "" {
			return handler(ctx, req)
		}

		if err := csrfManager.ValidateCSRFToken(cookieToken, headerToken); err != nil {
			return nil, status.Error(codes.PermissionDenied, "CSRF validation failed: "+err.Error())
		}

		return handler(ctx, req)
	}
}
