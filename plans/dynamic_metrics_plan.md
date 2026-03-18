# Plano: Sistema de Métricas Dinâmico

## Problema Atual

A implementação atual de métricas está **fortemente acoplada ao Prometheus**:
- Variáveis globais em `metrics/metrics.go` são do tipo Prometheus específico
- `Manager` chama diretamente `metrics.ActiveConsumers.Set()`, `metrics.MessagesConsumed.WithLabelValues()`, etc.
- `WithMetrics()` só aceita `prometheus.Registerer`
- Impossível usar OpenTelemetry, StatsD, ou múltiplos sistemas simultaneamente

## Objetivo

Criar um sistema de métricas **aberto e plugável** que permita:
1. Usar diferentes backends (Prometheus, OpenTelemetry, etc.)
2. Usar múltiplos backends simultaneamente
3. Manter compatibilidade com código existente
4. Facilidade de extensão para novos tipos de métricas

## Arquitetura Proposta

### 1. Interface `MetricsCollector`

```go
// metrics/collector.go
package metrics

// MetricsCollector define a interface para coleta de métricas.
// Implementações podem ser Prometheus, OpenTelemetry, ou qualquer outro sistema.
type MetricsCollector interface {
    // Counter retorna uma métrica do tipo Counter com as labels fornecidas.
    Counter(name string, labels ...string) Counter
    // Gauge retorna uma métrica do tipo Gauge com as labels fornecidas.
    Gauge(name string, labels ...string) Gauge
    // Registerer retorna o registerer subjacente para registro de métricas.
    // Pode ser nil se o collector não suportar registro explícito.
    Registerer() any
}

// Counter representa uma métrica counter.
type Counter interface {
    Inc()
    Add(float64)
}

// Gauge representa uma métrica gauge.
type Gauge interface {
    Set(float64)
    Inc()
    Dec()
    Add(float64)
}
```

### 2. Implementação Prometheus

```go
// metrics/prometheus.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

// PrometheusCollector implementa MetricsCollector usando Prometheus.
type PrometheusCollector struct {
    registry prometheus.Registerer
    // Cache de métricas para evitar lookup repetido
    counters  map[string]prometheus.Counter
    gauges    map[string]prometheus.Gauge
}

func NewPrometheusCollector(registry prometheus.Registerer) *PrometheusCollector {
    p := &PrometheusCollector{
        registry:  registry,
        counters:  make(map[string]prometheus.Counter),
        gauges:    make(map[string]prometheus.Gauge),
    }
    // Registrar métricas padrão do Hop
    p.registerDefaultMetrics()
    return p
}

func (p *PrometheusCollector) Counter(name string, labels ...string) Counter {
    key := name + "|" + strings.Join(labels, "|")
    if c, ok := p.counters[key]; ok {
        return &prometheusCounter{c}
    }
    // Criar nova métrica
    vec := prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: name, Help: name},
        labels,
    )
    p.registry.MustRegister(vec)
    c := vec.WithLabelValues(labels...)
    p.counters[key] = c
    return &prometheusCounter{c}
}

func (p *PrometheusCollector) Gauge(name string, labels ...string) Gauge {
    key := name + "|" + strings.Join(labels, "|")
    if g, ok := p.gauges[key]; ok {
        return &prometheusGauge{g}
    }
    vec := prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: name, Help: name},
        labels,
    )
    p.registry.MustRegister(vec)
    g := vec.WithLabelValues(labels...)
    p.gauges[key] = g
    return &prometheusGauge{g}
}

func (p *PrometheusCollector) Registerer() any {
    return p.registry
}

// Implementações wrapper
type prometheusCounter struct{ c prometheus.Counter }
func (pc *prometheusCounter) Inc() { pc.c.Inc() }
func (pc *prometheusCounter) Add(v float64) { pc.c.Add(v) }

type prometheusGauge struct{ g prometheus.Gauge }
func (pg *prometheusGauge) Set(v float64) { pg.g.Set(v) }
func (pg *prometheusGauge) Inc() { pg.g.Inc() }
func (pg *prometheusGauge) Dec() { pg.g.Dec() }
func (pg *prometheusGauge) Add(v float64) { pg.g.Add(v) }

// Métricas padrão do Hop
func (p *PrometheusCollector) registerDefaultMetrics() {
    // MessagesConsumed: counter com labels consumer, queue
    p.counters["hop_messages_consumed_total"] = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hop_messages_consumed_total",
            Help: "Total number of messages consumed",
        },
        []string{"consumer", "queue"},
    ).WithLabelValues() // Sem labels por padrão

    // ConsumptionErrors: counter com labels consumer, error_type
    p.counters["hop_consumption_errors_total"] = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "hop_consumption_errors_total",
            Help: "Total number of consumption errors",
        },
        []string{"consumer", "error_type"},
    ).WithLabelValues()

    // Reconnects: counter sem labels
    p.counters["hop_reconnects_total"] = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "hop_reconnects_total",
            Help: "Total number of reconnections",
        },
    )

    // ConnectionDuration: gauge sem labels
    p.gauges["hop_connection_duration_seconds"] = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "hop_connection_duration_seconds",
            Help: "Duration of current connection in seconds",
        },
    )

    // ActiveConsumers: gauge sem labels
    p.gauges["hop_active_consumers"] = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "hop_active_consumers",
            Help: "Number of active consumers",
        },
    )
}
```

