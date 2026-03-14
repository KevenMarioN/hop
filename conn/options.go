package conn

import "time"

type HopOption func(*hop)

func WithConnectionName(connectionName string) HopOption {
	return func(h *hop) {
		h.connectionName = connectionName
	}
}

func WithBackoff(multiplier float64, initialDelay, maxDelay time.Duration) HopOption {
	return func(h *hop) {
		h.backoffConfig = backoffConfig{
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			Multiplier:   multiplier,
		}
	}
}
