# Plano de Localização e Melhoria de Logs

## Objetivo
Converter comentários e mensagens de erro para inglês, melhorar a estrutura dos erros e adicionar logs onde necessário.

## Análise Atual

### Arquivos para Analisar
1. [`conn/conn.go`](conn/conn.go) - Conexão e consumo
2. [`protocol/topology.go`](protocol/topology.go) - Estruturas de protocolo
3. [`resilience/backoff.go`](resilience/backoff.go) - Resilience e backoff
4. [`hop.go`](hop.go) - Interface principal
5. [`conn/options.go`](conn/options.go) - Opções de conexão
6. [`resilience/options.go`](resilience/options.go) - Opções de backoff

### Problemas Identificados

#### 1. Mensagens de Erro em Português
- [`conn/conn.go:41`](conn/conn.go:41): `"failed initialized hop"`
- [`conn/conn.go:70`](conn/conn.go:70): `"failed close connection %s"`
- [`conn/conn.go:76`](conn/conn.go:76): `"failed retry"`
- [`conn/conn.go:85`](conn/conn.go:85): `"failed restart consumer %s"`
- [`conn/conn.go:92`](conn/conn.go:92): `"don't implemented"`
- [`conn/conn.go:101`](conn/conn.go:101): `"validate consumer fail: %w"`
- [`conn/conn.go:113`](conn/conn.go:113): `"failed starting channel for queue %s: %w"`
- [`conn/conn.go:119`](conn/conn.go:119): `"failed declare queue %s: %w"`
- [`conn/conn.go:125`](conn/conn.go:125): `"failed start consumer: %w"`
- [`conn/conn.go:136`](conn/conn.go:136): `"failed close connection: %w"`
- [`conn/conn.go:144`](conn/conn.go:144): `"failed wait group: %w"`
- [`conn/conn.go:148`](conn/conn.go:148): `"failed close connection: %w"`
- [`conn/conn.go:156-158`](conn/conn.go:156): `"Starting consume %s"`, `"Restarting consume %s"`
- [`conn/conn.go:166`](conn/conn.go:166): `"Finisher consume %s"`
- [`conn/conn.go:172`](conn/conn.go:172): `"await reconnect consumer context done: %w"`
- [`conn/conn.go:176`](conn/conn.go:176): `"failed execute handler: %w"`
- [`conn/conn.go:180`](conn/conn.go:180): `"start consumer context done: %w"`
- [`conn/conn.go:194`](conn/conn.go:194): `"not have consumers for wait"`
- [`conn/conn.go:198`](conn/conn.go:198): `"failed wait group: %w"`

#### 2. Mensagens de Erro em Protocol
- [`protocol/topology.go:47`](protocol/topology.go:47): `"name consumer don't empty"`
- [`protocol/topology.go:51`](protocol/topology.go:51): `"handler don't empty"`

#### 3. Logs Faltantes
- [`conn/conn.go:60`](conn/conn.go:60): Reconexão bem-sucedida sem log
- [`conn/conn.go:84-87`](conn/conn.go:84): Restart de consumer sem log de sucesso
- [`conn/conn.go:128-129`](conn/conn.go:128): Consumer registrado sem log

## Plano de Ação

### 1. Converter Mensagens de Erro para Inglês

#### 1.1 [`conn/conn.go`](conn/conn.go)
- [ ] `"failed initialized hop"` → `"failed to initialize hop connection"`
- [ ] `"failed close connection %s"` → `"failed to close connection %s"`
- [ ] `"failed retry"` → `"failed to retry connection"`
- [ ] `"failed restart consumer %s"` → `"failed to restart consumer %s"`
- [ ] `"don't implemented"` → `"not implemented"`
- [ ] `"validate consumer fail: %w"` → `"consumer validation failed: %w"`
- [ ] `"failed starting channel for queue %s: %w"` → `"failed to start channel for queue %s: %w"`
- [ ] `"failed declare queue %s: %w"` → `"failed to declare queue %s: %w"`
- [ ] `"failed start consumer: %w"` → `"failed to start consumer: %w"`
- [ ] `"failed close connection: %w"` → `"failed to close connection: %w"`
- [ ] `"failed wait group: %w"` → `"failed to wait for consumer group: %w"`
- [ ] `"Starting consume %s"` → `"Starting consumer %s"`
- [ ] `"Restarting consume %s"` → `"Restarting consumer %s"`
- [ ] `"Finisher consume %s"` → `"Consumer %s finished"`
- [ ] `"await reconnect consumer context done: %w"` → `"awaiting reconnection consumer context done: %w"`
- [ ] `"failed execute handler: %w"` → `"failed to execute handler: %w"`
- [ ] `"start consumer context done: %w"` → `"consumer context done: %w"`
- [ ] `"not have consumers for wait"` → `"no consumers registered to wait for"`

