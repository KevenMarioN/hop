package conn

type HopOption func(*hop)

func WithConnectionName(connectionName string) HopOption {
	return func(h *hop) {
		h.connectionName = connectionName
	}
}
