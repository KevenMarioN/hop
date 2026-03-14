package main

import (
	"context"
	"fmt"
	"log"

	"github.com/KevenMarioN/hop"
	"github.com/KevenMarioN/hop/protocol"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx := context.Background()

	// Cria conexão com RabbitMQ
	hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/")
	if err != nil {
		log.Fatalf("Falha ao criar conexão: %v", err)
	}

	defer func() {
		if err := hopClient.Close(); err != nil {
			fmt.Print(err)
		}
	}()

	// Registra consumer
	err = hopClient.Consume(protocol.Consumer{
		Name:    "my-consumer",
		AutoAck: false,
		Queue: protocol.Queue{
			Name:    "my-queue",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
			defer func() {
				if err := msg.Ack(true); err != nil {
					fmt.Print(err)
				}
			}()

			fmt.Printf("Mensagem recebida: %s\n", string(msg.Body))

			return nil
		},
	})
	if err != nil {
		fmt.Printf("Falha ao registrar consumer: %v", err)
		return
	}

	// Inicia consumers
	hopClient.StartConsumers(ctx)

	// Aguarda conclusão
	if err := hopClient.Shutdown(ctx); err != nil {
		fmt.Printf("Falha no shutdown: %v", err)
		return
	}
}