### 3. MultiCollector (Coleta Múltipla)

```go
// metrics/multi.go
package metrics

// MultiCollector permite usar múltiplos collectors simultaneamente.
type MultiCollector struct {
    collectors []MetricsCollector
}

func NewMultiCollector(collectors ...MetricsCollector) *MultiCollector {
    return &MultiCollector{collectors: collectors}
}

func (mc *MultiCollector) Counter(name string, labels ...string) Counter {
    return &multiCounter{
        collectors: mc.collectors,
        name:       name,
        labels:     labels,
    }
}

func (mc *MultiCollector) Gauge(name string, labels ...string) Gauge {
    return &multiGauge{
        collectors: mc.collectors,
        name:       name,
        labels:     labels,
    }
}

func (mc *MultiCollector) Registerer() any {
    // Retorna o primeiro registerer disponível (para compatibilidade)
    for _, c := range mc.collectors {
        if r := c.Registerer(); r != nil {
            return r
        }
    }
    return nil
}

type multiCounter struct {
    collectors []MetricsCollector
    name       string
    labels     []string
}

func (mc *multiCounter) Inc() {
    for _, c := range mc.collectors {
        c.Counter(mc.name, mc.labels...).Inc()
    }
}

func (mc *multiCounter) Add(v float64) {
    for _, c := range mc.collectors {
        c.Counter(mc.name, mc.labels...).Add(v)
    }
}

type multiGauge struct {
    collectors []MetricsCollector
    name       string
    labels     []string
}

func (mc *multiGauge) Set(v float64) {
    for _, c := range mc.collectors {
        c.Gauge(mc.name, mc.labels...).Set(v)
    }
}

func (mc *multiGauge) Inc() {
    for _, c := range mc.collectors {
        c.Gauge(mc.name, mc.labels...).Inc()
    }
}

func (mc *multiGauge) Dec() {
    for _, c := range mc.collectors {
        c.Gauge(mc.name, mc.labels...).Dec()
    }
}

func (mc *multiGauge) Add(v float64) {
    for _, c := range mc.collectors {
        c.Gauge(mc.name, mc.labels...).Add(v)
    }
}
```

### 4. NopCollector (No-op para quando métricas estão desabilitadas)

```go
// metrics/nop.go
package metrics

type nopCounter struct{}
func (nc *nopCounter) Inc() {}
func (nc *nopCounter) Add(v float64) {}

type nopGauge struct{}
func (ng *nopGauge) Set(v float64) {}
func (ng *nopGauge) Inc() {}
func (ng *nopGauge) Dec() {}
func (ng *nopGauge) Add(v float64) {}

var NopCollector = &nopCollector{}

type nopCollector struct{}

func (nc *nopCollector) Counter(name string, labels ...string) Counter {
    return &nopCounter{}
}

func (nc *nopCollector) Gauge(name string, labels ...string) Gauge {
    return &nopGauge{}
}

func (nc *nopCollector) Registerer() any {
    return nil
}
```

