package protocol

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
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
	// PrefetchCount limits how many unacknowledged messages the server will deliver.
	// Set to 0 for unlimited (not recommended for high throughput).
	PrefetchCount int
	// Queue configuration for the target queue.
	Queue Queue
	// Exchange configuration (optional). If nil, uses default exchange.
	Exchange *Exchange
	Exec     Handler // Public field for handler function (required)
	channel  *amqp.Channel
	msg      <-chan amqp.Delivery
}

// Validate checks if the consumer configuration is valid.
// Returns an error if:
// - Consumer name is empty
// - Handler function is nil
// - Queue configuration is invalid
// - Exchange configuration is invalid (if provided)
func (c *Consumer) Validate() error {
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

func (c *Consumer) Channel(channel *amqp.Channel) {
	c.channel = channel
}

func (c *Consumer) Close() error {
	if c.channel == nil {
		return nil
	}
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel for consumer %s: %w", c.Name, err)
	}
	return nil
}

func (c *Consumer) Listen() <-chan amqp.Delivery {
	// Check if channel is closed or nil
	if c.msg == nil || isChannelClosed(c.msg) {
		if c.channel == nil || c.channel.IsClosed() {
			log.Error().Msg("cannot start consumer: channel is nil or closed")
			return nil
		}

		// Set QoS before consuming if PrefetchCount is defined
		if c.PrefetchCount > 0 {
			if err := c.channel.Qos(c.PrefetchCount, 0, false); err != nil {
				log.Error().Err(err).Msg("failed to set QoS")
				return nil
			}
		}

		var err error
		if c.msg, err = c.channel.Consume(
			c.Queue.Name,
			c.Name,
			c.AutoAck,
			c.Exclusive,
			c.NoLocal,
			c.NoWait,
			c.Headers,
		); err != nil {
			log.Error().
				Err(err).
				Str("consumer", c.Name).
				Str("queue", c.Queue.Name).
				Msg("failed to start consumer")

			return nil
		}
	}

	return c.msg
}

// isChannelClosed checks if a channel is closed without blocking
func isChannelClosed(ch <-chan amqp.Delivery) bool {
	select {
	case _, ok := <-ch:
		return !ok // closed if ok is false
	default:
		return false // channel is open and has no data
	}
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

// Start begins the consumer loop. It listens for messages and processes them.
// It blocks until the context is cancelled or a fatal error occurs.
func (c *Consumer) Start(ctx context.Context) error {
	deliveries := c.Listen()
	if deliveries == nil {
		return errors.New("failed to get delivery channel")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				// Channel closed, likely due to connection issue
				return errors.New("delivery channel closed")
			}

			msg := Message{Delivery: d}

			// Execute the handler
			err := c.Execute(ctx, msg)

			// Handle Ack/Nack (if not AutoAck)
			if !c.AutoAck {
				if err != nil {
					// Handler failed: Nack the message
					// Note: You may want to make requeue configurable
					if nErr := msg.Failure(false, true); nErr != nil {
						log.Error().Err(nErr).Msg("failed to Nack message")
					}
				} else {
					// Handler succeeded: Ack the message
					if aErr := msg.Success(false); aErr != nil {
						log.Error().Err(aErr).Msg("failed to Ack message")
					}
				}
			}
		}
	}
}
