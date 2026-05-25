package app

import "portal-system/internal/handler/grpcserver"

type GRPCServers struct {
	Auth  *grpcserver.AuthServer
	User  *grpcserver.UserServer
	Admin *grpcserver.AdminServer
}

func newGRPCServers(svcs *Services) *GRPCServers {
	return &GRPCServers{
		Auth:  grpcserver.NewAuthServer(svcs.Auth),
		User:  grpcserver.NewUserServer(svcs.User),
		Admin: grpcserver.NewAdminServer(svcs.Admin, svcs.User, svcs.Role, svcs.Permission),
	}
}
