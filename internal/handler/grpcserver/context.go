package grpcserver

import (
	"context"
	"errors"
	"net"
	"portal-system/internal/handler/grpcserver/grpcctx"
	"portal-system/internal/service"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	gstatus "google.golang.org/grpc/status"
)

type contextKey string

const AuditUserContextKey contextKey = "audit_user"
const SessionIDContextKey contextKey = "session_id"

func getAuditFromCtx(ctx context.Context) *service.AuditMeta {
	meta := &service.AuditMeta{}

	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		host, _, err := net.SplitHostPort(p.Addr.String())
		if err == nil {
			meta.IPAddress = host
		} else {
			meta.IPAddress = p.Addr.String()
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ua := md.Get("user-agent")
		if len(ua) > 0 {
			meta.UserAgent = ua[0]
		}
	}

	return meta
}

func getActorFromCtx(ctx context.Context) (*service.AuditUser, error) {
	principal, exists := grpcctx.GetPrincipal(ctx)
	if principal == nil || !exists {
		return nil, errors.New("missing principal in context")
	}

	return &service.AuditUser{
		ID:       principal.UserID,
		Username: principal.Username,
		Email:    principal.Email,
		RoleCode: principal.RoleCode,
	}, nil
}

func getSessionIDFromCtx(ctx context.Context) (uuid.UUID, error) {
	principal, exists := grpcctx.GetPrincipal(ctx)
	if principal == nil || !exists {
		return uuid.Nil, errors.New("missing principal in context")
	}

	if principal.SessionID == uuid.Nil {
		return uuid.Nil, gstatus.Error(codes.Unauthenticated, "missing session id")
	}

	return principal.SessionID, nil
}
