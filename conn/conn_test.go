package conn

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KevenMarioN/hop/consumer"
	"github.com/KevenMarioN/hop/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TestWaitWithoutConsumers tests wait without consumers
func TestWaitWithoutConsumers(t *testing.T) {
	mgr := consumer.NewManager(&amqp.Connection{}, nil)
	hop := &hop{
		consumerMgr: mgr,
	}

	err := hop.Wait()
	if err == nil {
		t.Error("Expected error 'no consumers registered', but got nil")
	}

	if !strings.Contains(err.Error(), "no consumers registered") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestPublishNotImplemented tests that publish is not implemented
func TestPublishNotImplemented(t *testing.T) {
	hop := &hop{}

	err := hop.Publish(context.Background(), "exchange", "key", []byte("message"))
	if err == nil {
		t.Error("Expected error 'not implemented', but got nil")
	}

	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumeWithValidationError tests consume with validation error
func TestConsumeWithValidationError(t *testing.T) {
	hop := &hop{
		conn:        &amqp.Connection{},
		consumerMgr: consumer.NewManager(&amqp.Connection{}, nil),
	}

	// Consumer com nome vazio
	consumer := protocol.Consumer{
		Name: "",
		Exec: func(ctx context.Context, msg amqp.Delivery) error {
			return nil
		},
	}

	err := hop.Consume(consumer)
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "consumer name cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestConsumeWithNullHandler tests consume with null handler
func TestConsumeWithNullHandler(t *testing.T) {
	hop := &hop{
		conn:        &amqp.Connection{},
		consumerMgr: consumer.NewManager(&amqp.Connection{}, nil),
	}

	// Consumer com handler nulo
	consumer := protocol.Consumer{
		Name: "test-consumer",
		Exec: nil,
	}

	err := hop.Consume(consumer)
	if err == nil {
		t.Error("Expected validation error, but got nil")
	}

	if !strings.Contains(err.Error(), "handler cannot be empty") {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestCloseWithClosedConnection tests close with closed connection
func TestCloseWithClosedConnection(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithConsumers tests shutdown with active consumers
func TestShutdownWithConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithoutConsumers tests shutdown without consumers
func TestShutdownWithoutConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumersWithConsumers tests start consumers with registered consumers
func TestStartConsumersWithConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumersWithoutConsumers tests start consumers without consumers
func TestStartConsumersWithoutConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestWaitWithConsumers tests wait with active consumers
func TestWaitWithConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestMonitorConnectionWithCancelledContext tests monitorConnection with cancelled context
func TestMonitorConnectionWithCancelledContext(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestMonitorConnectionWithClosedConnection tests monitorConnection with closed connection
func TestMonitorConnectionWithClosedConnection(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestRestartConsumersWithConsumers tests restart consumers with registered consumers
func TestRestartConsumersWithConsumers(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithCancelledContext tests start consumer with cancelled context
func TestStartConsumerWithCancelledContext(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithClosedChannel tests start consumer with closed channel
func TestStartConsumerWithClosedChannel(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithReconnection tests start consumer with reconnection
func TestStartConsumerWithReconnection(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithQueueError tests consume with queue declaration error
func TestConsumeWithQueueError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithConsumerError tests consume with consumer error
func TestConsumeWithConsumerError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithExistingConsumer tests consume with existing consumer
func TestConsumeWithExistingConsumer(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestCloseWithError tests close with error
func TestCloseWithError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithWaitError tests shutdown with wait error
func TestShutdownWithWaitError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithCloseError tests shutdown with close error
func TestShutdownWithCloseError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestWaitWithError tests wait with error
func TestWaitWithError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithHandlerError tests start consumer with handler error
func TestStartConsumerWithHandlerError(t *testing.T) {
	// This test requires RabbitMQ running
	t.Skip("Requer RabbitMQ em execução")
}
