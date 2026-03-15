package conn

import (
	"context"
	"fmt"
	"time"

	"github.com/KevenMarioN/hop/consumer"
	"github.com/KevenMarioN/hop/protocol"
	"github.com/KevenMarioN/hop/resilience"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type hop struct {
	conn           *amqp.Connection
	connectionName string
	consumerMgr    *consumer.Manager
	backoffConfig  backoffConfig
	config         amqp.Config
}

type backoffConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func Connect(ctx context.Context, url string, opts ...HopOption) (*hop, error) {
	c := &hop{
		conn:           nil,
		connectionName: "hop-consumer",
		config:         amqp.Config{Properties: amqp.NewConnectionProperties()},
		backoffConfig: backoffConfig{
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	c.config.Properties.SetClientConnectionName(c.connectionName)

	var err error

	if c.conn, err = amqp.DialConfig(url, c.config); err != nil {
		return nil, fmt.Errorf("failed to initialize hop connection: %w", err)
	}

	log.Info().Msgf("Connected to RabbitMQ: %s", url)

	// Create consumer manager
	c.consumerMgr = consumer.NewManager(c.conn)

	go c.monitorConnection(ctx, url)

	return c, nil
}

func (h *hop) monitorConnection(ctx context.Context, url string) {
	closeChan := h.conn.NotifyClose(make(chan *amqp.Error, 1))
	tryReconnect := func() error {
		newConn, err := amqp.DialConfig(url, h.config)
		if err != nil {
			return err
		}

		h.conn = newConn
		closeChan = newConn.NotifyClose(make(chan *amqp.Error, 1))

		// Notify consumer manager about reconnection
		h.consumerMgr.NotifyReconnected(h.conn)
		log.Info().Msg("Successfully reconnected to RabbitMQ")

		return nil
	}

	for {
		select {
		case <-ctx.Done():
			if err := h.conn.Close(); err != nil {
				log.Error().Err(err).Msgf("failed to close connection %s", h.connectionName)
			}

			return
		case <-closeChan:
			log.Warn().Msgf("Connection closed: %s", h.connectionName)

			if err := resilience.KeepTrying(ctx, tryReconnect,
				resilience.WithInitialDelay(h.backoffConfig.InitialDelay),
				resilience.WithMaxDelay(h.backoffConfig.MaxDelay),
				resilience.WithMultiplier(h.backoffConfig.Multiplier)); err != nil {
				log.Error().Err(err).Msg("failed to retry connection")
			}
		}
	}
}

func (c *hop) Publish(ctx context.Context, exchange, key string, body []byte) error {
	return ErrNotImplemented
}

func (c *hop) Consume(args protocol.Consumer) error {
	return c.consumerMgr.Register(&args)
}

func (c *hop) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

func (c *hop) Shutdown(ctx context.Context) error {
	if err := c.consumerMgr.Wait(); err != nil {
		return fmt.Errorf("failed to wait consumer manager: %w", err)
	}

	if err := c.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

func (c *hop) StartConsumers(ctx context.Context) {
	if err := c.consumerMgr.Start(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to start consumers")
	}
}

func (c *hop) Wait() error {
	return c.consumerMgr.Wait()
}
