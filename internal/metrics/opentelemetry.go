package metrics

import (
	"context"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// OpenTelemetryCollector implements MetricsCollector using OpenTelemetry.
type OpenTelemetryCollector struct {
	meter       metric.Meter
	counters    map[string]metric.Int64Counter
	gauges      map[string]metric.Int64ObservableGauge
	gaugeValues map[string]int64
	gaugeMutex  sync.RWMutex
	histograms  map[string]metric.Float64Histogram
}

// NewOpenTelemetryCollector creates a new OpenTelemetry collector.
func NewOpenTelemetryCollector(meterName string) *OpenTelemetryCollector {
	meter := otel.Meter(meterName)

	oc := &OpenTelemetryCollector{
		meter:       meter,
		counters:    make(map[string]metric.Int64Counter),
		gauges:      make(map[string]metric.Int64ObservableGauge),
		gaugeValues: make(map[string]int64),
		histograms:  make(map[string]metric.Float64Histogram),
	}

	// Register default Hop metrics
	oc.registerDefaultMetrics()

	return oc
}

func (oc *OpenTelemetryCollector) Counter(name string, labels ...string) Counter {
	key := name + "|" + strings.Join(labels, "|")

	if c, ok := oc.counters[key]; ok {
		return &otelCounter{counter: c, labels: labels}
	}

	// Create new counter
	counter, err := oc.meter.Int64Counter(
		name,
		metric.WithDescription(name),
	)
	if err != nil {
		// In case of error, return no-op counter
		return &otelCounter{}
	}

	oc.counters[key] = counter

	return &otelCounter{counter: counter, labels: labels}
}

func (oc *OpenTelemetryCollector) Gauge(name string, labels ...string) Gauge {
	key := name + "|" + strings.Join(labels, "|")

	if g, ok := oc.gauges[key]; ok {
		return &otelGauge{gauge: g, collector: oc, key: key, labels: labels}
	}

	// Create new observable gauge
	gauge, err := oc.meter.Int64ObservableGauge(
		name,
		metric.WithDescription(name),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			oc.gaugeMutex.RLock()
			defer oc.gaugeMutex.RUnlock()

			if val, ok := oc.gaugeValues[key]; ok {
				o.Observe(val)
			}

			return nil
		}),
	)
	if err != nil {
		// In case of error, return no-op gauge
		return &otelGauge{}
	}

	oc.gauges[key] = gauge

	return &otelGauge{gauge: gauge, collector: oc, key: key, labels: labels}
}

func (oc *OpenTelemetryCollector) Histogram(name string, labels ...string) Histogram {
	key := name + "|" + strings.Join(labels, "|")

	if h, ok := oc.histograms[key]; ok {
		return &otelHistogram{histogram: h}
	}

	// Create new histogram
	histogram, err := oc.meter.Float64Histogram(
		name,
		metric.WithDescription(name),
	)
	if err != nil {
		// In case of error, return no-op histogram
		return &otelHistogram{}
	}

	oc.histograms[key] = histogram

	return &otelHistogram{histogram: histogram}
}

func (oc *OpenTelemetryCollector) Registerer() any {
	return nil // OpenTelemetry doesn't use registerer
}

// Wrapper implementations
type otelCounter struct {
	counter metric.Int64Counter
	labels  []string
}

func (oc *otelCounter) Inc() {
	if oc.counter != nil {
		oc.counter.Add(context.Background(), 1)
	}
}

func (oc *otelCounter) Add(v float64) {
	if oc.counter != nil {
		oc.counter.Add(context.Background(), int64(v))
	}
}

type otelGauge struct {
	gauge     metric.Int64ObservableGauge
	collector *OpenTelemetryCollector
	key       string
	labels    []string
}

func (og *otelGauge) Set(v float64) {
	if og.collector != nil {
		og.collector.gaugeMutex.Lock()
		defer og.collector.gaugeMutex.Unlock()

		og.collector.gaugeValues[og.key] = int64(v)
	}
}

func (og *otelGauge) Inc() {
	if og.collector != nil {
		og.collector.gaugeMutex.Lock()
		defer og.collector.gaugeMutex.Unlock()

		og.collector.gaugeValues[og.key]++
	}
}

func (og *otelGauge) Dec() {
	if og.collector != nil {
		og.collector.gaugeMutex.Lock()
		defer og.collector.gaugeMutex.Unlock()

		og.collector.gaugeValues[og.key]--
	}
}

func (og *otelGauge) Add(v float64) {
	if og.collector != nil {
		og.collector.gaugeMutex.Lock()
		defer og.collector.gaugeMutex.Unlock()

		og.collector.gaugeValues[og.key] += int64(v)
	}
}

type otelHistogram struct {
	histogram metric.Float64Histogram
}

func (oh *otelHistogram) Observe(value float64) {
	if oh.histogram != nil {
		oh.histogram.Record(context.Background(), value)
	}
}

// Default Hop metrics
func (oc *OpenTelemetryCollector) registerDefaultMetrics() {
	// MessagesConsumed: counter with consumer, queue labels
	oc.Counter("hop_messages_consumed_total", "consumer", "queue")

	// ConsumptionErrors: counter with consumer, error_type labels
	oc.Counter("hop_consumption_errors_total", "consumer", "error_type")

	// Reconnects: counter without labels
	oc.Counter("hop_reconnects_total")

	// ConnectionDuration: gauge without labels
	oc.Gauge("hop_connection_duration_seconds")

	// ActiveConsumers: gauge without labels
	oc.Gauge("hop_active_consumers")

	// MessageProcessingDuration: histogram with consumer, queue labels
	oc.Histogram("hop_message_processing_duration_seconds", "consumer", "queue")
}
