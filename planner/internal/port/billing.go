package port

import "context"

// BillingItemInput represents a line item in an invoice sent to the billing subsystem.
type BillingItemInput struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
}

// BillingInvoiceInput represents the invoice payload sent to the external billing subsystem.
type BillingInvoiceInput struct {
	Vendor       string             `json:"vendor"`
	Customer     string             `json:"customer"`
	Invoicing    string             `json:"invoicing"`
	Due          string             `json:"due"`
	ExternalCode *string            `json:"external_code,omitempty"`
	Items        []BillingItemInput `json:"items"`
}

// BillingCancelInput represents the payload sent to the external billing subsystem to cancel an invoice.
type BillingCancelInput struct {
	Vendor           string  `json:"vendor"`
	ID               int64   `json:"id,omitempty"`
	ExternalCode     *string `json:"external_code,omitempty"`
	CancellationDate *string `json:"cancellation,omitempty"`
}

// BillingClient defines the port interface for communicating with external billing subsystems.
type BillingClient interface {
	CreateInvoice(ctx context.Context, input BillingInvoiceInput) (int64, error)
	GetInvoiceByExternalCode(ctx context.Context, vendor, externalCode string) (*int64, error)
	CancelInvoice(ctx context.Context, input BillingCancelInput) error
}

