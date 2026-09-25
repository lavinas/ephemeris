package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"signup/internal/port"

	"github.com/nats-io/nats.go"
)

const (
	SubjectCustomerCreated = "customer.created"
	SubjectCustomerUpdated = "customer.updated"
	SubjectVendorCreated   = "vendor.created"
	SubjectVendorUpdated   = "vendor.updated"
)

// NATSPublisher implements port.CustomerEventPublisher using NATS.
type NATSPublisher struct {
	nc     *nats.Conn
	logger port.Logger
}

// NewNATSPublisher creates a new NATSPublisher and connects to NATS server.
func NewNATSPublisher(url string, logger port.Logger) (*NATSPublisher, error) {
	opts := []nats.Option{
		nats.Name("signup-service"),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(c *nats.Conn, err error) {
			if logger != nil && err != nil {
				logger.IPrintf(1, "NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(c *nats.Conn) {
			if logger != nil {
				logger.IPrintf(0, "NATS reconnected to %s", c.ConnectedUrl())
			}
		}),
		nats.ClosedHandler(func(c *nats.Conn) {
			if logger != nil {
				logger.IPrintf(0, "NATS connection closed")
			}
		}),
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", url, err)
	}

	if logger != nil {
		logger.IPrintf(0, "Connected to NATS at %s", url)
	}

	return &NATSPublisher{
		nc:     nc,
		logger: logger,
	}, nil
}

// PublishCustomerCreated publishes a customer creation event to NATS.
func (p *NATSPublisher) PublishCustomerCreated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	event := port.CustomerEvent{
		EventType: SubjectCustomerCreated,
		Timestamp: time.Now().UTC(),
		Vendor:    vendor,
		Data:      data,
	}
	return p.publish(SubjectCustomerCreated, event)
}

// PublishCustomerUpdated publishes a customer update event to NATS.
func (p *NATSPublisher) PublishCustomerUpdated(ctx context.Context, vendor string, data port.CustomerEventData) error {
	event := port.CustomerEvent{
		EventType: SubjectCustomerUpdated,
		Timestamp: time.Now().UTC(),
		Vendor:    vendor,
		Data:      data,
	}
	return p.publish(SubjectCustomerUpdated, event)
}

func (p *NATSPublisher) publish(subject string, event port.CustomerEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		if p.logger != nil {
			p.logger.IPrintf(2, "Failed to serialize customer event: %v", err)
		}
		return fmt.Errorf("failed to marshal customer event: %w", err)
	}

	if err := p.nc.Publish(subject, payload); err != nil {
		if p.logger != nil {
			p.logger.IPrintf(2, "Failed to publish customer event to %s: %v", subject, err)
		}
		return fmt.Errorf("failed to publish to NATS subject %s: %w", subject, err)
	}

	if p.logger != nil {
		p.logger.IPrintf(0, "Published event %s for vendor %s, customer %s", subject, event.Vendor, event.Data.Nickname)
	}
	return nil
}

// PublishVendorCreated publishes a vendor creation event to NATS.
func (p *NATSPublisher) PublishVendorCreated(ctx context.Context, data port.VendorEventData) error {
	event := port.VendorEvent{
		EventType: SubjectVendorCreated,
		Timestamp: time.Now().UTC(),
		Vendor:    data.Nickname,
		Data:      data,
	}
	return p.publishVendor(SubjectVendorCreated, event)
}

// PublishVendorUpdated publishes a vendor update event to NATS.
func (p *NATSPublisher) PublishVendorUpdated(ctx context.Context, data port.VendorEventData) error {
	event := port.VendorEvent{
		EventType: SubjectVendorUpdated,
		Timestamp: time.Now().UTC(),
		Vendor:    data.Nickname,
		Data:      data,
	}
	return p.publishVendor(SubjectVendorUpdated, event)
}

func (p *NATSPublisher) publishVendor(subject string, event port.VendorEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		if p.logger != nil {
			p.logger.IPrintf(2, "Failed to serialize vendor event: %v", err)
		}
		return fmt.Errorf("failed to marshal vendor event: %w", err)
	}

	if err := p.nc.Publish(subject, payload); err != nil {
		if p.logger != nil {
			p.logger.IPrintf(2, "Failed to publish vendor event to %s: %v", subject, err)
		}
		return fmt.Errorf("failed to publish to NATS subject %s: %w", subject, err)
	}

	if p.logger != nil {
		p.logger.IPrintf(0, "Published event %s for vendor %s", subject, event.Vendor)
	}
	return nil
}

// Close closes the underlying NATS connection.
func (p *NATSPublisher) Close() error {
	if p.nc != nil && !p.nc.IsClosed() {
		p.nc.Drain()
		p.nc.Close()
	}
	return nil
}
