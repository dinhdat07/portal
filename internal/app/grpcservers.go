package app

import (
	"portal-system/internal/handler/grpcserver"
	"portal-system/internal/infrastructure/security"
)

type GRPCServers struct {
	Auth  *grpcserver.AuthServer
	User  *grpcserver.UserServer
	Admin *grpcserver.AdminServer
}

func newGRPCServers(svcs *Services, csrfManager *security.CSRFManager) *GRPCServers {
	return &GRPCServers{
		Auth:  grpcserver.NewAuthServer(svcs.Auth, csrfManager),
		User:  grpcserver.NewUserServer(svcs.User, svcs.UserNotification),
		Admin: grpcserver.NewAdminServer(svcs.Admin, svcs.User, svcs.Role, svcs.Permission, svcs.Announcement),
	}
}
