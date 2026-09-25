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

// CustomerEventPublisher defines the contract for publishing customer and vendor events.
type CustomerEventPublisher interface {
	PublishCustomerCreated(ctx context.Context, vendor string, data CustomerEventData) error
	PublishCustomerUpdated(ctx context.Context, vendor string, data CustomerEventData) error
	PublishVendorCreated(ctx context.Context, data VendorEventData) error
	PublishVendorUpdated(ctx context.Context, data VendorEventData) error
	Close() error
}

// EventPublisher is an alias for CustomerEventPublisher.
type EventPublisher = CustomerEventPublisher

// VendorEvent represents an event related to a vendor lifecycle.
type VendorEvent struct {
	EventType string          `json:"event_type"` // "vendor.created" or "vendor.updated"
	Timestamp time.Time       `json:"timestamp"`
	Vendor    string          `json:"vendor"`
	Data      VendorEventData `json:"data"`
}

// VendorEventData represents the payload of vendor information in the event.
type VendorEventData struct {
	Nickname      string `json:"nickname"`
	LegalName     string `json:"legal_name"`
	TradingName   string `json:"trading_name"`
	Document      string `json:"document"`
	TaxDocument   string `json:"tax_document"`
	AccountBank   string `json:"account_bank"`
	AccountAgency string `json:"account_agency"`
	AccountNumber string `json:"account_number"`
	PixToken      string `json:"pix_token"`
	PixName       string `json:"pix_name"`
	PixCity       string `json:"pix_city"`
	LogoName      string `json:"logo_name"`
	Email         string `json:"email"`
	Whatsapp      string `json:"whatsapp"`
	LastRps       int64  `json:"last_rps"`
	SmtpHost      string `json:"smtp_host"`
	SmtpPort      int    `json:"smtp_port"`
	SmtpUser      string `json:"smtp_user"`
	SmtpPassword  string `json:"smtp_password"`
}
