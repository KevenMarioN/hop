package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/KevenMarioN/hop/protocol"
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
}

// NewManager creates a new consumer manager
func NewManager(conn *amqp.Connection) *Manager {
	m := &Manager{
		conn:      conn,
		consumers: make(map[string]*protocol.Consumer),
		reconnect: sync.NewCond(&sync.Mutex{}),
	}

	return m
}

// Register adds a consumer to the manager
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

// Start starts all registered consumers
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for name, consumer := range m.consumers {
		if err := m.recreateConsumer(consumer); err != nil {
			return fmt.Errorf("failed to start consumer %s: %w", name, err)
		}

		m.startConsumer(ctx, name, consumer)
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
						return ctx.Err()
					default:
					}

					if err := m.recreateConsumer(consumer); err != nil {
						return fmt.Errorf("failed restarting consumer: %w", err)
					}

					log.Info().Msgf("Consumer %s reset successful", name)
				} else {
					if err := consumer.Execute(ctx, msg); err != nil {
						return fmt.Errorf("failed to execute handler: %w", err)
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

	defer func() {
		if err != nil {
			if closeErr := channel.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("failed to close channel")
			}
		}
	}()

	// Always declare topology to ensure it exists
	if err := m.declareTopology(channel, consumer); err != nil {
		return fmt.Errorf("failed to declare topology: %w", err)
	}

	msg, err := channel.Consume(
		consumer.Queue.Name,
		consumer.Name,
		consumer.AutoAck,
		consumer.Exclusive,
		consumer.NoLocal,
		consumer.NoWait,
		consumer.Headers,
	)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	consumer.Msg(msg)

	m.consumers[consumer.Name] = consumer

	return nil
}

// declareTopology declares queue, exchange and bindings
func (m *Manager) declareTopology(channel *amqp.Channel, consumer *protocol.Consumer) error {
	// Declare queue
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

	// Declare exchange and bind if configured
	if consumer.Exchange != nil {
		err := channel.ExchangeDeclare(
			consumer.Exchange.Name,
			string(consumer.Exchange.Kind),
			consumer.Exchange.Durable,
			consumer.Exchange.AutoDelete,
			consumer.Exchange.Internal,
			consumer.Exchange.NoWait,
			consumer.Exchange.Headers,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", consumer.Exchange.Name, err)
		}

		// Determine binding key
		key := consumer.Key
		if key == "" {
			key = consumer.Queue.Name
		}

		err = channel.QueueBind(
			consumer.Queue.Name,
			key,
			consumer.Exchange.Name,
			consumer.Exchange.NoWait,
			consumer.Headers,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", consumer.Queue.Name, consumer.Exchange.Name, err)
		}
	}

	return nil
}

// Wait waits for all consumers to finish
func (m *Manager) Wait() error {
	if len(m.consumers) == 0 {
		return fmt.Errorf("no consumers registered")
	}

	if err := m.wg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for consumer group: %w", err)
	}

	return nil
}

// NotifyReconnected notifies consumers that the connection has been reestablished
func (m *Manager) NotifyReconnected(conn *amqp.Connection) {
	m.conn = conn
	m.reconnect.Broadcast()
}
