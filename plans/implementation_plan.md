# Plano de Implementação - Melhorias Prioritárias

## Visão Geral
Este plano detalha as melhorias necessárias para elevar a biblioteca Hop ao padrão de produção, seguindo práticas do mercado Go.

---

## Fase 1: Correção de Race Conditions (CRÍTICO)

### Tarefa 1.1: Corrigir `Manager.Start` - Proteger iteração sobre map
**Arquivo**: `consumer/manager.go:55-68`

**Problema**: Itera sobre `m.consumers` com RLock, mas `recreateConsumer` adquire Lock exclusivo, potencial race condition.

**Solução**: Copiar map antes de iterar.

```go
// Antes
func (m *Manager) Start(ctx context.Context) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    for name, consumer := range m.consumers {
        // ...
    }
}

// Depois
func (m *Manager) Start(ctx context.Context) error {
    m.mu.RLock()
    consumers := make(map[string]*protocol.Consumer, len(m.consumers))
    for name, consumer := range m.consumers {
        consumers[name] = consumer
    }
    m.mu.RUnlock()

    for name, consumer := range consumers {
        // ...
    }
}
```

**Testes**: Verificar se `conn/conn_test.go` cobre este cenário. Se necessário, adicionar teste de concorrência.

---

## Fase 2: Observabilidade (ALTA)

### Tarefa 2.1: Adicionar métricas com `prometheus`
**Novo arquivo**: `metrics/metrics.go`

**Estrutura**:
```go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    MessagesConsumed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hop_messages_consumed_total",
            Help: "Total number of messages consumed",
        },
        []string{"consumer", "queue"},
    )

    ConsumptionErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hop_consumption_errors_total",
            Help: "Total number of consumption errors",
        },
        []string{"consumer", "error_type"},
    )

    Reconnects = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "hop_reconnects_total",
            Help: "Total number of reconnections",
        },
    )

    ConnectionDuration = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "hop_connection_duration_seconds",
            Help: "Duration of current connection in seconds",
        },
    )
)
```

**Integração**:
- Em `consumer/manager.go:101` - incrementar `MessagesConsumed` no sucesso
- Em `consumer/manager.go:102` - incrementar `ConsumptionErrors` no erro
- Em `conn/conn.go:76` - incrementar `Reconnects` após reconexão
- Inicializar gauges em `Connect` e atualizar periodicamente

### Tarefa 2.2: Expor endpoint HTTP para métricas
**Novo arquivo**: `metrics/http.go`

```go
func NewHandler() http.Handler {
    return promhttp.Handler()
}
```

**Integração**: Adicionar opção `WithMetricsHandler` em `conn/options.go`.

---

## Fase 3: Documentação (MÉDIA)

### Tarefa 3.1: Adicionar go doc comments em tipos públicos
**Arquivos**: `protocol/topology.go`, `conn/options.go`, `hop.go`

**Exemplo**:
```go
// Consumer represents a RabbitMQ consumer configuration.
// It defines the queue, exchange, and handler for processing messages.
type Consumer struct {
    // Name is a unique identifier for the consumer.
    Name string
    // Exec is the handler function called for each message.
    Exec Handler
    // ...
}
```

### Tarefa 3.2: Documentar opções de configuração
**Arquivo**: `conn/options.go`

Adicionar comments para cada função option:
```go
// WithConnectionName sets a custom name for the AMQP connection.
// Useful for identifying connections in RabbitMQ management UI.
func WithConnectionName(connectionName string) HopOption {
    // ...
}
```

### Tarefa 3.3: Adicionar exemplos avançados
**Novo arquivo**: `examples/retry_handler/retry_handler.go`

Demonstrar:
- Handler com retry lógico
- Uso de context timeout
- Dead letter queue pattern

---

## Fase 4: Design - Consumer Imutável (MÉDIA)

### Tarefa 4.1: Criar `ConsumerBuilder` para construção fluida
**Novo arquivo**: `protocol/builder.go`

