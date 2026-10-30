package port

import (
	"context"
	"time"

	"planner/internal/domain"
)

// ComptimeFilter defines query criteria for listing comptime records.
type ComptimeFilter struct {
	VendorID         int64
	CustomerID       *int64
	CustomerNickname string
	ServiceID        *int64
	StartMonth       *time.Time
	EndMonth         *time.Time
	Page             int
	PageSize         int
}

// ComptimeView represents an enriched monthly comptime record for display.
type ComptimeView struct {
	ID                   int64     `json:"id"`
	VendorID             int64     `json:"vendor_id"`
	CustomerID           int64     `json:"customer_id"`
	CustomerNickname     string    `json:"customer_nickname"`
	CustomerName         string    `json:"customer_name"`
	ServiceID            int64     `json:"service_id"`
	ServiceName          string    `json:"service_name"`
	ServiceDescription   string    `json:"service_description"`
	ServiceMinutes       int       `json:"service_minutes"`
	Month                time.Time `json:"month"`
	InvoicedMinutes      int       `json:"invoiced_minutes"`
	ConsumedMinutes      int       `json:"consumed_minutes"`
	AdjustedMinutes      int       `json:"adjusted_minutes"`
	AdjustDescription    string    `json:"adjust_description"`
	PredictMinutes       int       `json:"predict_minutes"`
	DoneMinutes          int       `json:"done_minutes"`
	MonthBalance         int       `json:"month_balance"`
	Balance              int       `json:"balance"`
	PredictSessions      float64   `json:"predict_sessions"`
	DoneSessions         float64   `json:"done_sessions"`
	AdjustedSessions     float64   `json:"adjusted_sessions"`
	MonthBalanceSessions float64   `json:"month_balance_sessions"`
	BalanceSessions      float64   `json:"balance_sessions"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// ComptimeRepository defines data access methods for banco de horas (comptime).
type ComptimeRepository interface {
	// LockCustomerService acquires an advisory transaction lock for the customer and service to serialize mutations.
	LockCustomerService(ctx context.Context, customerID, serviceID int64) error

	// Comptime main records
	FindOrCreateComptime(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time) (*domain.Comptime, error)
	GetComptime(ctx context.Context, id int64) (*domain.Comptime, error)
	GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error)
	GetPreviousComptime(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error)
	GetSubsequentComptimes(ctx context.Context, customerID, serviceID int64, month time.Time) ([]domain.Comptime, error)
	SaveComptime(ctx context.Context, comptime *domain.Comptime) error
	DeleteComptime(ctx context.Context, id int64) error
	ListComptime(ctx context.Context, filter ComptimeFilter) ([]ComptimeView, int64, error)

	// ComptimeSession
	FindComptimeSessionBySessionID(ctx context.Context, sessionID int64) (*domain.ComptimeSession, error)
	SaveComptimeSession(ctx context.Context, cs *domain.ComptimeSession) error
	DeleteComptimeSession(ctx context.Context, sessionID int64) error

	// ComptimeInvoiceItem
	FindComptimeInvoiceItemByInvoiceItemID(ctx context.Context, invoiceItemID int64) (*domain.ComptimeInvoiceItem, error)
	FindComptimeInvoiceItemsByInvoiceID(ctx context.Context, invoiceID int64) ([]domain.ComptimeInvoiceItem, error)
	SaveComptimeInvoiceItem(ctx context.Context, cii *domain.ComptimeInvoiceItem) error
	DeleteComptimeInvoiceItem(ctx context.Context, invoiceItemID int64) error

	// ComptimeAdjust
	FindComptimeAdjustByComptimeID(ctx context.Context, comptimeID int64) (*domain.ComptimeAdjust, error)
	SaveComptimeAdjust(ctx context.Context, ca *domain.ComptimeAdjust) error
	DeleteComptimeAdjustByComptimeID(ctx context.Context, comptimeID int64) error

	// Queries for resynchronization job
	FindActiveSessionsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.Session, error)
	FindActiveInvoiceItemsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.InvoiceItem, error)
	GetSessionByID(id int64) (*domain.Session, error)
	GetServiceByID(id int64) (*domain.Service, error)
	GetInvoiceByID(id int64) (*domain.Invoice, error)
	GetInvoiceItemByID(ctx context.Context, id int64) (*domain.InvoiceItem, error)
	FindActivePlanForCustomerService(ctx context.Context, customerID, serviceID int64, refDate time.Time) (*domain.Plan, error)
}

// ComptimeService defines business operations for managing comptime.
type ComptimeService interface {
	ProcessSessionCreated(ctx context.Context, sessionID int64) error
	ProcessSessionUpdated(ctx context.Context, sessionID int64) error
	ProcessSessionDeleted(ctx context.Context, sessionID int64) error

	ProcessInvoiceItemCreated(ctx context.Context, invoiceItemID int64) error
	ProcessInvoiceItemUpdated(ctx context.Context, invoiceItemID int64) error
	ProcessInvoiceItemDeleted(ctx context.Context, invoiceItemID int64) error
	ProcessInvoiceDeleted(ctx context.Context, invoiceID int64) error

	SaveComptimeAdjust(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time, adjustedMinutes int, description string) (*domain.Comptime, *domain.ComptimeAdjust, error)
	SaveComptimeAdjustByID(ctx context.Context, comptimeID int64, adjustedMinutes int, description string) (*domain.Comptime, *domain.ComptimeAdjust, error)
	DeleteComptimeAdjust(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time) (*domain.Comptime, error)
	DeleteComptimeAdjustByID(ctx context.Context, comptimeID int64) (*domain.Comptime, error)
	GetComptimeAdjust(ctx context.Context, comptimeID int64) (*domain.ComptimeAdjust, error)

	GetComptime(ctx context.Context, id int64) (*domain.Comptime, error)
	GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error)
	ListComptime(ctx context.Context, filter ComptimeFilter) ([]ComptimeView, int64, error)
	DeleteComptime(ctx context.Context, id int64) error
	Resync(ctx context.Context, vendorID int64, startDate, endDate time.Time, dataType string, customerID, serviceID *int64) (*ResyncResult, error)
}

// ResyncResult contains summary metrics of the resynchronization process.
type ResyncResult struct {
	VendorID        int64         `json:"vendor_id"`
	CustomerID      *int64        `json:"customer_id,omitempty"`
	ServiceID       *int64        `json:"service_id,omitempty"`
	StartDate       time.Time     `json:"start_date"`
	EndDate         time.Time     `json:"end_date"`
	DataType        string        `json:"data_type"`
	SessionsSynced  int           `json:"sessions_synced"`
	InvoicesSynced  int           `json:"invoices_synced"`
	ComptimeUpdated int           `json:"comptime_updated"`
	Duration        time.Duration `json:"duration"`
}
