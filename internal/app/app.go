package app

import (
	"net/http"
	"portal-system/config"
	"portal-system/internal/handler"
	"portal-system/internal/infrastructure/ratelimit"
	"portal-system/internal/infrastructure/security"

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
	Authenticator *security.Authenticator
	Authorizer    *security.Authorizer

	AuthGRPC  *handler.AuthServer
	UserGRPC  *handler.UserServer
	AdminGRPC *handler.AdminServer

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
	Authenticator *security.Authenticator
	Authorizer    *security.Authorizer

	AuthGRPC  *handler.AuthServer
	UserGRPC  *handler.UserServer
	AdminGRPC *handler.AdminServer

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
