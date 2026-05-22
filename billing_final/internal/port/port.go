package port

import (
	"billing/internal/dto"
)

// Repository defines the interface for interacting with the data storage layer for invoices.
type Repository interface {
	// Save saves a new invoice to the database and returns the created invoice with its ID.
	Save(model interface{}) error
}

// Logger defines the interface for logging messages with different levels of severity.
type Logger interface {
	// IPrintf logs a formatted message at the specified log level.
	// The log level can be used to control the verbosity of the logging output.
	IPrintf(level int, format string, v ...interface{})
	// Close closes the logger and releases any resources it holds.
	Close()
}

// Service defines the interface for the business logic layer of the application.
type Service interface {
	// CreateInvoice handles the business logic for creating a new invoice based on the provided request data.
	CreateInvoice(request dto.CreateInvoiceRequest) dto.CreateInvoiceResponse
	// Close closes the service and releases any resources it holds.
	Close()
}
