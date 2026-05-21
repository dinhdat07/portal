package app

import (
	"context"
	"fmt"
	"log"
	"portal-system/services/notification-service/config"
	emailchannel "portal-system/services/notification-service/internal/channel/email"
	kafkax "portal-system/services/notification-service/internal/infrastructure/kafka"
	smtpx "portal-system/services/notification-service/internal/infrastructure/smtp"
	"portal-system/services/notification-service/internal/model"
	"portal-system/services/notification-service/internal/repository/impl"
	emailworker "portal-system/services/notification-service/internal/worker"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New() (*App, error) {
	log.Println("starting notification service bootstrap")

	cfg := config.Load()
	log.Printf(
		"notification service config loaded kafka_brokers=%v notification_topic=%s consumer_group=%s smtp_host=%s smtp_port=%s smtp_auth=%t smtp_tls=%t",
		cfg.Kafka.Brokers,
		cfg.Kafka.NotificationRequestedTopic,
		cfg.Kafka.ConsumerGroup,
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.UseAuth,
		cfg.SMTP.UseTLS,
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
	log.Println("kafka reader initialized")

	deliveryRepo := impl.NewGormDeliveryRepository(db)

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
	log.Println("smtp mailer initialized")
	emailRenderer := emailchannel.NewEmailRenderer()
	emailSender := emailchannel.NewSender(emailRenderer, smtpMailer)
	log.Println("email sender initialized")

	worker := emailworker.NewWorker(reader, emailSender, deliveryRepo,
		emailworker.Config{
			FetchRetryInitialBackoff: cfg.Worker.FetchRetryInitialBackoff,
			FetchRetryMaxBackoff:     cfg.Worker.FetchRetryMaxBackoff,
		},
	)
	log.Println("email notification worker initialized")

	return &App{
		EmailWorker: worker,
		KafkaReader: reader,
		DB:          db,
	}, nil
}
