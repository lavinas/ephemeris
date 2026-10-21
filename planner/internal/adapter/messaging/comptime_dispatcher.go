package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"

	"planner/internal/port"
)

const (
	SubjectComptimeSessionSync = "planner.comptime.session.sync"
	SubjectComptimeInvoiceSync = "planner.comptime.invoice.sync"
	QueueGroupComptimeSync     = "planner-comptime-sync"
)

// ComptimeSessionEvent represents a message published when session lifecycle changes.
type ComptimeSessionEvent struct {
	Action    string `json:"action"` // "create", "update", "delete"
	SessionID int64  `json:"session_id"`
}

// ComptimeInvoiceEvent represents a message published when invoice or item lifecycle changes.
type ComptimeInvoiceEvent struct {
	Action        string `json:"action"` // "create_item", "update_item", "delete_item", "delete_invoice"
	InvoiceID     int64  `json:"invoice_id,omitempty"`
	InvoiceItemID int64  `json:"invoice_item_id,omitempty"`
}

// SyncComptimeDispatcher dispatches comptime events directly and synchronously to ComptimeService.
type SyncComptimeDispatcher struct {
	service port.ComptimeService
}

// NewSyncComptimeDispatcher creates a new SyncComptimeDispatcher.
func NewSyncComptimeDispatcher(service port.ComptimeService) *SyncComptimeDispatcher {
	return &SyncComptimeDispatcher{service: service}
}

func (d *SyncComptimeDispatcher) DispatchSessionEvent(ctx context.Context, action string, sessionID int64) error {
	if d.service == nil {
		return nil
	}
	switch action {
	case "create":
		return d.service.ProcessSessionCreated(ctx, sessionID)
	case "update":
		return d.service.ProcessSessionUpdated(ctx, sessionID)
	case "delete":
		return d.service.ProcessSessionDeleted(ctx, sessionID)
	default:
		return fmt.Errorf("unknown session action: %s", action)
	}
}

func (d *SyncComptimeDispatcher) DispatchInvoiceItemEvent(ctx context.Context, action string, invoiceItemID int64) error {
	if d.service == nil {
		return nil
	}
	switch action {
	case "create":
		return d.service.ProcessInvoiceItemCreated(ctx, invoiceItemID)
	case "update":
		return d.service.ProcessInvoiceItemUpdated(ctx, invoiceItemID)
	case "delete":
		return d.service.ProcessInvoiceItemDeleted(ctx, invoiceItemID)
	default:
		return fmt.Errorf("unknown invoice item action: %s", action)
	}
}

func (d *SyncComptimeDispatcher) DispatchInvoiceDeletedEvent(ctx context.Context, invoiceID int64) error {
	if d.service == nil {
		return nil
	}
	return d.service.ProcessInvoiceDeleted(ctx, invoiceID)
}

// NATSComptimeDispatcher dispatches comptime events by publishing them to NATS.
type NATSComptimeDispatcher struct {
	nc       *nats.Conn
	fallback port.ComptimeEventDispatcher
}

// NewNATSComptimeDispatcher creates a new NATSComptimeDispatcher.
func NewNATSComptimeDispatcher(nc *nats.Conn, fallback port.ComptimeEventDispatcher) *NATSComptimeDispatcher {
	return &NATSComptimeDispatcher{
		nc:       nc,
		fallback: fallback,
	}
}

func (d *NATSComptimeDispatcher) DispatchSessionEvent(ctx context.Context, action string, sessionID int64) error {
	if d.nc == nil || d.nc.IsClosed() {
		if d.fallback != nil {
			return d.fallback.DispatchSessionEvent(ctx, action, sessionID)
		}
		return fmt.Errorf("NATS connection is closed and no fallback configured")
	}

	payload, err := json.Marshal(ComptimeSessionEvent{Action: action, SessionID: sessionID})
	if err != nil {
		return err
	}
	return d.nc.Publish(SubjectComptimeSessionSync, payload)
}

func (d *NATSComptimeDispatcher) DispatchInvoiceItemEvent(ctx context.Context, action string, invoiceItemID int64) error {
	if d.nc == nil || d.nc.IsClosed() {
		if d.fallback != nil {
			return d.fallback.DispatchInvoiceItemEvent(ctx, action, invoiceItemID)
		}
		return fmt.Errorf("NATS connection is closed and no fallback configured")
	}

	payload, err := json.Marshal(ComptimeInvoiceEvent{Action: action, InvoiceItemID: invoiceItemID})
	if err != nil {
		return err
	}
	return d.nc.Publish(SubjectComptimeInvoiceSync, payload)
}

func (d *NATSComptimeDispatcher) DispatchInvoiceDeletedEvent(ctx context.Context, invoiceID int64) error {
	if d.nc == nil || d.nc.IsClosed() {
		if d.fallback != nil {
			return d.fallback.DispatchInvoiceDeletedEvent(ctx, invoiceID)
		}
		return fmt.Errorf("NATS connection is closed and no fallback configured")
	}

	payload, err := json.Marshal(ComptimeInvoiceEvent{Action: "delete_invoice", InvoiceID: invoiceID})
	if err != nil {
		return err
	}
	return d.nc.Publish(SubjectComptimeInvoiceSync, payload)
}
