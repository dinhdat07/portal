package app

import (
	"net/http"
	"portal-system/config"
	"portal-system/internal/handler/grpcserver"
	"portal-system/internal/infrastructure/ratelimit"
	"portal-system/internal/infrastructure/security"
	"portal-system/internal/service"
	"portal-system/internal/worker"

	"buf.build/go/protovalidate"
	"github.com/redis/go-redis/v9"
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
	CSRFManager   *security.CSRFManager

	AuthGRPC  *grpcserver.AuthServer
	UserGRPC  *grpcserver.UserServer
	AdminGRPC *grpcserver.AdminServer

	UserService service.UserService

	OutboxPublisher    *worker.OutboxPublisher
	AnnouncementWorker *worker.AnnouncementWorker

	RateLimiter         ratelimit.Limiter
	RateLimitKeyBuilder ratelimit.KeyBuilder
	RateLimitConfig     *config.RateLimitConfig
	RedisClient         redis.UniversalClient
	KafkaBrokers        []string
}
