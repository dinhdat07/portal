package config

import (
	"fmt"
	"time"
)

type OutboxWorkerConfig struct {
	Interval            time.Duration
	BatchSize           int
	MaxRetry            int
	RetryInitialBackoff time.Duration
	RetryMaxBackoff     time.Duration
	RetryJitterRatio    float64
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

	initialBackoff, err := getEnvDuration("OUTBOX_WORKER_RETRY_INITIAL_BACKOFF", 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_INITIAL_BACKOFF: %w", err)
	}

	maxBackoff, err := getEnvDuration("OUTBOX_WORKER_RETRY_MAX_BACKOFF", 30*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_MAX_BACKOFF: %w", err)
	}

	jitterRatio, err := getEnvFloat("OUTBOX_WORKER_RETRY_JITTER_RATIO", 0.2)
	if err != nil {
		return nil, fmt.Errorf("invalid OUTBOX_WORKER_RETRY_JITTER_RATIO: %w", err)
	}

	return &OutboxWorkerConfig{
		Interval:            interval,
		BatchSize:           batchSize,
		MaxRetry:            maxRetry,
		RetryInitialBackoff: initialBackoff,
		RetryMaxBackoff:     maxBackoff,
		RetryJitterRatio:    jitterRatio,
	}, nil
}
