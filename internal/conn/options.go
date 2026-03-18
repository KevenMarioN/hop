package conn

import (
	"crypto/tls"
	"time"

	"github.com/KevenMarioN/hop/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// HopOption configures a hop connection using the functional options pattern.
type HopOption func(*hop)

// WithConnectionName sets a custom name for the AMQP connection.
// This name appears in RabbitMQ management UI and logs.
// Useful for identifying connections in a multi-service environment.
func WithConnectionName(connectionName string) HopOption {
	return func(h *hop) {
		h.connectionName = connectionName
	}
}

// WithBackoff configures the reconnection backoff strategy.
// - multiplier: exponential factor (e.g., 2.0 doubles delay each attempt)
// - initialDelay: starting delay before first retry
// - maxDelay: maximum delay between retries (ceiling)
func WithBackoff(multiplier float64, initialDelay, maxDelay time.Duration) HopOption {
	return func(h *hop) {
		h.backoffConfig = backoffConfig{
			InitialDelay: initialDelay,
			MaxDelay:     maxDelay,
			Multiplier:   multiplier,
		}
	}
}

// WithTLS enables TLS encryption for the AMQP connection.
// Provide a configured *tls.Config for secure communication.
func WithTLS(tls *tls.Config) HopOption {
	return func(h *hop) {
		h.config.TLSClientConfig = tls
	}
}

// WithServiceName sets the service_name property on the AMQP connection.
// This metadata helps with monitoring and debugging in distributed systems.
func WithServiceName(serviceName string) HopOption {
	return func(h *hop) {
		h.config.Properties["service_name"] = serviceName
	}
}

// WithMetrics enables metrics collection using the provided collector.
// Pass a MetricsCollector implementation (e.g., PrometheusCollector, MultiCollector).
// If nil is passed, metrics will be disabled.
func WithMetrics(collector metrics.MetricsCollector) HopOption {
	return func(h *hop) {
		if collector != nil {
			h.collector = collector
		}
	}
}

// WithPrometheusMetrics is a convenience wrapper for backward compatibility.
// It creates a PrometheusCollector with the given registry.
func WithPrometheusMetrics(registry prometheus.Registerer) HopOption {
	return func(h *hop) {
		if registry != nil {
			h.collector = metrics.NewPrometheusCollector(registry)
		}
	}
}
