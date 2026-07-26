package port

import (
	"time"

	"billing/internal/domain"
)

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
		invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, taxDate,
		cancellationDate *string) ([]domain.Invoice, error)
	GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error)
	GetInvoice(id int64) (*domain.Invoice, error)
	GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error)
	GetEmissionsCount(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) (int64, error)
	GetEmissionLastRPS(vendorID int64) (int64, error)
	GetEmission(id int64) (*domain.Emission, error)
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

// Service defines the interface for a service that processes input data and produces output data.
type Service interface {
	Run(in InDTO) (out OutDTO)
}

// Taxer defines the interface for sending emissions to an external system.
type Taxer interface {
	SendEmission(emission *domain.Emission) error
	ReceiveEmission(source string) (map[int64]*domain.EmissionItem, error)
}

// Biller defines the interface for generating PDF files for invoices.
type Biller interface {
	Generate(request InDTO, path string) error
	GetBinary(request InDTO) ([]byte, error)
	GetPDFBase64(request InDTO) (string, error)
	SendMail(request InDTO) error
}

// Pixer defines the interface for generating Pix payment payloads.
type Pixer interface {
	Get(request InDTO) (string, string, error)
}

// Issuer defines the interface for generating and sending receipts.
type Issuer interface {
	GetBase64(data InDTO, html_pdf string) ([]byte, error)
	SendMail(data InDTO, html_pdf string, html_email string) error
}
