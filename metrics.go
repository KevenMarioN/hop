package hop

import (
	"github.com/KevenMarioN/hop/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsCollector defines the interface for metrics collection.
// Implementations can be Prometheus, OpenTelemetry, or any other system.
type MetricsCollector = metrics.MetricsCollector

// Counter represents a counter metric.
type Counter = metrics.Counter

// Gauge represents a gauge metric.
type Gauge = metrics.Gauge

// Histogram represents a histogram metric for measuring value distributions.
type Histogram = metrics.Histogram

// NewPrometheusCollector creates a Prometheus-backed metrics collector.
// It registers metrics with the provided prometheus.Registerer.
func NewPrometheusCollector(registry prometheus.Registerer) MetricsCollector {
	return metrics.NewPrometheusCollector(registry)
}

// NewOpenTelemetryCollector creates an OpenTelemetry-backed metrics collector.
// It uses the provided service name for metric identification.
func NewOpenTelemetryCollector(serviceName string) MetricsCollector {
	return metrics.NewOpenTelemetryCollector(serviceName)
}

// NewMultiCollector combines multiple collectors into one.
// Metrics are sent to all provided collectors.
func NewMultiCollector(collectors ...MetricsCollector) MetricsCollector {
	return metrics.NewMultiCollector(collectors...)
}

// NopCollector is a no-op collector that discards all metrics.
// Useful for testing or when metrics are not needed.
var NopCollector = metrics.NopCollector
