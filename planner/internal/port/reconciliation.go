package port

import (
	"time"

	"planner/internal/domain"
)

// ReconciliationService defines the use cases for reconciling sessions with invoices.
type ReconciliationService interface {
	ReconcileSession(sessionID int64) error
	DeconcileSession(sessionID int64) error
	ReconcileInvoice(invoiceID int64) error
}

// ReconciliationRepository defines the specific data access methods required by the reconciliation service.
type ReconciliationRepository interface {
	GetSessionByID(id int64) (*domain.Session, error)
	FindMatchingInvoiceItem(customerID, serviceID int64, sessionDate time.Time) (*domain.InvoiceItem, *domain.Invoice, error)
	IsSessionReconciled(sessionID int64) (bool, error)
	FindUnreconciledSessions(customerID, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error)
	AssociateSessionToInvoiceItem(invoiceItemID, sessionID int64) error
	DissociateSessionFromInvoiceItem(sessionID int64) (*domain.InvoiceItem, error)
	FindActivePayPerViewPlan(customerID, serviceID int64, sessionDate time.Time) (*domain.Plan, *domain.PlanItem, error)
	GetInvoiceByID(id int64) (*domain.Invoice, error)
	SaveInvoice(invoice *domain.Invoice) error
}
