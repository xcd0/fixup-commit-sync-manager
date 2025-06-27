package retry

import (
	"fmt"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	Delay      time.Duration
}

type Operation func() error

func WithRetry(operation Operation, config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt < config.MaxRetries {
			time.Sleep(config.Delay)
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

func NewRetryConfig(maxRetries int, delay time.Duration) RetryConfig {
	return RetryConfig{
		MaxRetries: maxRetries,
		Delay:      delay,
	}
}
