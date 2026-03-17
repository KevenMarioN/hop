// Package hop provides a simple and resilient RabbitMQ client for Go.
//
// It offers automatic reconnection, graceful shutdown, and message consumption
// with built-in metrics support via Prometheus.
package hop

import (
	"context"

	"github.com/KevenMarioN/hop/internal/conn"
)

// Client is the main interface for interacting with RabbitMQ.
type Client interface {
	// Publish publishes a message to an exchange.
	// NOTE: Not implemented yet.
	Publish(ctx context.Context, exchange, key string, body []byte) error

	// Consume registers a consumer configuration.
	// The consumer will start processing when StartConsumers is called.
	Consume(args Consumer) error

	// StartConsumers begins message processing for all registered consumers.
	// This spawns goroutines and returns immediately.
	StartConsumers(ctx context.Context)

	// Close terminates the connection immediately without waiting for consumers.
	// For graceful shutdown, use Shutdown instead.
	Close() error

	// Shutdown gracefully stops all consumers and closes the connection.
	// It waits for active message processing to complete.
	Shutdown(ctx context.Context) error
}

// New creates a new Hop client connected to RabbitMQ.
// - ctx: Context for connection lifecycle
// - url: AMQP connection URL (e.g., amqp://user:pass@host:5672/)
// - opts: Optional configuration (connection name, backoff, TLS, metrics)
// Returns a Client implementation or error if connection fails.
func New(ctx context.Context, url string, opts ...conn.HopOption) (Client, error) {
	return conn.Connect(ctx, url, opts...)
}
