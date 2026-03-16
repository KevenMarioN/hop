package conn

import (
	"context"
	"testing"
	"time"

	"github.com/KevenMarioN/hop/metrics"
	"github.com/KevenMarioN/hop/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TestIntegrationPublishAndConsume testa publicação e consumo de mensagem com RabbitMQ real
func TestIntegrationPublishAndConsume(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo curto")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Falha ao conectar: %v", err)
	}
	defer client.Close()

	// Criar canal diretamente com AMQP (pois Publish não está implementado no hop)
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Falha ao conectar para publicação: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		t.Fatalf("Falha ao criar canal: %v", err)
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
		t.Fatalf("Falha ao declarar exchange: %v", err)
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
		t.Fatalf("Falha ao declarar queue: %v", err)
	}

	err = channel.QueueBind(
		queueName,
		"test_key",
		exchangeName,
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		t.Fatalf("Falha ao bind queue: %v", err)
	}

	// Publicar mensagem
	testMessage := []byte("Mensagem de teste de integração")
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
		t.Fatalf("Falha ao publicar mensagem: %v", err)
	}

	// Consumir mensagem
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
		t.Fatalf("Falha ao consumir mensagens: %v", err)
	}

	// Esperar pela mensagem
	select {
	case msg := <-msgs:
		if string(msg.Body) != string(testMessage) {
			t.Errorf("Mensagem recebida diferente da enviada. Esperado: %s, Recebido: %s",
				string(testMessage), string(msg.Body))
		}
		msg.Ack(false)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout esperando mensagem")
	}
}

// TestIntegrationWithMetrics testa integração com métricas
func TestIntegrationWithMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo curto")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Criar collector de métricas (NopCollector é uma variável, não uma função)
	collector := metrics.NopCollector

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/", WithMetrics(collector))
	if err != nil {
		t.Fatalf("Falha ao conectar: %v", err)
	}
	defer client.Close()

	// Verificar que a conexão foi estabelecida
	if client == nil {
		t.Fatal("Cliente é nulo")
	}
}

// TestIntegrationConsumerManager testa Manager com RabbitMQ real
func TestIntegrationConsumerManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste de integração em modo curto")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Conectar ao RabbitMQ local
	client, err := Connect(ctx, "amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Falha ao conectar: %v", err)
	}
	defer client.Close()

	// Criar consumer de teste usando ConsumerBuilder
	testQueue := "test_manager_queue"
	testConsumer, err := protocol.NewConsumerBuilder("test_manager_consumer").
		WithQueue(protocol.Queue{
			Name:              testQueue,
			ShouldCreateQueue: true,
			Durable:           true,
		}).
		WithHandler(func(ctx context.Context, msg protocol.Message) error {
			// Processar mensagem
			t.Logf("Mensagem recebida: %s", string(msg.Body))
			return nil
		}).
		Build()
	if err != nil {
		t.Fatalf("Falha ao criar consumer: %v", err)
	}

	// Registrar consumer
	err = client.Consume(*testConsumer)
	if err != nil {
		t.Fatalf("Falha ao registrar consumer: %v", err)
	}

	// Iniciar consumers
	client.StartConsumers(ctx)

	// Publicar mensagem de teste usando conexão separada
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	if err != nil {
		t.Fatalf("Falha ao conectar para publicação: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		t.Fatalf("Falha ao criar canal: %v", err)
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
		t.Fatalf("Falha ao declarar queue: %v", err)
	}

	testMessage := []byte("Mensagem para manager")
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
		t.Fatalf("Falha ao publicar mensagem: %v", err)
	}

	// Aguardar processamento
	time.Sleep(500 * time.Millisecond)

	// Fechar a conexão (isso deve disparar o shutdown)
	if err := client.Close(); err != nil {
		t.Logf("Erro ao fechar conexão: %v", err)
	}
}
