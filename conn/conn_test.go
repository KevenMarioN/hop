package conn

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KevenMarioN/hop/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TestWaitWithoutConsumers tests wait without consumers
func TestWaitWithoutConsumers(t *testing.T) {
	hop := &hop{
		consumers: make(map[string]protocol.Consumer),
	}

	err := hop.Wait()
	if err == nil {
		t.Error("Esperado erro 'no consumers registered', mas obteve nil")
	}

	if !errors.Is(err, ErrNoConsumers) {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestPublishNotImplemented tests that publish is not implemented
func TestPublishNotImplemented(t *testing.T) {
	hop := &hop{}

	err := hop.Publish(context.Background(), "exchange", "key", []byte("message"))
	if err == nil {
		t.Error("Esperado erro 'not implemented', mas obteve nil")
	}

	if !errors.Is(err, ErrNotImplemented) {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestConsumeWithValidationError tests consume with validation error
func TestConsumeWithValidationError(t *testing.T) {
	hop := &hop{
		conn: &amqp.Connection{},
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
		t.Error("Esperado erro de validação, mas obteve nil")
	}

	if !strings.Contains(err.Error(), "consumer name cannot be empty") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestConsumeWithNullHandler tests consume with null handler
func TestConsumeWithNullHandler(t *testing.T) {
	hop := &hop{
		conn: &amqp.Connection{},
	}

	// Consumer com handler nulo
	consumer := protocol.Consumer{
		Name: "test-consumer",
		Exec: nil,
	}

	err := hop.Consume(consumer)
	if err == nil {
		t.Error("Esperado erro de validação, mas obteve nil")
	}

	if !strings.Contains(err.Error(), "handler cannot be empty") {
		t.Errorf("Erro inesperado: %v", err)
	}
}

// TestCloseWithClosedConnection tests close with closed connection
func TestCloseWithClosedConnection(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithConsumers tests shutdown with active consumers
func TestShutdownWithConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithoutConsumers tests shutdown without consumers
func TestShutdownWithoutConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumersWithConsumers tests start consumers with registered consumers
func TestStartConsumersWithConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumersWithoutConsumers tests start consumers without consumers
func TestStartConsumersWithoutConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestWaitWithConsumers tests wait with active consumers
func TestWaitWithConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestMonitorConnectionWithCancelledContext tests monitorConnection with cancelled context
func TestMonitorConnectionWithCancelledContext(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestMonitorConnectionWithClosedConnection tests monitorConnection with closed connection
func TestMonitorConnectionWithClosedConnection(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestRestartConsumersWithConsumers tests restart consumers with registered consumers
func TestRestartConsumersWithConsumers(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithCancelledContext tests start consumer with cancelled context
func TestStartConsumerWithCancelledContext(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithClosedChannel tests start consumer with closed channel
func TestStartConsumerWithClosedChannel(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithReconnection tests start consumer with reconnection
func TestStartConsumerWithReconnection(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithQueueError tests consume with queue declaration error
func TestConsumeWithQueueError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithConsumerError tests consume with consumer error
func TestConsumeWithConsumerError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestConsumeWithExistingConsumer tests consume with existing consumer
func TestConsumeWithExistingConsumer(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestCloseWithError tests close with error
func TestCloseWithError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithWaitError tests shutdown with wait error
func TestShutdownWithWaitError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestShutdownWithCloseError tests shutdown with close error
func TestShutdownWithCloseError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestWaitWithError tests wait with error
func TestWaitWithError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}

// TestStartConsumerWithHandlerError tests start consumer with handler error
func TestStartConsumerWithHandlerError(t *testing.T) {
	// Este teste requer RabbitMQ em execução
	t.Skip("Requer RabbitMQ em execução")
}