### 5. Atualizar `Manager` para usar a interface

```go
// consumer/manager.go (alterações)
type Manager struct {
    conn      *amqp.Connection
    consumers map[string]*protocol.Consumer
    mu        sync.RWMutex
    wg        errgroup.Group
    reconnect *sync.Cond
    metrics   metrics.MetricsCollector  // ← Mudança: usar interface
    startTime time.Time
}

func NewManager(conn *amqp.Connection, collector metrics.MetricsCollector) *Manager {
    m := &Manager{
        conn:      conn,
        consumers: make(map[string]*protocol.Consumer),
        reconnect: sync.NewCond(&sync.Mutex{}),
        startTime: time.Now(),
        metrics:   collector,  // ← Usar collector
    }
    return m
}

// Em Register():
if m.metrics != nil {
    m.metrics.Gauge("hop_active_consumers").Set(float64(len(m.consumers)))
}

// Em StartConsumer (sucesso):
if m.metrics != nil {
    m.metrics.Counter("hop_messages_consumed_total", consumer.Name, consumer.Queue.Name).Inc()
}

// Em StartConsumer (erro):
if m.metrics != nil {
    m.metrics.Counter("hop_consumption_errors_total", consumer.Name, "handler_error").Inc()
}

// Em Wait():
if m.metrics != nil {
    duration := time.Since(m.startTime).Seconds()
    m.metrics.Gauge("hop_connection_duration_seconds").Set(duration)
}

// Em NotifyReconnected():
if m.metrics != nil {
    m.metrics.Counter("hop_reconnects_total").Inc()
}
```

### 6. Atualizar `conn/options.go`

```go
// conn/options.go (alterações)
import (
    "github.com/KevenMarioN/hop/metrics"
    "github.com/prometheus/client_golang/prometheus"
)

// WithMetrics accepts any MetricsCollector implementation.
// This enables using Prometheus, OpenTelemetry, or multiple backends simultaneously.
func WithMetrics(collector metrics.MetricsCollector) HopOption {
    return func(h *hop) {
        if collector != nil {
            // Se for um PrometheusCollector, registrar métricas padrão
            // Se for MultiCollector, cada collector é responsável por seu registro
            h.consumerMgr = consumer.NewManager(h.conn, collector)
        }
    }
}

// WithPrometheusMetrics is a convenience wrapper for backward compatibility.
// It creates a PrometheusCollector with the given registry.
func WithPrometheusMetrics(registry prometheus.Registerer) HopOption {
    return func(h *hop) {
        if registry != nil {
            collector := metrics.NewPrometheusCollector(registry)
            h.consumerMgr = consumer.NewManager(h.conn, collector)
        }
    }
}
```

### 7. Manter Compatibilidade com Código Existente

Para não quebrar código existente, podemos:

**Opção A**: Manter `WithMetrics(registry prometheus.Registerer)` e adicionar `WithCollector(collector MetricsCollector)`
- Código antigo continua funcionando
- Novo código usa a interface

**Opção B**: Mudar `WithMetrics` para aceitar `any` e fazer type-assertion internamente
- Mais simples, mas menos type-safe

**Recomendação**: **Opção A** - mais clara e mantém compatibilidade total.

### 8. Atualizar `metrics/http.go`

```go
// metrics/http.go (alterações)
package metrics

import (
    "net/http"
)

// Handler returns an HTTP handler for the given collector.
// For Prometheus, it uses the registry from the collector.
// For multi-collector, it uses the first Prometheus registry found.
func Handler(collector MetricsCollector) http.Handler {
    // Tentar obter registry do collector
    if reg, ok := collector.Registerer().(prometheus.Registerer); ok {
        return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
    }
    // Fallback: usar default registry
    return promhttp.Handler()
}
```

### 9. Atualizar `conn/conn.go`

