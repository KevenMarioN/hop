package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/KevenMarioN/hop/internal/metrics"
	"github.com/KevenMarioN/hop/internal/protocol"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Manager manages RabbitMQ consumers
type Manager struct {
	conn      *amqp.Connection
	consumers map[string]*protocol.Consumer
	mu        sync.RWMutex
	wg        errgroup.Group
	reconnect *sync.Cond
	collector metrics.MetricsCollector
	startTime time.Time
}

// NewManager creates a new consumer manager
func NewManager(conn *amqp.Connection, collector metrics.MetricsCollector) *Manager {
	m := &Manager{
		conn:      conn,
		consumers: make(map[string]*protocol.Consumer),
		reconnect: sync.NewCond(&sync.Mutex{}),
		startTime: time.Now(),
		collector: collector,
	}

	return m
}

// Register adds a consumer to the manager.
// Returns error if consumer validation fails or if consumer with same name already exists.
func (m *Manager) Register(consumer *protocol.Consumer) error {
	if err := consumer.Validate(); err != nil {
		return fmt.Errorf("consumer validation failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.consumers[consumer.Name]; exists {
		// Consumer already registered, update it
		m.consumers[consumer.Name] = consumer
		return nil
	}

	m.consumers[consumer.Name] = consumer

	return nil
}

// Start starts all registered consumers and begins message processing.
// This method spawns goroutines for each consumer and returns immediately.
// Consumers will automatically recover from connection failures.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()

	consumers := make(map[string]*protocol.Consumer, len(m.consumers))
	for name, consumer := range m.consumers {
		consumers[name] = consumer
	}

	m.mu.RUnlock()

	for name, consumer := range consumers {
		if err := m.recreateConsumer(consumer); err != nil {
			return fmt.Errorf("failed to start consumer %s: %w", name, err)
		}

		m.startConsumer(ctx, name, consumer)
	}

	// Update active consumers metric after starting
	if m.collector != nil {
		m.collector.Gauge("hop_active_consumers").Set(float64(len(m.consumers)))
	}

	return nil
}

// startConsumer starts a single consumer
func (m *Manager) startConsumer(ctx context.Context, name string, consumer *protocol.Consumer) {
	log.Info().Msgf("Starting consumer %s", name)

	m.wg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("consumer context done: %w", ctx.Err())
			case msg, ok := <-consumer.Listen():
				if !ok {
					log.Warn().Msgf("Consumer %s finished", name)

					// Wait for reconnection signal using sync.Cond
					m.reconnect.L.Lock()
					m.reconnect.Wait()
					m.reconnect.L.Unlock()

					// Check context again after waiting
					select {
					case <-ctx.Done():
						return fmt.Errorf("consumer context done when await reconnect: %w", ctx.Err())
					default:
					}

					if err := m.recreateConsumer(consumer); err != nil {
						return fmt.Errorf("failed restarting consumer: %w", err)
					}

					log.Info().Msgf("Consumer %s reset successful", name)
				} else {
					// Measure processing latency
					startTime := time.Now()

					if err := consumer.Execute(ctx, protocol.NewMessage(msg)); err != nil {
						log.Warn().Err(err).Msgf("Consumer %s: handler execution failed, closing channel", name)

						if err := consumer.Close(); err != nil {
							log.Warn().Err(err).Msgf("Consumer %s: closing channel", name)
							return err
						}
						// Record consumption error
						if m.collector != nil {
							m.collector.Counter("hop_consumption_errors_total", consumer.Name, "handler_error").Inc()
						}

						return fmt.Errorf("failed to execute handler: %w", err)
					}

					if m.collector != nil {
						// Record processing latency
						duration := time.Since(startTime).Seconds()
						m.collector.Histogram("hop_message_processing_duration_seconds",
							consumer.Name, consumer.Queue.Name).Observe(duration)

						// Record successful message consumption
						m.collector.Counter("hop_messages_consumed_total", consumer.Name, consumer.Queue.Name).Inc()
					}
				}
			}
		}
	})
}

// recreateConsumer re-creates a consumer after reconnection
func (m *Manager) recreateConsumer(consumer *protocol.Consumer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	channel, err := m.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	consumer.Channel(channel)

	// Always declare topology to ensure it exists
	if err := m.declareTopology(channel, consumer); err != nil {
		return fmt.Errorf("failed to declare topology: %w", err)
	}

	m.consumers[consumer.Name] = consumer

	return nil
}

// declareTopology declares queue, exchange and bindings
func (m *Manager) declareTopology(channel *amqp.Channel, consumer *protocol.Consumer) error {
	// Declare queue
	if consumer.Queue.ShouldCreateQueue {
		_, err := channel.QueueDeclare(
			consumer.Queue.Name,
			consumer.Queue.Durable,
			consumer.Queue.AutoDelete,
			consumer.Queue.Exclusive,
			consumer.Queue.NoWait,
			consumer.Queue.Headers,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", consumer.Queue.Name, err)
		}
	}

	// Declare exchange and bind if configured
	if consumer.Exchange != nil {
		if consumer.Exchange.ShouldCreateExchange {
			if err := channel.ExchangeDeclare(
				consumer.Exchange.Name,
				string(consumer.Exchange.Kind),
				consumer.Exchange.Durable,
				consumer.Exchange.AutoDelete,
				consumer.Exchange.Internal,
				consumer.Exchange.NoWait,
				consumer.Exchange.Headers,
			); err != nil {
				return fmt.Errorf("failed to declare exchange %s: %w", consumer.Exchange.Name, err)
			}
		}

		// Determine binding key
		key := consumer.Key
		if key == "" {
			key = consumer.Queue.Name
		}

		if err := channel.QueueBind(
			consumer.Queue.Name,
			key,
			consumer.Exchange.Name,
			consumer.Exchange.NoWait,
			consumer.Headers,
		); err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", consumer.Queue.Name, consumer.Exchange.Name, err)
		}
	}

	return nil
}

// Wait waits for all consumers to finish processing.
// It blocks until all consumer goroutines complete or context is cancelled.
// Returns error if no consumers are registered or if any consumer fails.
func (m *Manager) Wait() error {
	if len(m.consumers) == 0 {
		return fmt.Errorf("no consumers registered")
	}

	if err := m.wg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for consumer group: %w", err)
	}

	// Update connection duration metric
	if m.collector != nil {
		duration := time.Since(m.startTime).Seconds()
		m.collector.Gauge("hop_connection_duration_seconds").Set(duration)
	}

	return nil
}

// NotifyReconnected notifies consumers that the connection has been reestablished.
// This method is called internally by the connection monitor after a successful reconnection.
// It triggers all waiting consumers to recreate their channels and resume processing.
func (m *Manager) NotifyReconnected(conn *amqp.Connection) {
	m.conn = conn
	m.reconnect.Broadcast()

	// Increment reconnects metric
	if m.collector != nil {
		m.collector.Counter("hop_reconnects_total").Inc()
	}
}
