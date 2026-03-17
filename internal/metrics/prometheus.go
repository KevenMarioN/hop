package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusCollector implementa MetricsCollector usando Prometheus.
type PrometheusCollector struct {
	registry prometheus.Registerer
	// Cache de métricas para evitar lookup repetido
	counters      map[string]prometheus.Counter
	gauges        map[string]prometheus.Gauge
	histogramVecs map[string]*prometheus.HistogramVec
}

// NewPrometheusCollector cria um novo collector Prometheus.
func NewPrometheusCollector(registry prometheus.Registerer) *PrometheusCollector {
	p := &PrometheusCollector{
		registry:      registry,
		counters:      make(map[string]prometheus.Counter),
		gauges:        make(map[string]prometheus.Gauge),
		histogramVecs: make(map[string]*prometheus.HistogramVec),
	}
	// Registrar métricas padrão do Hop
	p.registerDefaultMetrics()

	return p
}

func (p *PrometheusCollector) Counter(name string, labels ...string) Counter {
	key := name + "|" + strings.Join(labels, "|")
	if c, ok := p.counters[key]; ok {
		return &prometheusCounter{c}
	}
	// Criar nova métrica
	vec := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: name, Help: name},
		labels,
	)
	p.registry.MustRegister(vec)
	c := vec.WithLabelValues(labels...)
	p.counters[key] = c

	return &prometheusCounter{c}
}

func (p *PrometheusCollector) Gauge(name string, labels ...string) Gauge {
	key := name + "|" + strings.Join(labels, "|")
	if g, ok := p.gauges[key]; ok {
		return &prometheusGauge{g}
	}

	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: name, Help: name},
		labels,
	)
	p.registry.MustRegister(vec)
	g := vec.WithLabelValues(labels...)
	p.gauges[key] = g

	return &prometheusGauge{g}
}

func (p *PrometheusCollector) Histogram(name string, labels ...string) Histogram {
	key := name + "|" + strings.Join(labels, "|")
	if vec, ok := p.histogramVecs[key]; ok {
		return &prometheusHistogram{vec.WithLabelValues(labels...)}
	}

	vec := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    name,
			Buckets: prometheus.DefBuckets,
		},
		labels,
	)
	p.registry.MustRegister(vec)
	p.histogramVecs[key] = vec

	return &prometheusHistogram{vec.WithLabelValues(labels...)}
}

func (p *PrometheusCollector) Registerer() any {
	return p.registry
}

// Implementações wrapper
type prometheusCounter struct{ c prometheus.Counter }

func (pc *prometheusCounter) Inc()          { pc.c.Inc() }
func (pc *prometheusCounter) Add(v float64) { pc.c.Add(v) }

type prometheusGauge struct{ g prometheus.Gauge }

func (pg *prometheusGauge) Set(v float64) { pg.g.Set(v) }
func (pg *prometheusGauge) Inc()          { pg.g.Inc() }
func (pg *prometheusGauge) Dec()          { pg.g.Dec() }
func (pg *prometheusGauge) Add(v float64) { pg.g.Add(v) }

type prometheusHistogram struct{ h prometheus.Observer }

func (ph *prometheusHistogram) Observe(value float64) { ph.h.Observe(value) }

// Métricas padrão do Hop
func (p *PrometheusCollector) registerDefaultMetrics() {
	// MessagesConsumed: counter com labels consumer, queue
	messagesConsumed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hop_messages_consumed_total",
			Help: "Total number of messages consumed",
		},
		[]string{"consumer", "queue"},
	)
	p.registry.MustRegister(messagesConsumed)
	p.counters["hop_messages_consumed_total|consumer|queue"] = messagesConsumed.WithLabelValues()

	// ConsumptionErrors: counter com labels consumer, error_type
	consumptionErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hop_consumption_errors_total",
			Help: "Total number of consumption errors",
		},
		[]string{"consumer", "error_type"},
	)
	p.registry.MustRegister(consumptionErrors)
	p.counters["hop_consumption_errors_total|consumer|error_type"] = consumptionErrors.WithLabelValues()

	// Reconnects: counter sem labels
	reconnects := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "hop_reconnects_total",
			Help: "Total number of reconnections",
		},
	)
	p.registry.MustRegister(reconnects)
	p.counters["hop_reconnects_total|"] = reconnects

	// ConnectionDuration: gauge sem labels
	connectionDuration := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hop_connection_duration_seconds",
			Help: "Duration of current connection in seconds",
		},
	)
	p.registry.MustRegister(connectionDuration)
	p.gauges["hop_connection_duration_seconds|"] = connectionDuration

	// ActiveConsumers: gauge sem labels
	activeConsumers := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hop_active_consumers",
			Help: "Number of active consumers",
		},
	)
	p.registry.MustRegister(activeConsumers)
	p.gauges["hop_active_consumers|"] = activeConsumers

	// MessageProcessingDuration: histogram com labels consumer, queue
	messageProcessingDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hop_message_processing_duration_seconds",
			Help:    "Duration of message processing in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"consumer", "queue"},
	)
	p.registry.MustRegister(messageProcessingDuration)
	p.histogramVecs["hop_message_processing_duration_seconds|consumer|queue"] = messageProcessingDuration
}
