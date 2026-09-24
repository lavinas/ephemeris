package messaging

import (
	"context"
)

// NoopConsumer is a no-op implementation of port.EventConsumer.
type NoopConsumer struct{}

// NewNoopConsumer creates a new NoopConsumer.
func NewNoopConsumer() *NoopConsumer {
	return &NoopConsumer{}
}

func (n *NoopConsumer) Start(ctx context.Context) error {
	return nil
}

func (n *NoopConsumer) Close() error {
	return nil
}
