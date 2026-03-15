# Hop - Biblioteca de Conexão RabbitMQ para Go

[![Go Report Card](https://goreportcard.com/badge/github.com/KevenMarioN/hop)](https://goreportcard.com/report/github.com/KevenMarioN/hop)
[![GoDoc](https://godoc.org/github.com/KevenMarioN/hop?status.svg)](https://godoc.org/github.com/KevenMarioN/hop)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Hop é uma biblioteca Go simples e resiliente para conexão com RabbitMQ, com suporte a auto-reconnect, graceful shutdown, métricas Prometheus e consumo de mensagens.

## 🚀 Instalação

```bash
go get github.com/KevenMarioN/hop
```

## 📦 Dependências

- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - Cliente AMQP oficial
- [zerolog](https://github.com/rs/zerolog) - Logging estruturado
- [errgroup](https://golang.org/x/sync/errgroup) - Gerenciamento de goroutines
- [prometheus/client_golang](https://github.com/prometheus/client_golang) - Métricas (opcional)

## 🔧 Uso Básico

### Consumo de Mensagens

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

	// Cria conexão com RabbitMQ
	hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
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
		fmt.Printf("Failed to register consumer: %v", err)
		return
	}

	// Inicia consumers
	hopClient.StartConsumers(ctx)

	// Aguarda conclusão
	if err := hopClient.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to shutdown: %v", err)
		return
	}
}
```

## ⚙️ Configuração Avançada

### Opções de Conexão

```go
import "github.com/KevenMarioN/hop/conn"

// Com nome de conexão personalizado
hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/",
	conn.WithConnectionName("meu-app"),
)
```

### Múltiplos Consumers

```go
// Consumer 1
err = hopClient.Consume(protocol.Consumer{
	Name:    "consumer-1",
	Queue: protocol.Queue{Name: "queue-1"},
	Exec: func(ctx context.Context, msg amqp091.Delivery) error {
		// Processa mensagem
		return nil
	},
})

// Consumer 2
err = hopClient.Consume(protocol.Consumer{
	Name:    "consumer-2",
	Queue: protocol.Queue{Name: "queue-2"},
	Exec: func(ctx context.Context, msg amqp091.Delivery) error {
		// Processa mensagem
		return nil
	},
})

// Inicia todos os consumers
hopClient.StartConsumers(ctx)
```

### Graceful Shutdown

```go
ctx := context.Background()
hopClient, err := hop.New(ctx, "amqp://user:pass@localhost:5672/")
if err != nil {
	panic(err)
}

// Registra consumers...

hopClient.StartConsumers(ctx)

// Aguarda conclusão de forma segura
if err := hopClient.Shutdown(ctx); err != nil {
	log.Error().Err(err).Msg("Falha no shutdown")
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

### Funções

#### `New(ctx context.Context, url string, opts ...conn.HopOption) (Client, error)`

Cria uma nova conexão com RabbitMQ.

**Parâmetros:**
- `ctx`: Contexto para controle de vida útil
- `url`: URL de conexão RabbitMQ (ex: `amqp://user:pass@host:5672/`)
- `opts`: Opções de conexão (opcional)

**Retorno:**
- `Client`: Interface do cliente Hop
- `error`: Erro se a conexão falhar

#### `Client.Consume(args protocol.Consumer) error`

Registra um consumer para consumo de mensagens.

**Parâmetros:**
- `args`: Configuração do consumer

**Retorno:**
- `error`: Erro se o registro falhar

#### `Client.StartConsumers(ctx context.Context)`

Inicia todos os consumers registrados.

**Parâmetros:**
- `ctx`: Contexto para controle de vida útil

#### `Client.Shutdown(ctx context.Context) error`

Encerra a conexão de forma segura, aguardando conclusão de todas as goroutines.

**Parâmetros:**
- `ctx`: Contexto para controle de timeout

**Retorno:**
- `error`: Erro se o shutdown falhar

#### `Client.Close() error`

Fecha a conexão RabbitMQ.

**Retorno:**
- `error`: Erro se o fechamento falhar

### Estruturas

#### `protocol.Consumer`

Configuração de um consumer.

```go
type Consumer struct {
	Name      string                 // Nome do consumer
	Key       string                 // Chave de binding (para exchanges)
	AutoAck   bool                   // Auto-acknowledge
	NoLocal   bool                   // Não consumir mensagens publicadas localmente
	Exclusive bool                   // Consumer exclusivo
	NoWait    bool                   // Não aguardar confirmação
	Headers   map[string]any         // Headers personalizados
	Queue     Queue                  // Configuração da fila
	Exchange  *Exchange              // Configuração do exchange (opcional)
	Exec      Handler                // Função de processamento
}
```

#### `protocol.Queue`

Configuração da fila.

```go
type Queue struct {
	Durable           bool            // Fila durável
	AutoDelete        bool            // Deletar automaticamente quando vazia
	Exclusive         bool            // Fila exclusiva para conexão
	NoWait            bool            // Não aguardar confirmação
	Name              string          // Nome da fila
	Headers           map[string]any  // Headers personalizados
	ShouldCreateQueue bool            // Flag para criar fila automaticamente
}
```

#### `protocol.Exchange`

Configuração do exchange.

```go
type Exchange struct {
	Durable    bool            // Exchange durável
	AutoDelete bool            // Deletar automaticamente quando vazia
	Exclusive  bool            // Exchange exclusiva para conexão
	NoWait     bool            // Não aguardar confirmação
	Internal   bool            // Exchange interna
	Kind       Kind            // Tipo de exchange (fanout, topic, direct)
	Name       string          // Nome do exchange
	Headers    map[string]any  // Headers personalizados
}
```

#### `protocol.Handler`

Função de processamento de mensagens.

```go
type Handler func(ctx context.Context, msg amqp091.Delivery) error
```

#### `protocol.Kind`

Tipos de exchange disponíveis.

```go
const (
	Fanout  Kind = "fanout"
	Topic   Kind = "topic"
	Direct  Kind = "direct"
	Default Kind = ""
)
```

## 🛡️ Recursos

### Auto-Reconnect

A biblioteca monitora a conexão e reconecta automaticamente em caso de falha.

### Resilience

Implementa exponential backoff para reconexões, evitando sobrecarga do servidor.

### Graceful Shutdown

Encerra conexões e goroutines de forma segura, garantindo que todas as mensagens em processamento sejam concluídas.

### Logging Estruturado

Utiliza zerolog para logging estruturado e performático.

### Métricas Prometheus

Colete métricas de consumo, erros, reconexões e duração da conexão. Ative com `WithMetrics()`.

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

## 🧪 Testes

Execute os testes unitários:

```bash
go test ./...
```

Execute os testes com cobertura:

```bash
go test -cover ./...
```

## 📝 Exemplos

Veja o diretório `examples/basic/` para exemplos completos de uso.

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor, siga estas diretrizes:

1. Fork o repositório
2. Crie uma branch para sua feature (`git checkout -b feature/amazing-feature`)
3. Commit suas alterações (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🙏 Agradecimentos

- [RabbitMQ AMQP Go Client](https://github.com/rabbitmq/amqp091-go)
- [Zerolog](https://github.com/rs/zerolog)
- [Go Sync](https://golang.org/x/sync)
