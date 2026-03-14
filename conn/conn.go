package conn

import (
	"context"
	"errors"
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
		return nil, fmt.Errorf("failed initialized hop: %w", err)
	}

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
		h.reconnect <- true
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			if err := h.conn.Close(); err != nil {
				log.Error().Err(err).Msgf("failed close connection %s", h.connectionName)
			}

			return
		case <-closeChan:
			if err := resilience.KeepTrying(ctx, tryReconnect); err != nil {
				log.Error().Err(err).Msg("failed retry")
			}
		}
	}
}

func (c *hop) Publish(ctx context.Context, exchange, key string, body []byte) error {
	return errors.New("don't implemented")
}

func (c *hop) Consume(ctx context.Context, args protocol.Consumer) error {
	if err := args.Validate(); err != nil {
		return fmt.Errorf("validate consumer fail: %w", err)
	}

	if _, ok := c.consumers[args.Name]; !ok {
		c.consumers[args.Name] = args
	} else {
		args.Reconnect = true
		c.consumers[args.Name] = args
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed starting channel for queue %s: %w", args.Queue.Name, err)
	}

	if _, err := channel.QueueDeclare(
		args.Queue.Name, args.Queue.Durable, args.Queue.AutoDelete, args.Queue.Exclusive,
		args.Queue.NoWait, args.Queue.Headers); err != nil {
		return fmt.Errorf("failed declare queue %s: %w", args.Queue.Name, err)
	}

	msg, err := channel.Consume(args.Queue.Name, args.Name, args.AutoAck, args.Exclusive,
		args.NoLocal, args.NoWait, args.Headers)
	if err != nil {
		return fmt.Errorf("failed start consumer: %w", err)
	}

	args.Msg(msg)
	c.consumers[args.Name] = args

	return nil
}

func (c *hop) Close() error {
	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed close connection: %w", err)
	}

	return nil
}

func (c *hop) StartConsumers(ctx context.Context) {
	for name, consumer := range c.consumers {
		if !consumer.Reconnect {
			log.Info().Msgf("Starting consume %s", name)
		} else {
			log.Info().Msgf("Restarting consume %s", name)
		}

		c.wg.Go(func() error {
			for msg := range consumer.Listen() {
				if err := consumer.Execute(ctx, msg); err != nil {
					return fmt.Errorf("failed execute handler: %w", err)
				}
			}
			log.Warn().Msgf("Finisher consume %s", name)
			select {
			case <-c.reconnect:
				if err := c.Consume(ctx, consumer); err != nil {
					return fmt.Errorf("failed restart consume %s: %w", name, err)
				}
				c.StartConsumers(ctx)
			case <-ctx.Done():
				return fmt.Errorf("context done: %w", ctx.Err())
			}

			return nil
		})
	}
}

func (c *hop) Wait() error {
	if len(c.consumers) == 0 {
		return fmt.Errorf("not have consumers for wait")
	}
	if err := c.wg.Wait(); err != nil {
		return fmt.Errorf("failed wait group")
	}

	return nil
}
