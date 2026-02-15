package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// IsTransientError checks if an error is a transient database error that should be retried
func IsTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Don't retry on non-transient errors
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, sql.ErrTxDone) {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Retry on these common transient error patterns
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"timeout",
		"deadline exceeded",
		"too many connections",
		"connection pool exhausted",
		"could not serialize",
		"temporary failure",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryConfig returns sensible defaults for database retries
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 50 * time.Millisecond,
		MaxWait:     500 * time.Millisecond,
	}
}

// WithRetry wraps a database operation with retry logic for transient errors
func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Execute the operation
		err := operation()

		// Success - return immediately
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on non-transient errors
		if !IsTransientError(err) {
			return err
		}

		// Don't retry if context is cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Don't wait after the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		// Calculate exponential backoff: initialWait * 2^(attempt-1)
		waitDuration := config.InitialWait * time.Duration(1<<uint(attempt-1))
		if waitDuration > config.MaxWait {
			waitDuration = config.MaxWait
		}

		// Wait before retrying
		select {
		case <-time.After(waitDuration):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}
