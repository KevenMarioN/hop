package protocol

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Kind represents the type of AMQP exchange.
type Kind string

func (k Kind) String() string {
	return string(k)
}

// Supported exchange types.
const (
	Fanout  Kind = "fanout" // Fanout exchange broadcasts to all bound queues
	Topic   Kind = "topic"  // Topic exchange routes based on pattern matching
	Direct  Kind = "direct" // Direct exchange routes by exact routing key
	Default Kind = ""       // Default exchange (amq.direct)
)

// Handler is a function that processes a consumed message.
// It receives the context (for cancellation/timeout) and the AMQP delivery.
// Return nil to acknowledge the message, or an error to trigger retry/NACK.
type Handler func(ctx context.Context, msg amqp.Delivery) error

// Queue defines the configuration for a RabbitMQ queue.
type Queue struct {
	// Durable indicates if the queue survives broker restarts.
	Durable bool
	// AutoDelete indicates if the queue is automatically deleted when no consumers.
	AutoDelete bool
	// Exclusive indicates if the queue is only accessible by the current connection.
	Exclusive bool
	// NoWait indicates if the server should respond immediately (no wait for confirmation).
	NoWait bool
	// Name is the queue identifier. Empty means server-generated name.
	Name string
	// Headers are additional arguments for queue declaration (used by plugins).
	Headers map[string]any
	// ShouldCreateQueue indicates if this library should declare the queue.
	// Set to false if queue is managed externally.
	ShouldCreateQueue bool
}

// Exchange defines the configuration for a RabbitMQ exchange.
type Exchange struct {
	// Durable indicates if the exchange survives broker restarts.
	Durable bool
	// AutoDelete indicates if the exchange is automatically deleted when no queues bound.
	AutoDelete bool
	// Exclusive indicates if the exchange is only accessible by the current connection.
	Exclusive bool
	// NoWait indicates if the server should respond immediately.
	NoWait bool
	// Internal indicates if the exchange is for internal broker use only.
	Internal bool
	// Kind is the exchange type (fanout, topic, direct, etc.).
	Kind Kind
	// Name is the exchange identifier.
	Name string
	// Headers are additional arguments for exchange declaration.
	Headers map[string]any
}

// Consumer represents a complete consumer configuration including queue, exchange, and handler.
type Consumer struct {
	// Name is a unique identifier for this consumer (used for logging and management).
	Name string
	// Key is the routing key for binding queue to exchange. Empty uses queue name.
	Key string
	// AutoAck indicates if messages are automatically acknowledged upon receipt.
	AutoAck bool
	// NoLocal indicates if messages published on this connection are not consumed.
	NoLocal bool
	// Exclusive indicates if only this consumer can access the queue.
	Exclusive bool
	// NoWait indicates if the server should respond immediately to consume request.
	NoWait bool
	// Headers are additional arguments for the consume request.
	Headers map[string]any
	// Queue configuration for the target queue.
	Queue Queue
	// Exchange configuration (optional). If nil, uses default exchange.
	Exchange *Exchange
	msg      <-chan amqp.Delivery
	Exec     Handler // Public field for handler function (required)
}

func (c Consumer) Validate() error {
	var errs = make([]error, 0)
	if c.Name == "" {
		errs = append(errs, errors.New("consumer name cannot be empty"))
	}

	if c.Exec == nil {
		errs = append(errs, errors.New("handler cannot be empty"))
	}

	return errors.Join(errs...)
}

func (c *Consumer) Msg(msg <-chan amqp.Delivery) {
	c.msg = msg
}

func (c *Consumer) Listen() <-chan amqp.Delivery {
	return c.msg
}

func (c *Consumer) Handler(handler Handler) {
	c.Exec = handler
}

func (c *Consumer) Execute(ctx context.Context, msg amqp.Delivery) error {
	// Logging removed for performance; metrics track consumption.
	// Enable debug logging in zerolog if per-message logging is needed.
	return c.Exec(ctx, msg)
}
