package main

import (
	"context"
	"fmt"
	"log"

	"github.com/KevenMarioN/hop"
	"github.com/KevenMarioN/hop/protocol"
)

func main() {
	ctx := context.Background()

	// Create connection with RabbitMQ
	hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}

	defer func() {
		if err := hopClient.Close(); err != nil {
			fmt.Print(err)
		}
	}()

	// Register consumer
	err = hopClient.Consume(protocol.Consumer{
		Name:    "my-consumer",
		AutoAck: false,
		Queue: protocol.Queue{
			Name:    "my-queue",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg protocol.Message) error {
			defer func() {
				if err := msg.Ack(true); err != nil {
					fmt.Print(err)
				}
			}()

			fmt.Printf("Received message: %s\n", string(msg.Body))

			return nil
		},
	})
	if err != nil {
		fmt.Printf("Failed to register consumer: %v", err)
		return
	}

	// Start consumers
	hopClient.StartConsumers(ctx)

	// Wait for completion
	if err := hopClient.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown: %v", err)
		return
	}
}
