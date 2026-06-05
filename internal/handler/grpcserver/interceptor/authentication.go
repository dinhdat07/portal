package interceptor

import (
	"context"
	"errors"
	"portal-system/internal/handler/grpcserver/grpcctx"
	"portal-system/internal/infrastructure/security"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func AuthenticationInterceptor(authenticator *security.Authenticator, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authenticator == nil {
			return nil, status.Error(codes.Internal, "internal server error")
		}

		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		tokenString, err := extractBearerTokenFromMetadata(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		principal, err := authenticator.Authenticate(ctx, tokenString)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authorize format")
		}

		ctx = grpcctx.SetPrincipal(ctx, principal)
		return handler(ctx, req)

	}
}

func extractBearerTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	// Priority 1: Authorization header (legacy localStorage flow)
	if authValues := md.Get("authorization"); len(authValues) > 0 {
		authHeader := strings.TrimSpace(authValues[0])
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") && strings.TrimSpace(parts[1]) != "" {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	// Priority 2: Cookie-based access token (new HttpOnly flow)
	if cookieValues := md.Get("cookie-access-token"); len(cookieValues) > 0 {
		token := strings.TrimSpace(cookieValues[0])
		if token != "" {
			return token, nil
		}
	}

	return "", errors.New("missing token")
}
