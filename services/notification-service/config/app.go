package config

type Config struct {
	Kafka KafkaConfig
	SMTP  SMTPConfig
}

func Load() Config {
	return Config{
		Kafka: LoadKafkaConfig(),
		SMTP:  LoadSMTPConfig(),
	}
}
