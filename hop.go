// hop.go
package hop

import (
	"context"

	"github.com/KevenMarioN/hop/conn"
	"github.com/KevenMarioN/hop/protocol"
)

type Client interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
	Consume(ctx context.Context, args protocol.Consumer) error
	StartConsumers(ctx context.Context)
	Wait() error
	Close() error
}

func New(ctx context.Context, url string, opts ...conn.HopOption) (Client, error) {
	return conn.Connect(ctx, url, opts...)
}
