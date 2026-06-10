package port

import "billing/internal/domain"

// Repository defines the interface for interacting with the data storage layer for invoices.
type Repository interface {
	// Save saves a new invoice to the database and returns the created invoice with its ID.
	Save(model interface{}) error
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	Close() error
	FindCustomers(page, pageSize int, vendorID int64, name, nickname,
		document *string, status *int, email, whatsapp *string) ([]domain.Customer, error)
	GetCustomer(vendorID int64, nickname string) (*domain.Customer, error)
	FindVendors(page, pageSize int, legalName, nickname, document *string,
		accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error)
	GetVendor(nickname string) (*domain.Vendor, error)
	FindInvoices(page, pageSize int, customer int64,
		invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, taxDate *string) ([]domain.Invoice, error)
}

// Logger defines the interface for logging messages with different levels of severity.
type Logger interface {
	// IPrintf logs a formatted message at the specified log level.
	// The log level can be used to control the verbosity of the logging output.
	IPrintf(level int, format string, v ...interface{})
	// Close closes the logger and releases any resources it holds.
	Close()
}

// InDTO represents a generic data transfer object for input of service methods.
type InDTO interface {
	Validate(repo Repository) error
	Reset()
}

// OutDTO represents a generic data transfer object for output of service methods.
type OutDTO interface {
	GetStatusCode() int
}

type Service interface {
	Run(in InDTO) (out OutDTO)
}
