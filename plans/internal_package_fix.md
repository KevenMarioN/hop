# Plano de Correção: Expor Tipos Públicos da Biblioteca Hop

## Problema

O Go **proíbe** que código externo importe pacotes `internal/`. Quando um usuário usar sua biblioteca via `go get github.com/KevenMarioN/hop`, ele **não conseguirá** importar:

```go
// ❌ ERRO - Não funciona para usuários externos
import "github.com/KevenMarioN/hop/internal/conn"
import "github.com/KevenMarioN/hop/internal/metrics"
import "github.com/KevenMarioN/hop/internal/protocol"
```

O compilador Go rejeitará com erro:
```
use of internal package github.com/KevenMarioN/hop/internal/xxx not allowed
```

## Análise do Código Atual

### Arquivos na Raiz (Públicos) ✅
| Arquivo | Tipos Expostos |
|---------|---------------|
| `hop.go` | `Client`, `New()` |
| `consumer.go` | `Consumer`, `Handler` (aliases) |
| `exchange.go` | `Exchange`, `Kind` (aliases + constantes) |
| `queue.go` | `Queue` (alias) |
| `message.go` | `Message` (alias) |

### Arquivos Internal (Privados) ❌
| Pacote | Tipos que precisam ser expostos |
|--------|--------------------------------|
| `internal/conn` | `HopOption`, `WithBackoff()`, `WithConnectionName()`, `WithTLS()`, `WithServiceName()`, `WithMetrics()`, `WithPrometheusMetrics()` |
| `internal/metrics` | `MetricsCollector`, `Counter`, `Gauge`, `Histogram`, `NewPrometheusCollector()`, `NewOpenTelemetryCollector()`, `NewMultiCollector()`, `NopCollector` |
| `internal/protocol` | Tipos já expostos via aliases ✅ |

## Correções Necessárias

### 1. Criar `options.go` na raiz

Expor todas as opções funcionais públicas:

```go
// options.go
package hop

import (
    "crypto/tls"
    "time"

    "github.com/KevenMarioN/hop/internal/conn"
    "github.com/prometheus/client_golang/prometheus"
)

// HopOption configures a hop connection.
type HopOption = conn.HopOption

// WithConnectionName sets a custom name for the AMQP connection.
func WithConnectionName(name string) HopOption {
    return conn.WithConnectionName(name)
}

// WithBackoff configures the reconnection backoff strategy.
func WithBackoff(multiplier float64, initialDelay, maxDelay time.Duration) HopOption {
    return conn.WithBackoff(multiplier, initialDelay, maxDelay)
}

// WithTLS enables TLS encryption for the AMQP connection.
func WithTLS(cfg *tls.Config) HopOption {
    return conn.WithTLS(cfg)
}

// WithServiceName sets the service_name property on the AMQP connection.
func WithServiceName(name string) HopOption {
    return conn.WithServiceName(name)
}

// WithMetrics enables metrics collection using the provided collector.
func WithMetrics(collector MetricsCollector) HopOption {
    return conn.WithMetrics(collector)
}

// WithPrometheusMetrics creates a Prometheus collector with the given registry.
func WithPrometheusMetrics(registry prometheus.Registerer) HopOption {
    return conn.WithPrometheusMetrics(registry)
}
```

### 2. Criar `metrics.go` na raiz

Expor interfaces e construtores de métricas:

```go
// metrics.go
package hop

import (
    "github.com/KevenMarioN/hop/internal/metrics"
    "github.com/prometheus/client_golang/prometheus"
)

// MetricsCollector defines the interface for metrics collection.
type MetricsCollector = metrics.MetricsCollector

// Counter represents a counter metric.
type Counter = metrics.Counter

// Gauge represents a gauge metric.
type Gauge = metrics.Gauge

// Histogram represents a histogram metric.
type Histogram = metrics.Histogram

// NewPrometheusCollector creates a Prometheus-backed metrics collector.
func NewPrometheusCollector(registry prometheus.Registerer) MetricsCollector {
    return metrics.NewPrometheusCollector(registry)
}

// NewOpenTelemetryCollector creates an OpenTelemetry-backed metrics collector.
func NewOpenTelemetryCollector(serviceName string) MetricsCollector {
    return metrics.NewOpenTelemetryCollector(serviceName)
}

// NewMultiCollector combines multiple collectors into one.
func NewMultiCollector(collectors ...MetricsCollector) MetricsCollector {
    return metrics.NewMultiCollector(collectors...)
}

// NopCollector is a no-op collector that discards all metrics.
var NopCollector = metrics.NopCollector
```

