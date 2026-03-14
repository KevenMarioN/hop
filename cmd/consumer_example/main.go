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

	if err := hop.Consume(protocol.Consumer{
		Name:      "example-hop-dollar",
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
			defer func() {
				if err := msg.Ack(true); err != nil {
					log.Error().Err(err).Msg("Failed confirm mensage")
				}
			}()

			log.Info().Str("consumer", "example").Msg(string(msg.Body))

			return nil
		},
	}); err != nil {
		log.Error().Err(err).Msg("main: failed consume")
	}

	hop.StartConsumers(ctx)

	if err := hop.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("main: failed wait hop")
		return
	}
}
