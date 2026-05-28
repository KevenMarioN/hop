package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KevenMarioN/hop"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	clientHop, err := hop.New(ctx, "amqp://admin:admin@localhost:5672/",
		hop.WithBackoff(2, time.Second*1, time.Minute*1),
		hop.WithMetrics(hop.NewOpenTelemetryCollector("hop")))
	if err != nil {
		log.Error().Err(err).Msg("failed start connection hop")
		return
	}

	go func() {
		sig := <-sigChan
		log.Info().Msgf("Capture signal: %v. Initiating Shutdown.", sig)

		if err := clientHop.Shutdown(); err != nil {
			log.Error().Err(err).Msg("failed shutdown")
		}
	}()

	if err := clientHop.Consume(hop.Consumer{
		Name:      "example-hop-dollar",
		AutoAck:   false,
		NoLocal:   false,
		Exclusive: false,
		NoWait:    false,
		Queue: hop.Queue{
			Durable:           true,
			Name:              "example.queue",
			NoWait:            false,
			ShouldCreateQueue: true,
		},
		Exec: func(ctx context.Context, msg hop.Message) error {
			defer func() {
				if err := msg.Success(); err != nil {
					log.Error().Err(err).Msg("Failed to confirm message")
				}
			}()

			log.Info().Str("consumer", "example").Msg(string(msg.Body))

			return nil
		},
	}); err != nil {
		log.Error().Err(err).Msg("main: failed consume")
	}

	if err := clientHop.Consume(hop.Consumer{
		Name:      "example-hop-eruo",
		AutoAck:   false,
		NoLocal:   false,
		Exclusive: false,
		NoWait:    false,
		Queue: hop.Queue{
			Durable:           true,
			Name:              "example.queue.euro",
			NoWait:            false,
			ShouldCreateQueue: true,
		},
		Key: "transferencia",
		Exchange: &hop.Exchange{
			Durable:              true,
			Kind:                 hop.Direct,
			Name:                 "banco",
			ShouldCreateExchange: true,
		},
		Exec: func(ctx context.Context, msg hop.Message) error {
			defer func() {
				if err := msg.Success(); err != nil {
					log.Error().Err(err).Msg("Failed to confirm message")
				}
			}()

			log.Info().Str("consumer", "example").Msg(string(msg.Body))

			if string(msg.Body) == `{"msg": "failed"}` {
				return errors.New("failed")
			}

			return nil
		},
	}); err != nil {
		log.Error().Err(err).Msg("main: failed consume")
	}

	clientHop.StartConsumers(ctx)
	clientHop.WaitConsumers()

	if err := clientHop.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed shutdown")
	}
}
