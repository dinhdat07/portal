package composer

import (
	"context"
	"fmt"
	"portal-system/config"
	"portal-system/internal/app"
	"portal-system/internal/infrastructure/logger"
	"portal-system/internal/infrastructure/ratelimit"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/infrastructure/storage"
	"portal-system/internal/infrastructure/validator"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Composer() (*app.App, error) {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	smtpCfg, err := config.LoadSMTPConfig()
	if err != nil {
		return nil, err
	}

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	rateLimitCfg, err := config.LoadRateLimitConfig()
	if err != nil {
		return nil, err
	}

	if rateLimitCfg.Enabled && !redisCfg.Enabled {
		return nil, fmt.Errorf("RATE_LIMIT_ENABLED=true requires REDIS_ENABLED=true")
	}

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	var rdb redis.UniversalClient
	if redisCfg.Enabled {
		rdb = redisx.NewClient(redisCfg)
		if err := redisx.Ping(context.Background(), rdb); err != nil {
			return nil, err
		}
	}

	var rateLimiter ratelimit.Limiter
	var rateLimitKeyBuilder ratelimit.KeyBuilder

	if rateLimitCfg.Enabled {
		rateLimiter, err = ratelimit.NewRedisLimiter(rdb)
		if err != nil {
			return nil, err
		}

		rateLimitKeyBuilder = ratelimit.NewKeyBuilder(rateLimitCfg.Prefix)
	}

	if err := storage.AutoMigrate(db); err != nil {
		return nil, err
	}

	if cfg.Env == "development" {
		if err := storage.SeedPermissions(db); err != nil {
			return nil, err
		}
		if err := storage.SeedRoles(db); err != nil {
			return nil, err
		}
		if err := storage.SeedRolePermissions(db); err != nil {
			return nil, err
		}
		if err := storage.SeedAdmin(db, cfg); err != nil {
			return nil, err
		}
	}

	validator := validator.NewValidator()
	infra := newInfra(cfg, smtpCfg, rdb)
	repos := newRepositories(db)
	svcs := newServices(cfg, infra, repos)
	grpcServers := newGRPCServers(svcs)

	return app.New(app.Deps{
		Config:              cfg,
		DB:                  db,
		Validator:           validator,
		Authenticator:       svcs.Authenticator,
		Authorizer:          svcs.Authorizer,
		AuthGRPC:            grpcServers.Auth,
		UserGRPC:            grpcServers.User,
		AdminGRPC:           grpcServers.Admin,
		RateLimiter:         rateLimiter,
		RateLimitKeyBuilder: rateLimitKeyBuilder,
		RateLimitConfig:     rateLimitCfg,
	}), nil
}
