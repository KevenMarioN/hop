// resilience/backoff.go
package resilience

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

// BackoffConfig defines the parameters for the backoff algorithm
type BackoffConfig struct {
	InitialDelay time.Duration // Ex: 100ms
	MaxDelay     time.Duration // Ex: 30s (ceiling)
	Multiplier   float64       // Ex: 2.0 (double the time)
}

// RetryOnce: Tries to connect once (used during initialization).
// Returns error immediately if it fails.
func RetryOnce(fn func() error) error {
	return fn()
}

func defaultConfig() BackoffConfig {
	return BackoffConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// KeepTrying: Tries indefinitely until successful or context is cancelled.
// Uses exponential backoff with reset.
func KeepTrying(ctx context.Context, fn func() error, opts ...BackoffOption) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	currentDelay := cfg.InitialDelay
	attemptCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentDelay):
			attemptCount++
			log.Warn().Msgf("Retrying operation (attempt %d) after %v delay", attemptCount, currentDelay)

			if err := fn(); err == nil {
				return nil
			}

			nextDelay := time.Duration(float64(currentDelay) * cfg.Multiplier)
			if nextDelay >= cfg.MaxDelay {
				currentDelay = cfg.InitialDelay
			} else {
				currentDelay = nextDelay
			}
		}
	}
}
