package messaging

import (
	"context"

	"signup/internal/port"
)

// NoopPublisher is a null-object implementation of port.CustomerEventPublisher.
type NoopPublisher struct{}

// NewNoopPublisher creates a new NoopPublisher.
func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (n *NoopPublisher) PublishCustomerCreated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	return nil
}

func (n *NoopPublisher) PublishCustomerUpdated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	return nil
}

func (n *NoopPublisher) Close() error {
	return nil
}
