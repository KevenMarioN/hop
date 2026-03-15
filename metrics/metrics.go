package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics for the Hop RabbitMQ client.
// All metrics are automatically registered when WithMetrics option is used.

// MessagesConsumed counts total messages successfully processed by each consumer.
// Labels: consumer (consumer name), queue (queue name)
var MessagesConsumed = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "hop_messages_consumed_total",
		Help: "Total number of messages consumed",
	},
	[]string{"consumer", "queue"},
)

// ConsumptionErrors counts errors that occur during message processing.
// Labels: consumer (consumer name), error_type (type of error)
var ConsumptionErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "hop_consumption_errors_total",
		Help: "Total number of consumption errors",
	},
	[]string{"consumer", "error_type"},
)

// Reconnects counts total reconnection attempts to RabbitMQ.
var Reconnects = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "hop_reconnects_total",
		Help: "Total number of reconnections",
	},
)

// ConnectionDuration tracks current connection uptime in seconds.
// This gauge is updated when consumers finish or connection is closed.
var ConnectionDuration = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "hop_connection_duration_seconds",
		Help: "Duration of current connection in seconds",
	},
)

// ActiveConsumers tracks the number of currently active consumers.
var ActiveConsumers = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "hop_active_consumers",
		Help: "Number of active consumers",
	},
)

// MustRegister registers all metrics with the provided registry.
// This function is called automatically by WithMetrics option.
// Panics if any metric fails to register.
func MustRegister(registry prometheus.Registerer) {
	registry.MustRegister(
		MessagesConsumed,
		ConsumptionErrors,
		Reconnects,
		ConnectionDuration,
		ActiveConsumers,
	)
}
