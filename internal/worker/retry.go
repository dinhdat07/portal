package outbox

import (
	"math/rand/v2"
	"time"
)

func nextRetryAt(retryCount int, cfg Config) time.Time {
	initialBackoff := cfg.RetryInitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = 30 * time.Second
	}

	maxBackoff := cfg.RetryMaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Minute
	}

	delay := initialBackoff
	for i := 1; i < retryCount; i++ {
		delay *= 2
		if delay > maxBackoff {
			delay = maxBackoff
			break
		}
	}

	delay = applyJitter(delay, cfg.RetryJitterRatio)

	return time.Now().UTC().Add(delay)
}

func applyJitter(delay time.Duration, ratio float64) time.Duration {
	if ratio <= 0 {
		return delay
	}

	if ratio > 1 {
		ratio = 1
	}

	delta := float64(delay) * ratio
	minDelay := float64(delay) - delta
	maxDelay := float64(delay) + delta

	jittered := minDelay + rand.Float64()*(maxDelay-minDelay)
	if jittered < float64(time.Millisecond) {
		return time.Millisecond
	}

	return time.Duration(jittered)
}