```go
type ConsumerBuilder struct {
    consumer Consumer
}

func NewConsumerBuilder(name string) *ConsumerBuilder {
    return &ConsumerBuilder{
        consumer: Consumer{
            Name: name,
            Queue: Queue{
                Durable: true, // default
            },
        },
    }
}

func (b *ConsumerBuilder) WithQueue(queue Queue) *ConsumerBuilder {
    b.consumer.Queue = queue
    return b
}

func (b *ConsumerBuilder) WithExchange(exchange *Exchange) *ConsumerBuilder {
    b.consumer.Exchange = exchange
    return b
}

func (b *ConsumerBuilder) WithHandler(handler Handler) *ConsumerBuilder {
    b.consumer.Exec = handler
    return b
}

func (b *ConsumerBuilder) Build() (*Consumer, error) {
    if err := b.consumer.Validate(); err != nil {
        return nil, err
    }
    return &b.consumer, nil
}
```

**Migração**: Manter `Consumer` atual por compatibilidade, adicionar `Builder` como alternativa.

### Tarefa 4.2: Remover métodos mutators (breaking change)
**Arquivo**: `protocol/topology.go`

Remover:
- `Msg()` (linha 74-76)
- `Handler()` (linha 82-84)

**Justificativa**: Consumer deve ser imutável após criação. Apenas `Manager` modifica internamente.

---

## Fase 5: Performance (BAIXA)

### Tarefa 5.1: Reduzir logging em produção
**Arquivo**: `protocol/topology.go:87`

```go
// Antes
log.Debug().Str(...).Msg("received message")

// Depois
if log.DebugEnabled() {
    log.Debug().Str(...).Msg("received message")
}
```

**Alternativa**: Remover log de mensagem individual, confiar em métricas.

### Tarefa 5.2: Adicionar connection pooling (opcional)
**Novo arquivo**: `conn/pool.go`

Implementar `ConnectionPool` para alta carga:
- Pool de conexões
- Round-robin ou aleatório
- Health check

**Decisão**: Avaliar necessidade após métricas. Pode ser overkill para maioria dos casos.

---

## Fase 6: Error Handling (MÉDIA)

### Tarefa 6.1: Documentar `Publish` como não implementado
**Arquivo**: `conn/conn.go:102-104`

```go
// Publish publishes a message to an exchange.
// NOTE: Not implemented yet. Contributions welcome!
func (c *hop) Publish(ctx context.Context, exchange, key string, body []byte) error {
    return ErrNotImplemented
}
```

### Tarefa 6.2: Revisar wrapping de erros
Verificar todos os `fmt.Errorf` e garantir uso de `%w` para erros wrapped.

---

## Ordem de Execução Recomendada

1. **Fase 1** (Race condition) - CRÍTICO, blocker para produção
2. **Fase 2** (Métricas) - ALTA, importante para observabilidade
3. **Fase 6** (Error handling) - MÉDIA, quick win
4. **Fase 3** (Documentação) - MÉDIA, melhora DX
5. **Fase 4** (ConsumerBuilder) - MÉDIA, breaking change menor
6. **Fase 5** (Performance) - BAIXA, otimização

---

## Checklist de Validação

- [ ] Todos os testes passam (`go test ./...`)
- [ ] Sem race conditions (`go test -race ./...`)
- [ ] Linter passa (`golangci-lint run`)
- [ ] Benchmarks adicionados para métricas
- [ ] README atualizado com novos exemplos
- [ ] CHANGELOG atualizado
- [ ] Versionamento semântico (major/minor/patch)

---

## Riscos

1. **Fase 1**: Se não corrigida, pode causar panics em produção
2. **Fase 4**: Breaking change - requer major version bump
3. **Fase 2**: Dependência externa adicionada (prometheus)

---

## Estimativa de Esforço

- Fase 1: 2-4 horas
- Fase 2: 4-8 horas
- Fase 3: 2-4 horas
- Fase 4: 4-6 horas
- Fase 5: 2-4 horas
- Fase 6: 1-2 horas

**Total**: 15-28 horas de desenvolvimento

---

## Notas

- Manter compatibilidade backwards sempre que possível
- Usar feature flags para mudanças breaking
- Atualizar exemplos existentes gradualmente
- Considerar release gradual (alpha → beta → stable)
