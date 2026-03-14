package resilience

import "time"

// Option para configurar o Backoff
type BackoffOption func(*BackoffConfig)

// WithInitialDelay configura o delay inicial
func WithInitialDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		c.InitialDelay = d
	}
}

// WithMaxDelay configura o teto máximo de espera
func WithMaxDelay(d time.Duration) BackoffOption {
	return func(c *BackoffConfig) {
		c.MaxDelay = d
	}
}
