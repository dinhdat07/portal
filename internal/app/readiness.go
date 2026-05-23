package app

import (
	"context"
	"time"

	kafkax "portal-system/internal/infrastructure/kafka"
	redisx "portal-system/internal/infrastructure/redis"
	"portal-system/internal/observability/health"
)

func (a *App) readinessReport(ctx context.Context) health.Report {
	checks := map[string]string{
		"db":    health.StatusError,
		"kafka": health.StatusError,
	}

	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if a.DB != nil {
		if sqlDB, err := a.DB.DB(); err == nil && sqlDB.PingContext(checkCtx) == nil {
			checks["db"] = health.StatusOK
		}
	}

	if err := kafkax.Ping(checkCtx, a.KafkaBrokers); err == nil {
		checks["kafka"] = health.StatusOK
	}

	if a.RateLimitConfig != nil && a.RateLimitConfig.Enabled {
		checks["redis"] = health.StatusError
		if err := redisx.Ping(checkCtx, a.RedisClient); err == nil {
			checks["redis"] = health.StatusOK
		}
	}

	return health.NewReport(checks)
}
