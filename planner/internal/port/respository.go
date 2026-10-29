package port

import (
	"time"

	"planner/internal/domain"
)

// Repository defines the interface for interacting with the data storage layer.
type Repository interface {
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Save(model interface{}) error
	Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error)
	FindCount(conditions map[string]interface{}) (int64, error)
	FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error)
	Close() error
	FindCustomers(page, pageSize int, vendorID int64, name, nickname,
		document *string, status *int, email, whatsapp *string) ([]domain.Customer, error)
	GetCustomer(vendorID int64, nickname string) (*domain.Customer, error)
	GetCustomerByID(id int64) (*domain.Customer, error)
	FindVendors(page, pageSize int, legalName, nickname, document *string,
		accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error)
	GetVendor(nickname string) (*domain.Vendor, error)
	GetVendorByID(id int64) (*domain.Vendor, error)
	FindServices(vendorID int64) ([]domain.Service, error)
	GetServiceByID(id int64) (*domain.Service, error)
	SavePlan(plan *domain.Plan) error
	FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error)
	GetPlanByID(id int64) (*domain.Plan, error)
	DeletePlan(id int64) error
	FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error)
	SaveInvoice(invoice *domain.Invoice) error
	SaveInvoices(invoices []*domain.Invoice) error
	GetInvoiceByID(id int64) (*domain.Invoice, error)
	FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error)
	FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error)
	DeleteInvoice(id int64) error
	UpdateInvoice(invoice *domain.Invoice) error
	DeleteInvoicesByIDs(ids []int64) error
	FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error)
	FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error)
	GetInvoiceForSession(sessionID int64) (*domain.Invoice, error)
	CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error)
}
