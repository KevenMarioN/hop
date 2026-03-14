package resilience

import "time"

// Option para configurar o Backoff
type BackoffOption func(*BackoffConfig)

// WithInitialDelay configura o delay inicial
func WithInitialDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		if d > 0 {
			c.InitialDelay = d
		}
	}
}

// WithMaxDelay configura o teto máximo de espera
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		if d > 0 {
			c.MaxDelay = d
		}
	}
}

// WithMultiplier configura o teto máximo de espera
func WithMultiplier(e float64) BackoffOption {
	return func(c *BackoffConfig) {
		if e > 0 {
			c.Multiplier = e
		}
	}
}