```go
// conn/conn.go (alterações)
func Connect(ctx context.Context, url string, opts ...HopOption) (*hop, error) {
    c := &hop{
        // ... outros campos
    }

    for _, opt := range opts {
        opt(c)  // As opções agora criam o consumerMgr com o collector apropriado
    }

    // Se consumerMgr não foi criado por nenhuma opção, criar com NopCollector
    if c.consumerMgr == nil {
        c.consumerMgr = consumer.NewManager(c.conn, metrics.NopCollector)
    }

    // ... resto do código
}
```

### 10. Atualizar Documentação

- Adicionar seção "Custom Metrics Backend" no README
- Mostrar exemplo com OpenTelemetry (se disponível)
- Mostrar exemplo com MultiCollector
- Documentar a interface `MetricsCollector`

## Fases de Implementação

### Fase 1: Criar Interface e Wrappers
1. Criar `metrics/collector.go` com interface `MetricsCollector`, `Counter`, `Gauge`
2. Criar `metrics/nop.go` com implementação no-op
3. Criar `metrics/prometheus.go` com `PrometheusCollector`
4. Atualizar `metrics/http.go` para usar collector

### Fase 2: Atualizar Manager
1. Modificar `Manager` para aceitar `MetricsCollector` em vez de `prometheus.Registerer`
2. Substituir todas as chamadas diretas a `metrics.X` por `m.metrics.X`
3. Adicionar verificação `if m.metrics != nil`

### Fase 3: Atualizar Options e Conn
1. Modificar `WithMetrics` para aceitar `MetricsCollector`
2. Adicionar `WithPrometheusMetrics` para compatibilidade
3. Atualizar `Connect` para criar `consumerMgr` com `NopCollector` se não especificado

### Fase 4: Atualizar Documentação
1. README.md e README.en.md com exemplos de custom metrics
2. Adicionar exemplo OpenTelemetry (conceitual)
3. Adicionar exemplo MultiCollector

### Fase 5: Testes
1. Criar testes para `PrometheusCollector`
2. Criar testes para `MultiCollector`
3. Criar testes para `NopCollector`
4. Atualizar testes existentes que usam métricas

## Vantagens da Abordagem

1. **Desacoplamento total**: O núcleo do Hop não depende de nenhum sistema de métricas específico
2. **Extensibilidade**: Fácil adicionar novos backends (OpenTelemetry, StatsD, DogStatsD, etc.)
3. **Múltiplos backends**: `MultiCollector` permite enviar métricas para vários sistemas simultaneamente
4. **Compatibilidade**: Código existente funciona com `WithPrometheusMetrics`
5. **Testabilidade**: Fácil mockar `MetricsCollector` em testes
6. **Performance**: Cache de métricas no `PrometheusCollector` evita lookup repetido

## Considerações

1. **Cache**: O `PrometheusCollector` cacheia métricas para evitar criar novo objeto a cada chamada. Isso é importante para performance.
2. **Thread-safety**: A interface não exige thread-safety, mas as implementações devem ser thread-safe se usadas concorrentemente. Prometheus já é thread-safe.
3. **Label values**: A interface usa `labels ...string` em vez de `[]string` para flexibilidade.
4. **Registerer() any**: Retorna `any` para suportar diferentes tipos de registerer (Prometheus, OTel, etc.). Quem usa pode fazer type-assertion se necessário.

## Exemplo de Uso com OpenTelemetry (futuro)

```go
import "github.com/KevenMarioN/hop/metrics"

// Criar collector OpenTelemetry (quando houver implementação)
otelCollector := metrics.NewOpenTelemetryCollector(otelMeter)
multi := metrics.NewMultiCollector(
    otelCollector,
    metrics.NewPrometheusCollector(promRegistry),
)
hopClient, _ := hop.New(ctx, url, conn.WithMetrics(multi))
```

## Conclusão

Este plano transforma o sistema de métricas de acoplado para desacoplado, permitindo:
- Uso de qualquer sistema de métricas
- Múltiplos sistemas simultaneamente
- Manutenção da compatibilidade com código existente
- Fácil extensão no futuro

O custo é um pequeno overhead de indireção (interface) que é insignificante comparado ao benefício de flexibilidade.
