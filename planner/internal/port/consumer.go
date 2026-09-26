package port

import "context"

// EventConsumer defines the interface for an event consumer.
type EventConsumer interface {
	Start(ctx context.Context) error
	Close() error
}
