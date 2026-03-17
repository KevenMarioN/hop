package protocol

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

// Validate checks if the consumer configuration is valid.
// Returns an error if:
// - Consumer name is empty
// - Handler function is nil
// - Queue configuration is invalid
// - Exchange configuration is invalid (if provided)
func (c Consumer) Validate() error {
	var errs = make([]error, 0)

	if c.Name == "" {
		errs = append(errs, errors.New("consumer name cannot be empty"))
	}

	if c.Exec == nil {
		errs = append(errs, errors.New("handler cannot be empty"))
	}

	// Validate queue configuration
	if err := c.Queue.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("queue validation failed: %w", err))
	}

	// Validate exchange configuration if provided
	if c.Exchange != nil {
		if err := c.Exchange.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("exchange validation failed: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("consumer validation failed: %w", errors.Join(errs...))
	}

	return nil
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
