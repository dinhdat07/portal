package config

import (
	"fmt"
	"time"
)

type OutboxWorkerConfig struct {
	Interval          time.Duration
	BatchSize         int
	MaxRetry          int
	RetryDelay1       time.Duration
	RetryDelay2       time.Duration
	RetryDelay3       time.Duration
	RetryDelayDefault time.Duration
}

func LoadOutboxWorkerConfig() (*OutboxWorkerConfig, error) {
	loadEnv()

	interval, err := getEnvDuration("OUTBOX_WORKER_INTERVAL", 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_INTERVAL: %w", err)
	}

	batchSize, err := getEnvInt("OUTBOX_WORKER_BATCH_SIZE", 50)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_BATCH_SIZE: %w", err)
	}

	maxRetry, err := getEnvInt("OUTBOX_WORKER_MAX_RETRY", 10)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_MAX_RETRY: %w", err)
	}

	delay1, err := getEnvDuration("OUTBOX_WORKER_RETRY_DELAY_1", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_DELAY_1: %w", err)
	}

	delay2, err := getEnvDuration("OUTBOX_WORKER_RETRY_DELAY_2", 2*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_DELAY_2: %w", err)
	}

	delay3, err := getEnvDuration("OUTBOX_WORKER_RETRY_DELAY_3", 10*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_DELAY_3: %w", err)
	}

	delayDefault, err := getEnvDuration("OUTBOX_WORKER_RETRY_DELAY_DEFAULT", 30*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_DELAY_DEFAULT: %w", err)
	}

	return &OutboxWorkerConfig{
		Interval:          interval,
		BatchSize:         batchSize,
		MaxRetry:          maxRetry,
		RetryDelay1:       delay1,
		RetryDelay2:       delay2,
		RetryDelay3:       delay3,
		RetryDelayDefault: delayDefault,
	}, nil
}
