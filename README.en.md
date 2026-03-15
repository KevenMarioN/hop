# Hop - RabbitMQ Connection Library for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/KevenMarioN/hop)](https://goreportcard.com/report/github.com/KevenMarioN/hop)
[![GoDoc](https://godoc.org/github.com/KevenMarioN/hop?status.svg)](https://godoc.org/github.com/KevenMarioN/hop)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Hop is a simple and resilient Go library for RabbitMQ connection, with support for auto-reconnect, graceful shutdown, and message consumption.

## 🚀 Installation

```bash
go get github.com/KevenMarioN/hop
```

## 📦 Dependencies

- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - Official AMQP client
- [zerolog](https://github.com/rs/zerolog) - Structured logging
- [errgroup](https://golang.org/x/sync/errgroup) - Goroutine management
- [prometheus/client_golang](https://github.com/prometheus/client_golang) - Metrics (optional)

## 🔧 Basic Usage

### Message Consumption

```go
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

	// Create RabbitMQ connection
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
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
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
```

## ⚙️ Advanced Configuration

### Connection Options

```go
import "github.com/KevenMarioN/hop/conn"

// With custom connection name
hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/",
	conn.WithConnectionName("my-app"),
)

// With Prometheus metrics
import "github.com/prometheus/client_golang/prometheus"

registry := prometheus.NewRegistry()
hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/",
	conn.WithMetrics(registry),
)
```

### Multiple Consumers

```go
// Consumer 1
err = hopClient.Consume(protocol.Consumer{
	Name:    "consumer-1",
	Queue: protocol.Queue{Name: "queue-1"},
	Exec: func(ctx context.Context, msg amqp091.Delivery) error {
		// Process message
		return nil
	},
})

// Consumer 2
err = hopClient.Consume(protocol.Consumer{
	Name:    "consumer-2",
	Queue: protocol.Queue{Name: "queue-2"},
	Exec: func(ctx context.Context, msg amqp091.Delivery) error {
		// Process message
		return nil
	},
})

// Start all consumers
hopClient.StartConsumers(ctx)
```

### Graceful Shutdown

```go
ctx := context.Background()
hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/")
if err != nil {
	panic(err)
}

// Register consumers...

hopClient.StartConsumers(ctx)

// Wait for safe completion
if err := hopClient.Shutdown(ctx); err != nil {
	log.Error().Err(err).Msg("Shutdown failed")
}
```

## 📚 API Reference

### Client Interface

```go
type Client interface {
	Consume(args protocol.Consumer) error
	StartConsumers(ctx context.Context)
	Shutdown(ctx context.Context) error
	Close() error
}
```

### Opções de Configuração

```go
import "github.com/KevenMarioN/hop/conn"

// Configurações disponíveis:
conn.WithConnectionName("my-app")          // Nome da conexão
conn.WithBackoff(2, 100*time.Millisecond, 30*time.Second) // Backoff
conn.WithTLS(tlsConfig)                   // TLS
conn.WithServiceName("my-service")        // Nome do serviço
conn.WithMetrics(prometheusRegistry)      // Métricas Prometheus
```

### Functions

#### `New(ctx context.Context, url string, opts ...conn.HopOption) (Client, error)`

Creates a new RabbitMQ connection.

**Parameters:**
- `ctx`: Context for lifecycle control
- `url`: RabbitMQ connection URL (e.g., `amqp://user:pass@host:5672/`)
- `opts`: Connection options (optional)

**Returns:**
- `Client`: Hop client interface
- `error`: Error if connection fails

#### `Client.Consume(args protocol.Consumer) error`

Registers a consumer for message consumption.

**Parameters:**
- `args`: Consumer configuration

**Returns:**
- `error`: Error if registration fails

#### `Client.StartConsumers(ctx context.Context)`

Starts all registered consumers.

**Parameters:**
- `ctx`: Context for lifecycle control

#### `Client.Shutdown(ctx context.Context) error`

Safely closes the connection, waiting for all goroutines to complete.

**Parameters:**
- `ctx`: Context for timeout control

**Returns:**
- `error`: Error if shutdown fails

#### `Client.Close() error`

Closes the RabbitMQ connection.

**Returns:**
- `error`: Error if closing fails

### Structures

#### `protocol.Consumer`

Consumer configuration.

```go
type Consumer struct {
	Name      string                 // Consumer name
	Key       string                 // Binding key (for exchanges)
	AutoAck   bool                   // Auto-acknowledge
	NoLocal   bool                   // Don't consume locally published messages
	Exclusive bool                   // Exclusive consumer
	NoWait    bool                   // Don't wait for confirmation
	Headers   map[string]any         // Custom headers
	Queue     Queue                  // Queue configuration
	Exchange  *Exchange              // Exchange configuration (optional)
	Exec      Handler                // Processing function
}
```

#### `protocol.Queue`

Queue configuration.

```go
type Queue struct {
	Durable           bool            // Durable queue
	AutoDelete        bool            // Delete automatically when empty
	Exclusive         bool            // Exclusive to connection
	NoWait            bool            // Don't wait for confirmation
	Name              string          // Queue name
	Headers           map[string]any  // Custom headers
	ShouldCreateQueue bool            // Flag to create queue automatically
}
```

#### `protocol.Exchange`

Exchange configuration.

```go
type Exchange struct {
	Durable    bool            // Durable exchange
	AutoDelete bool            // Delete automatically when empty
	Exclusive  bool            // Exclusive to connection
	NoWait     bool            // Don't wait for confirmation
	Internal   bool            // Internal exchange
	Kind       Kind            // Exchange type (fanout, topic, direct)
	Name       string          // Exchange name
	Headers    map[string]any  // Custom headers
}
```

#### `protocol.Handler`

Message processing function.

```go
type Handler func(ctx context.Context, msg amqp091.Delivery) error
```

#### `protocol.Kind`

Available exchange types.

```go
const (
	Fanout  Kind = "fanout"
	Topic   Kind = "topic"
	Direct  Kind = "direct"
	Default Kind = ""
)
```

## 🛡️ Features

### Auto-Reconnect

The library monitors the connection and automatically reconnects in case of failure.

### Resilience

Implements exponential backoff for reconnections, avoiding server overload.

### Graceful Shutdown

Safely closes connections and goroutines, ensuring all messages in processing are completed.

### Structured Logging

Uses zerolog for structured and performant logging.

### Prometheus Metrics

Collect metrics for consumption, errors, reconnections, and connection duration. Enable with `WithMetrics()`.

### ConsumerBuilder

API fluida para construção de consumers de forma type-safe e imutável:

```go
consumer, err := protocol.NewConsumerBuilder("my-consumer").
    WithQueue(protocol.Queue{Name: "my-queue", Durable: true}).
    WithExchange(&protocol.Exchange{Name: "my-exchange", Kind: protocol.Direct}).
    WithHandler(func(ctx context.Context, msg amqp091.Delivery) error {
        // Processa mensagem
        return nil
    }).
    Build()
```

## 🛡️ Features

### Auto-Reconnect

The library monitors the connection and automatically reconnects in case of failure.

### Resilience

Implements exponential backoff for reconnections, avoiding server overload.

### Graceful Shutdown

Safely closes connections and goroutines, ensuring all messages in processing are completed.

## 🧪 Tests

Run unit tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## 📝 Examples

See the `examples/basic/` directory for complete usage examples.

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [RabbitMQ AMQP Go Client](https://github.com/rabbitmq/amqp091-go)
- [Zerolog](https://github.com/rs/zerolog)
- [Go Sync](https://golang.org/x/sync)
