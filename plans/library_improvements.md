# Plano de Melhorias para a Biblioteca Hop

## Análise Atual

A biblioteca Hop é uma implementação de cliente RabbitMQ em Go com as seguintes funcionalidades:
- Conexão com RabbitMQ com auto-reconnect
- Consumo de mensagens com handlers customizados
- Publicação de mensagens (ainda não implementada)
- Resilience com exponential backoff

## Problemas Identificados

### 1. Estruturais
- **Goroutine leak**: `startConsumer` chama recursivamente a si mesmo, criando goroutines sem controle
- **Context não propagado**: O contexto não é verificado durante o processamento de mensagens
- **Canal fechado**: Não há tratamento adequado do fechamento do canal de mensagens
- **Publicação não implementada**: A função `Publish` retorna erro "don't implemented"

### 2. Funcionalidades Faltantes
- **Publicação de mensagens**: Implementar `Publish` com exchange e routing key
- **Exchange management**: Declarar exchanges automaticamente
- **Dead Letter Queue**: Suporte a DLQ para mensagens rejeitadas
- **Retry logic**: Retry automático para mensagens que falham
- **Metrics e observability**: Métricas de consumo e publicação
- **Graceful shutdown**: Fechar conexões e canais corretamente

### 3. Documentação e Testes
- **Documentação**: Faltam exemplos e documentação completa
- **Testes unitários**: Sem testes para as funcionalidades
- **Testes de integração**: Sem testes com RabbitMQ real

## Plano de Melhorias

### 1. Melhorias Estruturais

#### 1.1 Corrigir Goroutine Leak
```go
func (c *hop) startConsumer(ctx context.Context, name string, consumer protocol.Consumer) {
    // ... código atual ...
    // REMOVER: c.startConsumer(ctx, name, consumer) - causa leak
    // SUBSTITUIR POR: sinalizar reconexão via canal
}
```

#### 1.2 Implementar Graceful Shutdown
- Adicionar `Shutdown(ctx context.Context)` ao `Client` interface
- Fechar canais e conexões corretamente
- Aguardar conclusão de goroutines

#### 1.3 Melhorar Tratamento de Erros
- Adicionar `errors.Is` para erros específicos
- Log estruturado com contexto
- Retry com backoff configurável

### 2. Funcionalidades Novas

#### 2.1 Publicação de Mensagens
```go
func (c *hop) Publish(ctx context.Context, exchange, key string, body []byte, opts ...PublishOption) error
```

#### 2.2 Exchange Management
- Declaração automática de exchanges
- Suporte a diferentes tipos de exchange (direct, fanout, topic, headers)

#### 2.3 Dead Letter Queue
- Configuração de DLQ para mensagens rejeitadas
- TTL e max-length para filas

#### 2.4 Retry Logic
- Retry automático para mensagens que falham
- Configuração de max retries e delay

#### 2.5 Metrics e Observability
- Métricas de mensagens consumidas/publicadas
- Métricas de erros e reconexões
- Integração com Prometheus/OpenTelemetry

### 3. Documentação e Testes

#### 3.1 Documentação
- README completo com exemplos
- Documentação de API com godoc
- Exemplos de uso para diferentes cenários

#### 3.2 Testes Unitários
- Testes para conexão e reconexão
- Testes para consumo de mensagens
- Testes para publicação de mensagens
- Testes para resilience

#### 3.3 Testes de Integração
- Testes com RabbitMQ em Docker
- Testes de cenários de falha
- Testes de performance

### 4. Exemplos de Uso

#### 4.1 Consumo Básico
```go
// Exemplo de consumo de mensagens
```

#### 4.2 Publicação de Mensagens
```go
// Exemplo de publicação
```

#### 4.3 Configuração Avançada
```go
// Exemplo com DLQ, retry, etc.
```

## Próximos Passos

1. **Prioridade Alta**: Corrigir goroutine leak e implementar graceful shutdown
2. **Prioridade Média**: Implementar publicação de mensagens
3. **Prioridade Média**: Adicionar métricas e observability
4. **Prioridade Baixa**: Documentação e testes

## Arquivos a Criar/Modificar

### Novos Arquivos
- `publish.go` - Implementação de publicação
- `metrics.go` - Métricas e observability
- `options/publish.go` - Opções de publicação
- `examples/` - Diretório de exemplos
- `tests/` - Diretório de testes

### Arquivos a Modificar
- `conn/conn.go` - Corrigir goroutine leak
- `hop.go` - Adicionar `Shutdown` ao interface
- `protocol/topology.go` - Adicionar opções de publicação
- `README.md` - Documentação completa

## Estimativa de Esforço

- Correção de goroutine leak: 2-3 horas
- Implementação de publicação: 3-4 horas
- Métricas e observability: 4-5 horas
- Documentação e testes: 6-8 horas

**Total estimado: 15-20 horas**
