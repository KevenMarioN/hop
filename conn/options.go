package conn

import (
	"crypto/tls"
	"time"
)

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

func WithTLS(tls *tls.Config) HopOption {
	return func(h *hop) {
		h.config.TLSClientConfig = tls
	}
}

func WithServiceName(serviceName string) HopOption {
	return func(h *hop) {
		h.config.Properties["service_name"] = serviceName
	}
}
