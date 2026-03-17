package protocol

// ConsumerBuilder provides a fluent API for building Consumer configurations.
// It ensures that all required fields are set before building and returns
// an immutable Consumer instance.
type ConsumerBuilder struct {
	consumer Consumer
}

// NewConsumerBuilder creates a new ConsumerBuilder with a given consumer name.
// The name is required and must be unique.
func NewConsumerBuilder(name string) *ConsumerBuilder {
	return &ConsumerBuilder{
		consumer: Consumer{
			Name: name,
			Queue: Queue{
				Durable: true, // Default: durable queues
			},
		},
	}
}

// WithQueue sets the queue configuration.
func (b *ConsumerBuilder) WithQueue(queue Queue) *ConsumerBuilder {
	b.consumer.Queue = queue
	return b
}

// WithExchange sets the exchange configuration.
func (b *ConsumerBuilder) WithExchange(exchange *Exchange) *ConsumerBuilder {
	b.consumer.Exchange = exchange
	return b
}

// WithKey sets the routing key for exchange-to-queue binding.
// If empty, the queue name is used as the routing key.
func (b *ConsumerBuilder) WithKey(key string) *ConsumerBuilder {
	b.consumer.Key = key
	return b
}

// WithAutoAck sets whether messages are automatically acknowledged.
// Default is false (manual acknowledgment).
func (b *ConsumerBuilder) WithAutoAck(autoAck bool) *ConsumerBuilder {
	b.consumer.AutoAck = autoAck
	return b
}

// WithExclusive sets whether the consumer has exclusive access to the queue.
func (b *ConsumerBuilder) WithExclusive(exclusive bool) *ConsumerBuilder {
	b.consumer.Exclusive = exclusive
	return b
}

// WithNoLocal sets whether messages published on this connection are not consumed.
func (b *ConsumerBuilder) WithNoLocal(noLocal bool) *ConsumerBuilder {
	b.consumer.NoLocal = noLocal
	return b
}

// WithNoWait sets whether the server should respond immediately to consume request.
func (b *ConsumerBuilder) WithNoWait(noWait bool) *ConsumerBuilder {
	b.consumer.NoWait = noWait
	return b
}

// WithHeaders sets additional arguments for the consume request.
func (b *ConsumerBuilder) WithHeaders(headers map[string]any) *ConsumerBuilder {
	b.consumer.Headers = headers
	return b
}

// WithHandler sets the message handler function.
// This is required and must be non-nil.
func (b *ConsumerBuilder) WithHandler(handler Handler) *ConsumerBuilder {
	b.consumer.Exec = handler
	return b
}

// Build validates the configuration and returns an immutable Consumer.
// Returns error if required fields are missing or invalid.
func (b *ConsumerBuilder) Build() (*Consumer, error) {
	if err := b.consumer.Validate(); err != nil {
		return nil, err
	}

	// Return a copy to ensure immutability
	c := b.consumer

	return &c, nil
}
