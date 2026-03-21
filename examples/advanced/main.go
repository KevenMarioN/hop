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
)

func main() {
	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Capture interrupt signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create connection with RabbitMQ with custom configuration
	hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/",
		hop.WithConnectionName("meu-app-advanced"),
	)
	if err != nil {
		fmt.Printf("Failed to create connection: %v", err)
		return
	}

	defer func() {
		if err := hopClient.Close(); err != nil {
			fmt.Print(err)
		}
	}()

	// Register consumer 1 - Order processing
	err = hopClient.Consume(hop.Consumer{
		Name:    "order-processor",
		AutoAck: false,
		Queue: hop.Queue{
			Name:    "orders",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg hop.Message) error {
			defer func() {
				if err := msg.Ack(true); err != nil {
					fmt.Print(err)
				}
			}()

			fmt.Printf("[Order] Processing order: %s\n", string(msg.Body))
			// Simulate processing
			time.Sleep(100 * time.Millisecond)

			return nil
		},
	})
	if err != nil {
		fmt.Printf("Failed to register order-processor consumer: %v", err)
		return
	}

	// Register consumer 2 - Notification processing
	err = hopClient.Consume(hop.Consumer{
		Name:    "notification-processor",
		AutoAck: false,
		Queue: hop.Queue{
			Name:    "notifications",
			Durable: true,
		},
		Exec: func(ctx context.Context, msg hop.Message) error {
			defer func() {
				if err := msg.Ack(true); err != nil {
					fmt.Print(err)
				}
			}()

			fmt.Printf("[Notification] Processing notification: %s\n", string(msg.Body))
			// Simulate processing
			time.Sleep(50 * time.Millisecond)

			return nil
		},
	})
	if err != nil {
		fmt.Printf("Failed to register notification-processor consumer: %v", err)
		return
	}

	// Register consumer 3 - Log processing
	err = hopClient.Consume(hop.Consumer{
		Name:    "log-processor",
		AutoAck: true, // Auto-ack for logs
		Queue: hop.Queue{
			Name:    "logs",
			Durable: false, // Non-durable queue for logs
		},
		Exec: func(ctx context.Context, msg hop.Message) error {
			fmt.Printf("[Log] %s\n", string(msg.Body))
			return nil
		},
	})
	if err != nil {
		fmt.Printf("Failed to register log-processor consumer: %v", err)
	}

	// Start all consumers
	hopClient.StartConsumers(ctx)

	fmt.Println("Server started. Press Ctrl+C to exit...")

	// Wait for interrupt signal or timeout
	select {
	case <-sigChan:
		fmt.Println("\nInterrupt signal received. Shutting down...")
	case <-ctx.Done():
		fmt.Println("\nTimeout reached. Shutting down...")
	}

	// Graceful shutdown
	fmt.Println("Waiting for message processing to complete...")

	if err := hopClient.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("Server shut down successfully.")
}
