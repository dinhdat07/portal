package outbox

import "time"

func nextRetryAt(retryCount int) time.Time {
	var delay time.Duration

	switch {
	case retryCount <= 1:
		delay = 30 * time.Second
	case retryCount == 2:
		delay = 2 * time.Minute
	case retryCount == 3:
		delay = 10 * time.Minute
	default:
		delay = 30 * time.Minute
	}

	return time.Now().UTC().Add(delay)
}
