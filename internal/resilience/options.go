package resilience

import "time"

// Option to configure the Backoff
type BackoffOption func(*BackoffConfig)

// WithInitialDelay configures the initial delay
func WithInitialDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		if d > 0 {
			c.InitialDelay = d
		}
	}
}

// WithMaxDelay configures the maximum wait ceiling
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		if d > 0 {
			c.MaxDelay = d
		}
	}
}

// WithMultiplier configures the maximum wait ceiling
func WithMultiplier(e float64) BackoffOption {
	return func(c *BackoffConfig) {
		if e > 0 {
			c.Multiplier = e
		}
	}
}
