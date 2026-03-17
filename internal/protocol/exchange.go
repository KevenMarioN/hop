package protocol

import (
	"errors"
	"fmt"
)

// Exchange defines the configuration for a RabbitMQ exchange.
type Exchange struct {
	// ShouldCreateExchange indicates if this library should declare the exchange.
	// Set to false if exchange is managed externally.
	ShouldCreateExchange bool
	// Durable indicates if the exchange survives broker restarts.
	Durable bool
	// AutoDelete indicates if the exchange is automatically deleted when no queues bound.
	AutoDelete bool
	// Exclusive indicates if the exchange is only accessible by the current connection.
	Exclusive bool
	// NoWait indicates if the server should respond immediately.
	NoWait bool
	// Internal indicates if the exchange is for internal broker use only.
	Internal bool
	// Kind is the exchange type (fanout, topic, direct, etc.).
	Kind Kind
	// Name is the exchange identifier.
	Name string
	// Headers are additional arguments for exchange declaration.
	Headers map[string]any
}

// Kind represents the type of AMQP exchange.
type Kind string

// Supported exchange types.
const (
	Fanout  Kind = "fanout" // Fanout exchange broadcasts to all bound queues
	Topic   Kind = "topic"  // Topic exchange routes based on pattern matching
	Direct  Kind = "direct" // Direct exchange routes by exact routing key
	Default Kind = ""       // Default exchange (amq.direct)
)

func (k Kind) String() string {
	return string(k)
}

// Validate checks if the exchange configuration is valid.
// Returns an error if:
// - Exchange name is empty
// - Exchange kind is invalid (not one of: fanout, topic, direct, or empty)
func (e Exchange) Validate() error {
	var errs = make([]error, 0)

	if e.Name == "" {
		errs = append(errs, errors.New("exchange name cannot be empty"))
	}

	// Validate exchange kind
	validKinds := []Kind{Fanout, Topic, Direct, Default}
	kindValid := false

	for _, validKind := range validKinds {
		if e.Kind == validKind {
			kindValid = true
			break
		}
	}

	if !kindValid {
		errs = append(errs, fmt.Errorf(
			"invalid exchange kind: %q (valid kinds: fanout, topic, direct, or empty for default)", e.Kind))
	}

	if len(errs) > 0 {
		return fmt.Errorf("exchange validation failed: %w", errors.Join(errs...))
	}

	return nil
}
