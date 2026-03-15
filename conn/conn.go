package conn

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	reconnect      *sync.Cond
	backoffConfig  backoffConfig
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
		consumers:      make(map[string]protocol.Consumer, 0),
		reconnect:      sync.NewCond(&sync.Mutex{}),
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

		h.reconnect.Broadcast()

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
	return c.consume(&args)
}

func (c *hop) consume(args *protocol.Consumer) error {
	if err := args.Validate(); err != nil {
		return fmt.Errorf("consumer validation failed: %w", err)
	}

	var newConsumer bool

	if _, ok := c.consumers[args.Name]; !ok {
		c.consumers[args.Name] = *args
		newConsumer = true
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to start channel for queue %s: %w", args.Queue.Name, err)
	}

	if args.Queue.ShouldCreateQueue {
		if _, err := channel.QueueDeclare(
			args.Queue.Name, args.Queue.Durable, args.Queue.AutoDelete, args.Queue.Exclusive,
			args.Queue.NoWait, args.Queue.Headers); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", args.Queue.Name, err)
		}

		if args.Exchange != nil {
			if err := channel.ExchangeDeclare(
				args.Exchange.Name, args.Exchange.Kind.String(), args.Exchange.Durable,
				args.Exchange.AutoDelete, args.Exchange.Internal, args.Exchange.NoWait, args.Exchange.Headers); err != nil {
				return fmt.Errorf("failed to declare exchange %s: %w", args.Queue.Name, err)
			}

			if args.Key == "" {
				args.Key = args.Queue.Name
			}

			if err := channel.QueueBind(
				args.Queue.Name, args.Key, args.Exchange.Name, args.Exchange.NoWait, args.Headers); err != nil {
				return fmt.Errorf("failed to bind exchange %s: %w", args.Queue.Name, err)
			}
		}
	}

	msg, err := channel.Consume(args.Queue.Name, args.Name, args.AutoAck, args.Exclusive,
		args.NoLocal, args.NoWait, args.Headers)
	if err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	args.Msg(msg)
	c.consumers[args.Name] = *args

	if newConsumer {
		log.Info().Msgf("Consumer %s registered successfully for queue %s", args.Name, args.Queue.Name)
	}

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
	log.Info().Msgf("Starting consumer %s", name)

	c.wg.Go(func() error {
		for {
			select {
			case msg, ok := <-consumer.Listen():
				if !ok {
					log.Warn().Msgf("Consumer %s finished", name)

					// Wait for reconnection signal using sync.Cond
					c.reconnect.L.Lock()
					c.reconnect.Wait()
					c.reconnect.L.Unlock()

					if err := c.consume(&consumer); err != nil {
						return fmt.Errorf("failed restarting consumer")
					}

					log.Info().Msgf("Consumer %s reset successful", name)
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