#### 1.2 [`protocol/topology.go`](protocol/topology.go)
- [ ] `"name consumer don't empty"` → `"consumer name cannot be empty"`
- [ ] `"handler don't empty"` → `"handler cannot be empty"`

### 2. Melhorar Estrutura de Erros

#### 2.1 Adicionar Tipos de Erro Customizados
Criar arquivo [`conn/errors.go`](conn/errors.go):
```go
package conn

import "errors"

var (
	ErrConnectionFailed    = errors.New("connection failed")
	ErrChannelFailed       = errors.New("channel failed")
	ErrQueueDeclareFailed  = errors.New("queue declaration failed")
	ErrConsumerFailed      = errors.New("consumer failed")
	ErrValidationFailed    = errors.New("validation failed")
	ErrNotImplemented      = errors.New("not implemented")
	ErrNoConsumers         = errors.New("no consumers registered")
)
```

#### 2.2 Usar Erros Envolvidos
- [ ] Substituir `fmt.Errorf` por `fmt.Errorf("...: %w", err)` onde aplicável
- [ ] Adicionar contexto aos erros

### 3. Adicionar Logs Onde Necessário

#### 3.1 [`conn/conn.go`](conn/conn.go)
- [ ] **Linha 60**: Adicionar log de sucesso na reconexão
  ```go
  log.Info().Msgf("Successfully reconnected to RabbitMQ")
  ```
- [ ] **Linha 84-87**: Adicionar log de sucesso no restart de consumer
  ```go
  log.Info().Msgf("Successfully restarted consumer %s", name)
  ```
- [ ] **Linha 128-129**: Adicionar log de sucesso no registro de consumer
  ```go
  log.Info().Msgf("Consumer %s registered successfully", args.Name)
  ```
- [ ] **Linha 44**: Adicionar log de conexão estabelecida
  ```go
  log.Info().Msgf("Connected to RabbitMQ: %s", url)
  ```
- [ ] **Linha 74**: Adicionar log de conexão fechada
  ```go
  log.Warn().Msgf("Connection closed: %s", h.connectionName)
  ```

#### 3.2 [`protocol/topology.go`](protocol/topology.go)
- [ ] **Linha 44-55**: Adicionar log de validação
  ```go
  log.Debug().Msgf("Validating consumer: %s", c.Name)
  ```

#### 3.3 [`resilience/backoff.go`](resilience/backoff.go)
- [ ] **Linha 40-58**: Adicionar log de tentativa de reconexão
  ```go
  log.Info().Msgf("Attempting to reconnect (attempt %d)", attemptCount)
  ```

### 4. Atualizar Testes

#### 4.1 [`conn/conn_test.go`](conn/conn_test.go)
- [ ] Atualizar mensagens de erro esperadas nos testes
- [ ] Adicionar testes para novos logs

#### 4.2 [`protocol/topology_test.go`](protocol/topology_test.go)
- [ ] Atualizar mensagens de erro esperadas nos testes

## Arquivos a Criar

1. [`conn/errors.go`](conn/errors.go) - Tipos de erro customizados

## Arquivos a Modificar

1. [`conn/conn.go`](conn/conn.go) - Mensagens de erro e logs
2. [`protocol/topology.go`](protocol/topology.go) - Mensagens de erro
3. [`resilience/backoff.go`](resilience/backoff.go) - Logs
4. [`conn/conn_test.go`](conn/conn_test.go) - Atualizar testes
5. [`protocol/topology_test.go`](protocol/topology_test.go) - Atualizar testes

## Ordem de Execução

1. **Prioridade Alta**: Criar [`conn/errors.go`](conn/errors.go)
2. **Prioridade Alta**: Atualizar mensagens de erro em [`conn/conn.go`](conn/conn.go)
3. **Prioridade Alta**: Atualizar mensagens de erro em [`protocol/topology.go`](protocol/topology.go)
4. **Prioridade Média**: Adicionar logs em [`conn/conn.go`](conn/conn.go)
5. **Prioridade Média**: Adicionar logs em [`resilience/backoff.go`](resilience/backoff.go)
6. **Prioridade Baixa**: Atualizar testes
7. **Prioridade Baixa**: Revisar e refinar logs

## Estimativa de Esforço

- Criação de erros customizados: 1 hora
- Atualização de mensagens de erro: 2 horas
- Adição de logs: 2 horas
- Atualização de testes: 1 hora
- Revisão e refinamento: 1 hora

**Total estimado: 7 horas**
