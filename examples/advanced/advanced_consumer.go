package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KevenMarioN/hop"
	"github.com/KevenMarioN/hop/conn"
	"github.com/KevenMarioN/hop/protocol"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	// Cria contexto com timeout para graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Captura sinais de interrupção para graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Cria conexão com RabbitMQ com configuração personalizada
	hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/",
		conn.WithConnectionName("meu-app-advanced"),
	)
	if err != nil {
		log.Fatalf("Falha ao criar conexão: %v", err)
	}
	defer hopClient.Close()

	// Registra consumer 1 - Processamento de pedidos
	err = hopClient.Consume(protocol.Consumer{
		Name:    "order-processor",
		AutoAck: false,
		Queue: protocol.Queue{
			Name:    "orders",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
			defer msg.Ack(true)
			fmt.Printf("[Order] Processando pedido: %s\n", string(msg.Body))
			// Simula processamento
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	})
	if err != nil {
		log.Fatalf("Falha ao registrar consumer order-processor: %v", err)
	}

	// Registra consumer 2 - Processamento de notificações
	err = hopClient.Consume(protocol.Consumer{
		Name:    "notification-processor",
		AutoAck: false,
		Queue: protocol.Queue{
			Name:    "notifications",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
			defer msg.Ack(true)
			fmt.Printf("[Notification] Processando notificação: %s\n", string(msg.Body))
			// Simula processamento
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	})
	if err != nil {
		log.Fatalf("Falha ao registrar consumer notification-processor: %v", err)
	}

	// Registra consumer 3 - Processamento de logs
	err = hopClient.Consume(protocol.Consumer{
		Name:    "log-processor",
		AutoAck: true, // Auto-ack para logs
		Queue: protocol.Queue{
			Name:    "logs",
			Durable: false, // Fila não durável para logs
		},
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
			fmt.Printf("[Log] %s\n", string(msg.Body))
			return nil
		},
	})
	if err != nil {
		log.Fatalf("Falha ao registrar consumer log-processor: %v", err)
	}

	// Inicia todos os consumers
	hopClient.StartConsumers(ctx)

	fmt.Println("Servidor iniciado. Pressione Ctrl+C para encerrar...")

	// Aguarda sinal de interrupção ou timeout
	select {
	case <-sigChan:
		fmt.Println("\nSinal de interrupção recebido. Encerrando...")
	case <-ctx.Done():
		fmt.Println("\nTimeout atingido. Encerrando...")
	}

	// Graceful shutdown
	fmt.Println("Aguardando conclusão das mensagens em processamento...")
	if err := hopClient.Shutdown(ctx); err != nil {
		log.Printf("Erro durante shutdown: %v", err)
	}

	fmt.Println("Servidor encerrado com sucesso.")
}