### 3. Atualizar `hop.go`

A assinatura de `New()` deve usar o tipo público:

```go
// Alterar de:
func New(ctx context.Context, url string, opts ...conn.HopOption) (Client, error)

// Para:
func New(ctx context.Context, url string, opts ...HopOption) (Client, error)
```

### 4. Atualizar exemplo `cmd/consumer_example/main.go`

O exemplo deve usar apenas a API pública:

```go
package main

import (
    "context"
    "errors"
    "time"

    "github.com/KevenMarioN/hop"
    "github.com/rs/zerolog/log"
)

func main() {
    ctx := context.Background()

    // ✅ Agora usa apenas API pública
    clientHop, err := hop.New(ctx, "amqp://admin:admin@localhost:5672/",
        hop.WithBackoff(2, time.Second*1, time.Minute*1),
        hop.WithMetrics(hop.NewOpenTelemetryCollector("hop")),
    )
    if err != nil {
        log.Error().Err(err).Msg("failed start connection hop")
        return
    }

    if err := clientHop.Consume(hop.Consumer{
        Name:    "example-hop-dollar",
        AutoAck: false,
        Queue: hop.Queue{  // ✅ Usa hop.Queue em vez de protocol.Queue
            Durable:           true,
            Name:              "example.queue",
            ShouldCreateQueue: true,
        },
        Exec: func(ctx context.Context, msg hop.Message) error {
            defer msg.Success()
            log.Info().Msg(string(msg.Body))
            return nil
        },
    }); err != nil {
        log.Error().Err(err).Msg("failed consume")
    }

    clientHop.StartConsumers(ctx)
    clientHop.Shutdown(ctx)
}
```

## Diagrama de Estrutura

```mermaid
graph TB
    subgraph Público - API Estável
        H[hop.go<br/>Client + New]
        O[options.go<br/>HopOption + WithXxx]
        M[metrics.go<br/>MetricsCollector]
        C[consumer.go<br/>Consumer + Handler]
        E[exchange.go<br/>Exchange + Kind]
        Q[queue.go<br/>Queue]
        MSG[message.go<br/>Message]
    end

    subgraph Internal - Implementação
        IC[conn/<br/>Conexão + Reconnect]
        IM[metrics/<br/>Prometheus + OTel]
        IP[protocol/<br/>Tipos AMQP]
        IRES[resilience/<br/>Backoff]
        ICS[consumer/<br/>Manager]
    end

    H --> IC
    O --> IC
    M --> IM
    C --> IP
    E --> IP
    Q --> IP
    MSG --> IP
```

## Checklist de Implementação

- [ ] Criar `options.go` na raiz com aliases e wrappers para `HopOption`
- [ ] Criar `metrics.go` na raiz com aliases e construtores públicos
- [ ] Atualizar `hop.go` para usar `HopOption` público
- [ ] Atualizar `cmd/consumer_example/main.go` para usar apenas API pública
- [ ] Atualizar exemplos em `examples/` para usar API pública
- [ ] Verificar se todos os testes ainda passam
- [ ] Atualizar documentação se necessário

## Impacto nos Usuários

### Antes (Não funciona externamente)
```go
import (
    "github.com/KevenMarioN/hop"
    "github.com/KevenMarioN/hop/internal/conn"      // ❌
    "github.com/KevenMarioN/hop/internal/metrics"    // ❌
    "github.com/KevenMarioN/hop/internal/protocol"   // ❌
)

hop.New(ctx, url, conn.WithBackoff(...), conn.WithMetrics(...))
```

### Depois (API Pública)
```go
import "github.com/KevenMarioN/hop"

hop.New(ctx, url, 
    hop.WithBackoff(...),
    hop.WithMetrics(hop.NewOpenTelemetryCollector("my-service")),
)
```

## Notas

1. **Aliases vs Wrappers**: Use type aliases (`=`) para tipos e wrappers para funções
2. **Compatibilidade**: A API pública permanece estável mesmo se `internal/` mudar
3. **Documentação**: Adicione godoc comments em todos os símbolos públicos