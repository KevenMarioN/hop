package metrics

// MetricsCollector defines the interface for metrics collection.
// Implementations can be Prometheus, OpenTelemetry, or any other system.
type MetricsCollector interface {
	// Counter returns a Counter metric with the provided labels.
	Counter(name string, labels ...string) Counter
	// Gauge returns a Gauge metric with the provided labels.
	Gauge(name string, labels ...string) Gauge
	// Histogram returns a Histogram metric with the provided labels.
	Histogram(name string, labels ...string) Histogram
	// Registerer returns the underlying registerer for metric registration.
	// May be nil if the collector doesn't support explicit registration.
	Registerer() any
}

// Counter represents a counter metric.
type Counter interface {
	Inc()
	Add(float64)
}

// Gauge represents a gauge metric.
type Gauge interface {
	Set(float64)
	Inc()
	Dec()
	Add(float64)
}

// Histogram represents a histogram metric for measuring value distributions.
type Histogram interface {
	Observe(value float64)
}
