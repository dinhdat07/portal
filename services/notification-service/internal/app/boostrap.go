package app

import (
	"context"
	"fmt"
	"log/slog"
	"portal-system/services/notification-service/config"
	emailchannel "portal-system/services/notification-service/internal/channel/email"
	kafkax "portal-system/services/notification-service/internal/infrastructure/kafka"
	logger "portal-system/services/notification-service/internal/infrastructure/logger"
	metricsx "portal-system/services/notification-service/internal/infrastructure/metrics"
	smtpx "portal-system/services/notification-service/internal/infrastructure/smtp"
	"portal-system/services/notification-service/internal/model"
	"portal-system/services/notification-service/internal/repository/impl"
	emailworker "portal-system/services/notification-service/internal/worker"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New() (*App, error) {
	cfg := config.Load()

	slogLogger := logger.New(logger.Config{
		Env:    cfg.Logger.Env,
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
	slog.SetDefault(slogLogger)

	slogLogger.Info("starting_notification_service_bootstrap")

	slogLogger.Info("notification_service_config_loaded",
		slog.Any("kafka_brokers", cfg.Kafka.Brokers),
		slog.String("notification_topic", cfg.Kafka.NotificationRequestedTopic),
		slog.String("consumer_group", cfg.Kafka.ConsumerGroup),
		slog.String("smtp_host", cfg.SMTP.Host),
		slog.String("smtp_port", cfg.SMTP.Port),
		slog.Bool("smtp_use_auth", cfg.SMTP.UseAuth),
		slog.Bool("smtp_use_tls", cfg.SMTP.UseTLS),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := kafkax.Ping(ctx, cfg.Kafka.Brokers); err != nil {
		return nil, fmt.Errorf("verify kafka connection: %w", err)
	}

	if err := smtpx.Ping(ctx, cfg.SMTP.Host, cfg.SMTP.Port); err != nil {
		return nil, fmt.Errorf("verify smtp connection: %w", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.NotificationDelivery{}); err != nil {
		return nil, fmt.Errorf("auto migrate notification deliveries: %w", err)
	}

	reader := kafkax.NewReader(
		cfg.Kafka.Brokers,
		[]string{cfg.Kafka.NotificationRequestedTopic},
		cfg.Kafka.ConsumerGroup,
	)
	slogLogger.Info("kafka_reader_initialized")

	deliveryRepo := impl.NewGormDeliveryRepository(db)
	txManager := impl.NewGormTxManager(db)

	smtpMailer := smtpx.NewMailer(smtpx.Config{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		UseAuth:  cfg.SMTP.UseAuth,
		UseTLS:   cfg.SMTP.UseTLS,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
		FromName: cfg.SMTP.FromName,
	})

	emailMetrics, retryMetrics := metricsx.NewPrometheusMetrics(prometheus.DefaultRegisterer)

	slogLogger.Info("smtp_mailer_initialized")
	emailRenderer := emailchannel.NewEmailRenderer()
	emailSender := emailchannel.NewSender(emailRenderer, smtpMailer)
	slogLogger.Info("email_sender_initialized")

	worker := emailworker.NewWorker(reader, emailSender, txManager, deliveryRepo, slogLogger, emailMetrics,
		emailworker.Config{
			FetchRetryInitialBackoff:    cfg.Worker.FetchRetryInitialBackoff,
			FetchRetryMaxBackoff:        cfg.Worker.FetchRetryMaxBackoff,
			MaxRetry:                    cfg.Worker.MaxRetry,
			DeliveryRetryInitialBackoff: cfg.Worker.DeliveryRetryInitialBackoff,
			DeliveryRetryMaxBackoff:     cfg.Worker.DeliveryRetryMaxBackoff,
			DeliveryRetryJitterRatio:    cfg.Worker.DeliveryRetryJitterRatio,
		},
	)
	slogLogger.Info("email_notification_worker_initialized")

	retryWorker := emailworker.NewRetryWorker(emailSender, deliveryRepo, slogLogger, retryMetrics,
		emailworker.RetryWorkerConfig{
			Interval:                    cfg.Worker.RetryWorkerInterval,
			BatchSize:                   cfg.Worker.RetryWorkerBatchSize,
			MaxRetry:                    cfg.Worker.MaxRetry,
			DeliveryRetryInitialBackoff: cfg.Worker.DeliveryRetryInitialBackoff,
			DeliveryRetryMaxBackoff:     cfg.Worker.DeliveryRetryMaxBackoff,
			DeliveryRetryJitterRatio:    cfg.Worker.DeliveryRetryJitterRatio,
		},
	)
	slogLogger.Info("email_notification_retry_worker_initialized")

	return &App{
		EmailWorker:  worker,
		RetryWorker:  retryWorker,
		KafkaReader:  reader,
		DB:           db,
		MetricsPort:  cfg.MetricsPort,
		KafkaBrokers: cfg.Kafka.Brokers,
		SMTPHost:     cfg.SMTP.Host,
		SMTPPort:     cfg.SMTP.Port,
	}, nil
}
