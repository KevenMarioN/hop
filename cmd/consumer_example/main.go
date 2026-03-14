package main

import (
	"context"

	"github.com/KevenMarioN/hop"
	"github.com/KevenMarioN/hop/protocol"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	hop, err := hop.New(ctx, "amqp://admin:admin@localhost:5672/")
	if err != nil {
		log.Error().Err(err).Msg("failed start connection hop")
		return
	}

	defer func() {
		if err := hop.Close(); err != nil {
			log.Error().Err(err).Msg("main: failed close")
			return
		}
	}()

	if err := hop.Consume(ctx, protocol.Consumer{
		Name:      "example",
		AutoAck:   false,
		NoLocal:   false,
		Exclusive: false,
		NoWait:    false,
		Queue: protocol.Queue{
			Durable: true,
			Name:    "example.queue",
			NoWait:  false,
		},
		Exec: func(ctx context.Context, msg amqp091.Delivery) error {
			defer msg.Ack(true)
			log.Info().Str("consumer", "example").Msg(string(msg.Body))
			return nil
		},
	}); err != nil {
		log.Error().Err(err).Msg("main: failed consume")
	}

	hop.StartConsumers(ctx)

	if err := hop.Wait(); err != nil {
		log.Error().Err(err).Msg("main: failed wait hop")
		return
	}
}
