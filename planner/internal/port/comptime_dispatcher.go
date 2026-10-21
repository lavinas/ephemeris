package port

import "context"

// ComptimeEventDispatcher defines asynchronous or synchronous dispatch of comptime events.
type ComptimeEventDispatcher interface {
	DispatchSessionEvent(ctx context.Context, action string, sessionID int64) error
	DispatchInvoiceItemEvent(ctx context.Context, action string, invoiceItemID int64) error
	DispatchInvoiceDeletedEvent(ctx context.Context, invoiceID int64) error
}
