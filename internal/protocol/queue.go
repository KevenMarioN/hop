package protocol

// Queue defines the configuration for a RabbitMQ queue.
type Queue struct {
	// Durable indicates if the queue survives broker restarts.
	Durable bool
	// AutoDelete indicates if the queue is automatically deleted when no consumers.
	AutoDelete bool
	// Exclusive indicates if the queue is only accessible by the current connection.
	Exclusive bool
	// NoWait indicates if the server should respond immediately (no wait for confirmation).
	NoWait bool
	// Name is the queue identifier. Empty means server-generated name.
	Name string
	// Headers are additional arguments for queue declaration (used by plugins).
	Headers map[string]any
	// ShouldCreateQueue indicates if this library should declare the queue.
	// Set to false if queue is managed externally.
	ShouldCreateQueue bool
}
