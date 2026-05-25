package app

import (
	"context"
	"time"

	"portal-system/internal/observability/health"
	kafkax "portal-system/services/notification-service/internal/infrastructure/kafka"
	smtpx "portal-system/services/notification-service/internal/infrastructure/smtp"
)

func (a *App) readinessReport(ctx context.Context) health.Report {
	checks := map[string]string{
		"db":    health.StatusError,
		"kafka": health.StatusError,
		"smtp":  health.StatusError,
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

	if err := smtpx.Ping(checkCtx, a.SMTPHost, a.SMTPPort); err == nil {
		checks["smtp"] = health.StatusOK
	}

	return health.NewReport(checks)
}
