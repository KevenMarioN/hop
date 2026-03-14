package conn

import (
	"context"
	"fmt"

	"github.com/KevenMarioN/hop/protocol"
	"github.com/KevenMarioN/hop/resilience"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type hop struct {
	conn           *amqp.Connection
	connectionName string
	consumers      map[string]protocol.Consumer
	wg             errgroup.Group
	reconnect      chan bool
}

func Connect(ctx context.Context, url string, opts ...HopOption) (*hop, error) {
	c := &hop{
		conn:           nil,
		connectionName: "hop-consumer",
		consumers:      make(map[string]protocol.Consumer, 0),
		reconnect:      make(chan bool, 1),
	}

	for _, opt := range opts {
		opt(c)
	}

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(c.connectionName)

	var err error

	if c.conn, err = amqp.DialConfig(url, config); err != nil {
		return nil, fmt.Errorf("failed to initialize hop connection: %w", err)
	}

	log.Info().Msgf("Connected to RabbitMQ: %s", url)
	go c.monitorConnection(ctx, url, config)

	return c, nil
}

func (h *hop) monitorConnection(ctx context.Context, url string, config amqp.Config) {
	closeChan := h.conn.NotifyClose(make(chan *amqp.Error, 1))
	tryReconnect := func() error {
		newConn, err := amqp.DialConfig(url, config)
		if err != nil {
			return err
		}

		h.conn = newConn
		closeChan = newConn.NotifyClose(make(chan *amqp.Error, 1))

		h.restartConsumers(ctx)
		h.reconnect <- true
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
			if err := resilience.KeepTrying(ctx, tryReconnect); err != nil {
				log.Error().Err(err).Msg("failed to retry connection")
			}
		}
	}
}

func (h *hop) restartConsumers(ctx context.Context) {
	for name, consumer := range h.consumers {
		if err := h.consume(&consumer); err != nil {
			log.Error().Err(err).Msgf("failed to restart consumer %s", name)
		} else {
			log.Info().Msgf("Successfully restarted consumer %s", name)
		}
		h.startConsumer(ctx, name, consumer)
	}
}

func (c *hop) Publish(ctx context.Context, exchange, key string, body []byte) error {
	return ErrNotImplemented
}

func (c *hop) Consume(args protocol.Consumer) error {
	return c.consume(&args)
}

func (c *hop) consume(args *protocol.Consumer) error {
	if err := args.Validate(); err != nil {
		return fmt.Errorf("consumer validation failed: %w", err)
	}

	if _, ok := c.consumers[args.Name]; !ok {
		c.consumers[args.Name] = *args
	} else {
		args.Reconnect = true
		c.consumers[args.Name] = *args
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to start channel for queue %s: %w", args.Queue.Name, err)
	}

	if _, err := channel.QueueDeclare(
		args.Queue.Name, args.Queue.Durable, args.Queue.AutoDelete, args.Queue.Exclusive,
		args.Queue.NoWait, args.Queue.Headers); err != nil {
		return fmt.Errorf("failed to declare queue %s: %w", args.Queue.Name, err)
	}

	msg, err := channel.Consume(args.Queue.Name, args.Name, args.AutoAck, args.Exclusive,
		args.NoLocal, args.NoWait, args.Headers)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	args.Msg(msg)
	c.consumers[args.Name] = *args
	log.Info().Msgf("Consumer %s registered successfully", args.Name)

	return nil
}

func (c *hop) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

func (c *hop) Shutdown(ctx context.Context) error {
	if err := c.wg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for consumer group: %w", err)
	}

	if err := c.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

func (c *hop) startConsumer(ctx context.Context, name string, consumer protocol.Consumer) {
	if !consumer.Reconnect {
		log.Info().Msgf("Starting consumer %s", name)
	} else {
		log.Info().Msgf("Restarting consumer %s", name)
	}

	c.wg.Go(func() error {
		for {
			select {
			case msg, ok := <-consumer.Listen():
				if !ok {
					log.Warn().Msgf("Consumer %s finished", name)

					select {
					case <-c.reconnect:
						return nil
					case <-ctx.Done():
						return fmt.Errorf("awaiting reconnection consumer context done: %w", ctx.Err())
					}
				} else {
					if err := consumer.Execute(ctx, msg); err != nil {
						return fmt.Errorf("failed to execute handler: %w", err)
					}
				}
			case <-ctx.Done():
				return fmt.Errorf("consumer context done: %w", ctx.Err())
			}
		}
	})
}

func (c *hop) StartConsumers(ctx context.Context) {
	for name, consumer := range c.consumers {
		c.startConsumer(ctx, name, consumer)
	}
}

func (c *hop) Wait() error {
	if len(c.consumers) == 0 {
		return ErrNoConsumers
	}

	if err := c.wg.Wait(); err != nil {
		return fmt.Errorf("failed to wait for consumer group: %w", err)
	}

	return nil
}
