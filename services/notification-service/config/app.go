package config

type Config struct {
	Kafka  KafkaConfig
	SMTP   SMTPConfig
	Worker WorkerConfig
	DBUrl  string
}

func Load() Config {
	return Config{
		Kafka:  LoadKafkaConfig(),
		SMTP:   LoadSMTPConfig(),
		Worker: LoadWorkerConfig(),
		DBUrl:  getEnv("DB_URL", ""),
	}
}

