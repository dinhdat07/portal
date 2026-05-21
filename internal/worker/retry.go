package outbox

import "time"

func nextRetryAt(retryCount int, cfg Config) time.Time {
	var delay time.Duration

	switch {
	case retryCount <= 1:
		delay = cfg.RetryDelay1
	case retryCount == 2:
		delay = cfg.RetryDelay2
	case retryCount == 3:
		delay = cfg.RetryDelay3
	default:
		delay = cfg.RetryDelayDefault
	}

	return time.Now().UTC().Add(delay)
}

