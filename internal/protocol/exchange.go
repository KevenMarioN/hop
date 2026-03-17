package protocol

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

func (k Kind) String() string {
	return string(k)
}
