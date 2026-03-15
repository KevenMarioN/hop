package protocol

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Kind string

func (k Kind) String() string {
	return string(k)
}

const (
	Fanout  Kind = "fanout"
	Topic   Kind = "topic"
	Direct  Kind = "direct"
	Default Kind = ""
)

type Handler func(ctx context.Context, msg amqp.Delivery) error

type Queue struct {
	Durable           bool
	AutoDelete        bool
	Exclusive         bool
	NoWait            bool
	Name              string
	Headers           map[string]any
	ShouldCreateQueue bool
}

type Exchange struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Internal   bool
	Kind       Kind
	Name       string
	Headers    map[string]any
}

type Consumer struct {
	Name      string
	Key       string
	AutoAck   bool
	NoLocal   bool
	Exclusive bool
	NoWait    bool
	Headers   map[string]any
	Queue     Queue
	Exchange  *Exchange
	msg       <-chan amqp.Delivery
	Exec      Handler
}

func (c Consumer) Validate() error {
	var errs = make([]error, 0)
	if c.Name == "" {
		errs = append(errs, errors.New("consumer name cannot be empty"))
	}

	if c.Exec == nil {
		errs = append(errs, errors.New("handler cannot be empty"))
	}

	return errors.Join(errs...)
}

func (c *Consumer) Msg(msg <-chan amqp.Delivery) {
	c.msg = msg
}

func (c *Consumer) Listen() <-chan amqp.Delivery {
	return c.msg
}

func (c *Consumer) Handler(handler Handler) {
	c.Exec = handler
}

func (c *Consumer) Execute(ctx context.Context, msg amqp.Delivery) error {
	log.Info().Str("consumer", c.Name).Str("queue", c.Queue.Name).Int("size_body", len(msg.Body)).Msg("received message")
	return c.Exec(ctx, msg)
}
