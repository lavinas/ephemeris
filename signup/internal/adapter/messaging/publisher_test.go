package messaging

import (
	"context"
	"testing"

	"signup/internal/port"
)

func TestNoopPublisher(t *testing.T) {
	pub := NewNoopPublisher()

	data := port.CustomerEventData{
		Name:     "Test",
		Nickname: "test",
	}

	if err := pub.PublishCustomerCreated(context.Background(), "acme", data); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if err := pub.PublishCustomerUpdated(context.Background(), "acme", data); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	vendorData := port.VendorEventData{
		Nickname:  "acme",
		LegalName: "Acme Corp",
	}

	if err := pub.PublishVendorCreated(context.Background(), vendorData); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if err := pub.PublishVendorUpdated(context.Background(), vendorData); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if err := pub.Close(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
