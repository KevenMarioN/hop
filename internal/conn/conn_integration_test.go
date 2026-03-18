package conn

import (
	"context"
	"testing"
	"time"

	"github.com/KevenMarioN/hop/internal/metrics"
	"github.com/KevenMarioN/hop/internal/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TestIntegrationPublishAndConsume tests publishing and consuming messages with real RabbitMQ
func TestIntegrationPublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Create channel directly with AMQP (since Publish is not implemented in hop)
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect for publishing: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}
	defer channel.Close()

	// Declarar exchange e queue
	exchangeName := "test_integration_exchange"
	queueName := "test_integration_queue"

	err = channel.ExchangeDeclare(
		exchangeName,
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare exchange: %v", err)
	}

	_, err = channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	err = channel.QueueBind(
		queueName,
		"test_key",
		exchangeName,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Failed to bind queue: %v", err)
	}

	// Publish message
	testMessage := []byte("Integration test message")

	err = channel.Publish(
		exchangeName,
		"test_key",
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        testMessage,
		},
	)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Consume message
	msgs, err := channel.Consume(
		queueName,
		"test_consumer",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Failed to consume messages: %v", err)
	}

	// Wait for message
	select {
	case msg := <-msgs:
		if string(msg.Body) != string(testMessage) {
			t.Errorf("Received message different from sent. Expected: %s, Received: %s",
				string(testMessage), string(msg.Body))
		}

		msg.Ack(false)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

// TestIntegrationWithMetrics tests integration with metrics
func TestIntegrationWithMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create metrics collector (NopCollector is a variable, not a function)
	collector := metrics.NopCollector

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/", WithMetrics(collector))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Verificar que a conexão foi estabelecida
	if client == nil {
		t.Fatal("Client is nil")
	}
}

// TestIntegrationConsumerManager tests Manager with real RabbitMQ
func TestIntegrationConsumerManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// Create test consumer using ConsumerBuilder
	testQueue := "test_manager_queue"

	testConsumer, err := protocol.NewConsumerBuilder("test_manager_consumer").
		WithQueue(protocol.Queue{
			Name:              testQueue,
			ShouldCreateQueue: true,
			Durable:           true,
		}).
		WithHandler(func(ctx context.Context, msg protocol.Message) error {
			// Process message
			t.Logf("Received message: %s", string(msg.Body))
			return nil
		}).
		Build()
	if err != nil {
		t.Fatalf("Failed to create consumer: %v", err)
	}

	// Register consumer
	err = client.Consume(*testConsumer)
	if err != nil {
		t.Fatalf("Failed to register consumer: %v", err)
	}

	// Iniciar consumers
	client.StartConsumers(ctx)

	// Publish test message using separate connection
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Failed to connect for publishing: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		testQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	testMessage := []byte("Message for manager")

	err = channel.Publish(
		"",
		testQueue,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        testMessage,
		},
	)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Close the connection (this should trigger shutdown)
	if err := client.Close(); err != nil {
		t.Logf("Error closing connection: %v", err)
	}
}
