package app

import (
	"context"
	"fmt"
	"log/slog"
	"portal-system/config"
	"portal-system/internal/infrastructure/logger"
	metricsx "portal-system/internal/infrastructure/metrics"
	"portal-system/internal/infrastructure/ratelimit"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/infrastructure/storage"
	"portal-system/internal/infrastructure/validator"
	outbox "portal-system/internal/worker"

	kafkainfra "portal-system/internal/infrastructure/kafka"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New() (*App, error) {
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	slogLogger := logger.New(logger.Config{
		Env:    cfg.Env,
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
	slog.SetDefault(slogLogger)

	redisCfg, err := config.LoadRedisConfig()
	if err != nil {
		return nil, err
	}

	kafkaCfg, err := config.LoadKafkaConfig()
	if err != nil {
		return nil, err
	}

	rateLimitCfg, err := config.LoadRateLimitConfig()
	if err != nil {
		return nil, err
	}

	workerCfg, err := config.LoadOutboxWorkerConfig()
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

	writer := kafkainfra.NewWriter(kafkaCfg.Brokers)

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
	infra := newInfra(cfg, writer, kafkaCfg, rdb)
	repos := newRepositories(db)
	svcs := newServices(cfg, infra, repos)
	grpcServers := newGRPCServers(svcs)

	outboxMetrics := metricsx.NewPrometheusOutboxMetrics(prometheus.DefaultRegisterer)

	outboxWorker := outbox.NewWorker(repos.TxManager, repos.OutboxRepo, infra.KafkaPublisher, slogLogger, outboxMetrics, outbox.Config{
		Interval:            workerCfg.Interval,
		BatchSize:           workerCfg.BatchSize,
		MaxRetry:            workerCfg.MaxRetry,
		RetryInitialBackoff: workerCfg.RetryInitialBackoff,
		RetryMaxBackoff:     workerCfg.RetryMaxBackoff,
		RetryJitterRatio:    workerCfg.RetryJitterRatio,
	})

	return &App{
		Config:              cfg,
		DB:                  db,
		Validator:           validator,
		Authenticator:       svcs.Authenticator,
		Authorizer:          svcs.Authorizer,
		AuthGRPC:            grpcServers.Auth,
		UserGRPC:            grpcServers.User,
		AdminGRPC:           grpcServers.Admin,
		OutboxWorker:        outboxWorker,
		RateLimiter:         rateLimiter,
		RateLimitKeyBuilder: rateLimitKeyBuilder,
		RateLimitConfig:     rateLimitCfg,
		RedisClient:         rdb,
		KafkaBrokers:        kafkaCfg.Brokers,
	}, nil
}
