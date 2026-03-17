package protocol

import (
	"context"
)

// Handler is a function that processes a consumed message.
// It receives the context (for cancellation/timeout) and the AMQP delivery.
// Return nil to acknowledge the message, or an error to trigger retry/NACK.
type Handler func(ctx context.Context, msg Message) error
