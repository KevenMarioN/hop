package metrics

import (
	"context"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// OpenTelemetryCollector implementa MetricsCollector usando OpenTelemetry.
type OpenTelemetryCollector struct {
	meter       metric.Meter
	counters    map[string]metric.Int64Counter
	gauges      map[string]metric.Int64ObservableGauge
	gaugeValues map[string]int64
	gaugeMutex  sync.RWMutex
	histograms  map[string]metric.Float64Histogram
}

// NewOpenTelemetryCollector cria um novo collector OpenTelemetry.
func NewOpenTelemetryCollector(meterName string) *OpenTelemetryCollector {
	meter := otel.Meter(meterName)

	oc := &OpenTelemetryCollector{
		meter:       meter,
		counters:    make(map[string]metric.Int64Counter),
		gauges:      make(map[string]metric.Int64ObservableGauge),
		gaugeValues: make(map[string]int64),
		histograms:  make(map[string]metric.Float64Histogram),
	}

	// Registrar métricas padrão do Hop
	oc.registerDefaultMetrics()

	return oc
}

func (oc *OpenTelemetryCollector) Counter(name string, labels ...string) Counter {
	key := name + "|" + strings.Join(labels, "|")

	if c, ok := oc.counters[key]; ok {
		return &otelCounter{counter: c, labels: labels}
	}

	// Criar novo counter
	counter, err := oc.meter.Int64Counter(
		name,
		metric.WithDescription(name),
	)
	if err != nil {
		// Em caso de erro, retornar counter no-op
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

	// Criar novo gauge observável
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
		// Em caso de erro, retornar gauge no-op
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

	// Criar novo histogram
	histogram, err := oc.meter.Float64Histogram(
		name,
		metric.WithDescription(name),
	)
	if err != nil {
		// Em caso de erro, retornar histogram no-op
		return &otelHistogram{}
	}

	oc.histograms[key] = histogram
	return &otelHistogram{histogram: histogram}
}

func (oc *OpenTelemetryCollector) Registerer() any {
	return nil // OpenTelemetry não usa registerer
}

// Implementações wrapper
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

// Métricas padrão do Hop
func (oc *OpenTelemetryCollector) registerDefaultMetrics() {
	// MessagesConsumed: counter com labels consumer, queue
	oc.Counter("hop_messages_consumed_total", "consumer", "queue")

	// ConsumptionErrors: counter com labels consumer, error_type
	oc.Counter("hop_consumption_errors_total", "consumer", "error_type")

	// Reconnects: counter sem labels
	oc.Counter("hop_reconnects_total")

	// ConnectionDuration: gauge sem labels
	oc.Gauge("hop_connection_duration_seconds")

	// ActiveConsumers: gauge sem labels
	oc.Gauge("hop_active_consumers")

	// MessageProcessingDuration: histogram com labels consumer, queue
	oc.Histogram("hop_message_processing_duration_seconds", "consumer", "queue")
}
