package app

import (
	"net/http"
	"portal-system/internal/config"
	portalgrpc "portal-system/internal/grpc"
	"portal-system/internal/platform/ratelimit"
	auth "portal-system/internal/platform/security"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB

	GRPCServer *grpc.Server
	HTTPServer *http.Server

	Validator     protovalidate.Validator
	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer

	AuthGRPC  *portalgrpc.AuthServer
	UserGRPC  *portalgrpc.UserServer
	AdminGRPC *portalgrpc.AdminServer

	RateLimiter         ratelimit.Limiter
	RateLimitKeyBuilder ratelimit.KeyBuilder
	RateLimitConfig     *config.RateLimitConfig
}

type Deps struct {
	Config *config.Config
	DB     *gorm.DB

	GRPCServer *grpc.Server
	HTTPServer *http.Server

	Validator     protovalidate.Validator
	Authenticator *auth.Authenticator
	Authorizer    *auth.Authorizer

	AuthGRPC  *portalgrpc.AuthServer
	UserGRPC  *portalgrpc.UserServer
	AdminGRPC *portalgrpc.AdminServer

	RateLimiter         ratelimit.Limiter
	RateLimitKeyBuilder ratelimit.KeyBuilder
	RateLimitConfig     *config.RateLimitConfig
}

func New(deps Deps) *App {
	return &App{
		Config:              deps.Config,
		GRPCServer:          deps.GRPCServer,
		HTTPServer:          deps.HTTPServer,
		DB:                  deps.DB,
		Validator:           deps.Validator,
		Authenticator:       deps.Authenticator,
		Authorizer:          deps.Authorizer,
		AuthGRPC:            deps.AuthGRPC,
		UserGRPC:            deps.UserGRPC,
		AdminGRPC:           deps.AdminGRPC,
		RateLimiter:         deps.RateLimiter,
		RateLimitKeyBuilder: deps.RateLimitKeyBuilder,
		RateLimitConfig:     deps.RateLimitConfig,
	}
}
