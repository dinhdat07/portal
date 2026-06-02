package app

import (
	"context"
	"fmt"
	"log/slog"
	"portal-system/config"
	"portal-system/internal/infrastructure/database"
	"portal-system/internal/infrastructure/logger"
	metricsx "portal-system/internal/infrastructure/metrics"
	"portal-system/internal/infrastructure/ratelimit"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/infrastructure/storage"
	"portal-system/internal/infrastructure/validator"
	"portal-system/internal/worker"

	kafkainfra "portal-system/internal/infrastructure/kafka"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
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

	announcementWorkerCfg, err := config.LoadAnnouncementWorkerConfig()
	if err != nil {
		return nil, err
	}

	if rateLimitCfg.Enabled && !redisCfg.Enabled {
		return nil, fmt.Errorf("RATE_LIMIT_ENABLED=true requires REDIS_ENABLED=true")
	}

	db, err := database.GetInstance(cfg.DBUrl)
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

	outboxPublisher := worker.NewOutboxPublisher(repos.TxManager, repos.OutboxRepo, infra.KafkaPublisher, slogLogger, outboxMetrics, worker.Config{
		Interval:            workerCfg.Interval,
		BatchSize:           workerCfg.BatchSize,
		MaxRetry:            workerCfg.MaxRetry,
		RetryInitialBackoff: workerCfg.RetryInitialBackoff,
		RetryMaxBackoff:     workerCfg.RetryMaxBackoff,
		RetryJitterRatio:    workerCfg.RetryJitterRatio,
	})

	announcementWorker := worker.NewAnnouncementWorker(
		repos.TxManager,
		repos.AnnouncementRepo,
		repos.UserRepo,
		repos.UserNotificationRepo,
		repos.OutboxRepo,
		slogLogger,
		worker.AnnouncementWorkerConfig{
			Interval:          announcementWorkerCfg.Interval,
			BatchSize:         announcementWorkerCfg.BatchSize,
			MaxUsersPerBatch:  announcementWorkerCfg.MaxUsersPerBatch,
			MaxRetry:          announcementWorkerCfg.MaxRetry,
			EventTTL:          announcementWorkerCfg.EventTTL,
			NotificationTopic: infra.NotificationTopic,
		},
	)

	return &App{
		Config:              cfg,
		DB:                  db,
		Validator:           validator,
		Authenticator:       svcs.Authenticator,
		Authorizer:          svcs.Authorizer,
		AuthGRPC:            grpcServers.Auth,
		UserGRPC:            grpcServers.User,
		AdminGRPC:           grpcServers.Admin,
		OutboxPublisher:     outboxPublisher,
		AnnouncementWorker:  announcementWorker,
		RateLimiter:         rateLimiter,
		RateLimitKeyBuilder: rateLimitKeyBuilder,
		RateLimitConfig:     rateLimitCfg,
		RedisClient:         rdb,
		KafkaBrokers:        kafkaCfg.Brokers,
	}, nil
}
