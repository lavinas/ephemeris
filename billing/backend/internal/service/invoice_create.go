package service

import (
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
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
	// Convert to domain entities
	domainInvoices, err := in.GetDomain()
	if err != nil {
		s.logger.IPrintf(2, "Failed to convert to domain entities: %v", err)
		return dto.NewInvoiceCreateResponse(500, "internal error", "contact support please")
	}
	// save domain entities
	if err := s.repo.Save(domainInvoices); err != nil {
		s.logger.IPrintf(2, "Failed to save invoices: %v", err)
		return dto.NewInvoiceCreateResponse(500, "internal error", "contact support please")
	}
	s.logger.IPrintf(2, "Input validation successful for request: %v", in)
	return dto.NewInvoiceCreateResponse(200, "success", "Invoice created successfully")
}
