package app

import "portal-system/internal/handler"

type GRPCServers struct {
	Auth  *handler.AuthServer
	User  *handler.UserServer
	Admin *handler.AdminServer
}

func newGRPCServers(svcs *Services) *GRPCServers {
	return &GRPCServers{
		Auth:  handler.NewAuthServer(svcs.Auth),
		User:  handler.NewUserServer(svcs.User),
		Admin: handler.NewAdminServer(svcs.Admin, svcs.User, svcs.Role, svcs.Permission),
	}
}
