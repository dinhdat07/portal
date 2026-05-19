package app

import (
	"portal-system/services/notification-service/config"
	emailchannel "portal-system/services/notification-service/internal/channel/email"
	kafkax "portal-system/services/notification-service/internal/infrastructure/kafka"
	smtpx "portal-system/services/notification-service/internal/infrastructure/smtp"
	emailworker "portal-system/services/notification-service/internal/worker"
)

func New() (*App, error) {
	cfg := config.Load()

	reader := kafkax.NewReader(
		cfg.Kafka.Brokers,
		[]string{cfg.Kafka.NotificationRequestedTopic},
		cfg.Kafka.ConsumerGroup,
	)

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

	emailRenderer := emailchannel.NewEmailRenderer()
	emailSender := emailchannel.NewSender(emailRenderer, smtpMailer)

	worker := emailworker.NewWorker(reader, emailSender)

	return &App{
		EmailWorker: worker,
		KafkaReader: reader,
	}, nil
}
