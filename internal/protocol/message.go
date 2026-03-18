package protocol

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
