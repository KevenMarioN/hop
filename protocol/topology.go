package protocol

import (
	"context"
	"errors"
	"fmt"

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
type Handler func(ctx context.Context, msg Message) error

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
	// ShouldCreateExchange indicates if this library should declare the exchange.
	// Set to false if exchange is managed externally.
	ShouldCreateExchange bool
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
	channel  *amqp.Channel
}

// Message wraps amqp.Delivery to provide a cleaner interface for message handling.
// It embeds all fields and methods from amqp.Delivery while allowing for future
// Hop-specific extensions to the message format.
//
// Key embedded fields from amqp.Delivery:
// - Body: []byte containing the message payload
// - Headers: map[string]interface{} with message headers
// - ContentType: string describing the message format
// - DeliveryTag: uint64 identifier for delivery tracking
// - Exchange: string name of originating exchange
// - RoutingKey: string routing key used for delivery
// - ConsumerTag: string identifier of the consumer
// - MessageCount: uint32 number of messages remaining in queue
type Message struct {
	amqp.Delivery
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

func (c *Consumer) Channel(channel *amqp.Channel) {
	c.channel = channel
}

func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed close channel to consumer %s", c.Name)
	}
	return nil
}

func (c *Consumer) Listen() <-chan amqp.Delivery {
	return c.msg
}

func (c *Consumer) Handler(handler Handler) {
	c.Exec = handler
}

// Execute processes a single message using the consumer's handler function.
// It validates the message and invokes the registered handler, which should
// return nil to acknowledge the message or an error to trigger retry/NACK behavior.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - msg: The message to process (wrapped amqp.Delivery)
//
// Returns:
//   - error: nil on success, or an error if the handler fails
//
// Note: Per-message logging has been removed for performance. Metrics are
// automatically tracked for observability. Enable debug logging in zerolog
// if per-message logging is needed.
func (c *Consumer) Execute(ctx context.Context, msg Message) error {
	// Logging removed for performance; metrics track consumption.
	// Enable debug logging in zerolog if per-message logging is needed.
	return c.Exec(ctx, msg)
}

// NewMessage creates a new Message wrapper from an amqp.Delivery.
// This is useful when you need to pass a delivery to a handler that expects
// the Message type instead of the raw amqp.Delivery.
//
// Parameters:
//   - delivery: The raw AMQP delivery from the RabbitMQ broker
//
// Returns:
//   - Message: A wrapped delivery with additional helper methods
func NewMessage(delivery amqp.Delivery) Message {
	return Message{delivery}
}

// success acknowledges the message with the specified multi flag.
// This is a private helper method used by Success() and SuccessMultiple().
func (m *Message) success(multi bool) error {
	if err := m.Ack(multi); err != nil {
		return fmt.Errorf("failed confirm: %w", err)
	}
	return nil
}

// Success acknowledges the message as successfully processed.
// This sends a basic.ack to the broker, confirming the message was handled correctly.
//
// Returns:
//   - error: nil on success, or an error if the acknowledgment fails
func (m *Message) Success() error {
	return m.success(false)
}

// SuccessMultiple acknowledges multiple messages as successfully processed.
// This sends a basic.ack with the multiple flag set to true, confirming all
// messages up to and including this one were handled correctly.
//
// Returns:
//   - error: nil on success, or an error if the acknowledgment fails
func (m *Message) SuccessMultiple() error {
	return m.success(true)
}

// failure rejects the message with the specified flags.
// This is a private helper method used by Failure(), FailureMultiple(),
// Retry(), and RetryMultiple().
func (m *Message) failure(multi, requeue bool) error {
	if err := m.Nack(multi, requeue); err != nil {
		return fmt.Errorf("failed failure: %w", err)
	}
	return nil
}

// Failure rejects the message as unsuccessfully processed.
// This sends a basic.nack to the broker without requeueing the message.
//
// Returns:
//   - error: nil on success, or an error if the rejection fails
func (m *Message) Failure() error {
	return m.failure(false, false)
}

// FailureMultiple rejects multiple messages as unsuccessfully processed.
// This sends a basic.nack with the multiple flag set to true, rejecting all
// messages up to and including this one without requeueing them.
//
// Returns:
//   - error: nil on success, or an error if the rejection fails
func (m *Message) FailureMultiple() error {
	return m.failure(true, false)
}

// Retry rejects the message for retry.
// This sends a basic.nack to the broker with requeue set to true,
// causing the message to be redelivered to the same or another consumer.
//
// Returns:
//   - error: nil on success, or an error if the rejection fails
func (m *Message) Retry() error {
	return m.failure(false, true)
}

// RetryMultiple rejects multiple messages for retry.
// This sends a basic.nack with the multiple flag set to true and requeue set to true,
// causing all messages up to and including this one to be redelivered.
//
// Returns:
//   - error: nil on success, or an error if the rejection fails
func (m *Message) RetryMultiple() error {
	return m.failure(true, true)
}
