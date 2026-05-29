package service


import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/domain"
)

// InvoiceCreate is responsible for handling the business logic of creating a new invoice.
type InvoiceCreate struct {
	Base
}

// NewInvoiceCreate creates a new instance of InvoiceCreate.
func NewInvoiceCreate(repo port.Repository, logger port.Logger) *InvoiceCreate {
	return &InvoiceCreate{
		Base: *NewBase(repo, logger),
	}
}

// Run processes the request to create a new invoice and returns the response.
func (s *InvoiceCreate) Run(inDTO port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create invoice request: %v", inDTO)
	in, ok := inDTO.(*dto.InvoiceCreateRequest)
	if !ok {
		s.logger.IPrintf(2, "Invalid input type: expected InvoiceCreateRequest")
		return dto.NewInvoiceCreateResponse(400, "bad request", "Invalid input type")
	}
	// Validate input
	if err := in.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceCreateResponse(400, "bad request",
			fmt.Sprintf("Validation failed: %v", err))
	}
	return dto.NewInvoiceCreateResponse(200, "success", "Invoice created successfully")
}

// getInvoices retrieves all invoices from the repository and returns them as a slice of domain.Invoice entities.
func (s *InvoiceCreate) getDomain(in []dto.InvoiceCreate) ([]domain.Invoice, error) {
	invoices := make([]domain.Invoice, len(in))
	for i, item := range in {
		invoices[i] = *item.GetDomain(0, 0) // Replace 0, 0 with actual businessID and customerID
	}
	return invoices, nil
}

