package interceptor

import (
	"context"
	"portal-system/internal/domain"
	"portal-system/internal/handler/grpcserver/grpcctx"
	"portal-system/internal/infrastructure/security"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func PermissionInterceptor(authorizer *security.Authorizer, methodPermissions map[string]domain.PermissionCode) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authorizer == nil {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		requiredPerm, hasRule := methodPermissions[info.FullMethod]
		if !hasRule {
			return handler(ctx, req)
		}

		principal, ok := grpcctx.GetPrincipal(ctx)
		if !ok || principal == nil {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		allowed := authorizer.HasPermission(ctx, principal, requiredPerm)
		if !allowed {
			return nil, status.Error(codes.PermissionDenied, "forbidden")
		}

		ctx = grpcctx.SetPrincipal(ctx, principal)
		return handler(ctx, req)

	}
}
