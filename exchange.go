package hop

import "github.com/KevenMarioN/hop/internal/protocol"

type Exchange = protocol.Exchange

type Kind = protocol.Kind

// Supported exchange types.
const (
	Fanout  Kind = "fanout" // Fanout exchange broadcasts to all bound queues
	Topic   Kind = "topic"  // Topic exchange routes based on pattern matching
	Direct  Kind = "direct" // Direct exchange routes by exact routing key
	Default Kind = ""       // Default exchange (amq.direct)
)
