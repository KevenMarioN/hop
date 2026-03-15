package conn

import "errors"

var (
	// ErrConnectionFailed indicates failure to connect to RabbitMQ
	ErrConnectionFailed = errors.New("connection failed")

	// ErrChannelFailed indicates failure to create channel
	ErrChannelFailed = errors.New("channel failed")

	// ErrQueueDeclareFailed indicates failure to declare queue
	ErrQueueDeclareFailed = errors.New("queue declaration failed")

	// ErrConsumerFailed indicates failure to register consumer
	ErrConsumerFailed = errors.New("consumer failed")

	// ErrValidationFailed indicates validation failure
	ErrValidationFailed = errors.New("validation failed")

	// ErrNotImplemented indicates functionality not implemented
	ErrNotImplemented = errors.New("not implemented")

	// ErrNoConsumers indicates no consumers registered
	ErrNoConsumers = errors.New("no consumers registered")
)
