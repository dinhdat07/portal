package config

import (
	"time"
)

type WorkerConfig struct {
	FetchRetryInitialBackoff time.Duration
	FetchRetryMaxBackoff     time.Duration
	FetchRetryJitterRatio    float64
}

func LoadWorkerConfig() WorkerConfig {
	return WorkerConfig{
		FetchRetryInitialBackoff: time.Duration(getIntEnv("EMAIL_WORKER_FETCH_RETRY_INITIAL_BACKOFF_SECONDS", 1)) * time.Second,
		FetchRetryMaxBackoff:     time.Duration(getIntEnv("EMAIL_WORKER_FETCH_RETRY_MAX_BACKOFF_SECONDS", 30)) * time.Second,
		FetchRetryJitterRatio:    getFloatEnv("EMAIL_WORKER_FETCH_RETRY_JITTER_RATIO", 0.2),
	}
}
