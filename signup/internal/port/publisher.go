package port

import (
	"context"
	"time"
)

// CustomerEvent represents an event related to a customer lifecycle.
type CustomerEvent struct {
	EventType string            `json:"event_type"` // "customer.created" or "customer.updated"
	Timestamp time.Time         `json:"timestamp"`
	Vendor    string            `json:"vendor"`
	Data      CustomerEventData `json:"data"`
}

// CustomerEventData represents the payload of customer information in the event.
type CustomerEventData struct {
	Name     string  `json:"name"`
	Nickname string  `json:"nickname"`
	Document *string `json:"document,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
	Status   *int    `json:"status,omitempty"`
}

// CustomerEventPublisher defines the contract for publishing customer events.
type CustomerEventPublisher interface {
	PublishCustomerCreated(ctx context.Context, vendor string, data CustomerEventData) error
	PublishCustomerUpdated(ctx context.Context, vendor string, data CustomerEventData) error
	Close() error
}
