package protocol

import (
	"errors"
	"fmt"
)

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

// Validate checks if the queue configuration is valid.
// Returns an error if:
// - Queue name is empty
// - Headers map contains invalid values
func (q Queue) Validate() error {
	var errs = make([]error, 0)

	if q.Name == "" {
		errs = append(errs, errors.New("queue name cannot be empty"))
	}

	// Validate headers
	if q.Headers != nil {
		for key, value := range q.Headers {
			if key == "" {
				errs = append(errs, errors.New("queue header key cannot be empty"))
			}

			if value == nil {
				errs = append(errs, fmt.Errorf("queue header value for key %q cannot be nil", key))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("queue validation failed: %w", errors.Join(errs...))
	}

	return nil
}
