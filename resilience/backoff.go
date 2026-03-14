// resilience/backoff.go
package resilience

import (
	"context"
	"time"
)

// BackoffConfig define os parâmetros do algoritmo
type BackoffConfig struct {
	InitialDelay time.Duration // Ex: 100ms
	MaxDelay     time.Duration // Ex: 30s (o teto)
	Multiplier   float64       // Ex: 2.0 (dobrar o tempo)
}

// RetryOnce: Tenta conectar uma única vez (usado na inicialização).
// Retorna erro imediatamente se falhar.
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

// KeepTrying: Tenta indefinidamente até conseguir ou o contexto ser cancelado.
// Usa exponential backoff com reset.
func KeepTrying(ctx context.Context, fn func() error, opts ...BackoffOption) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	currentDelay := cfg.InitialDelay

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentDelay):
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
