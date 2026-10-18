package service

import (
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceDelete is a service that handles the soft-deletion of invoices.
type InvoiceDelete struct {
	Base
}

// NewInvoiceDelete creates a new instance of InvoiceDelete.
func NewInvoiceDelete(repo port.Repository, logger port.Logger) *InvoiceDelete {
	return &InvoiceDelete{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the invoice soft-delete process.
func (s *InvoiceDelete) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing delete invoice request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Invoice delete validation error: %v", err)
		return dto.NewInvoiceDeleteResponse(400, "Bad Request", err.Error())
	}

	req, ok := input.(*dto.InvoiceDeleteRequest)
	if !ok {
		return dto.NewInvoiceDeleteResponse(400, "Bad Request", "payload de entrada inválido")
	}

	invoiceID := req.InvoiceID
	existing, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Error checking invoice existence: %v", err)
		return dto.NewInvoiceDeleteResponse(500, "Internal Server Error", err.Error())
	}
	if existing == nil {
		return dto.NewInvoiceDeleteResponse(404, "Not Found", "fatura não encontrada")
	}
	if existing.ReferenceInvoiceID != nil {
		s.logger.IPrintf(2, "Invoice with ID %d already invoiced (billing #%d), cannot delete", invoiceID, *existing.ReferenceInvoiceID)
		return dto.NewInvoiceDeleteResponse(400, "Bad Request", "pré-fatura já faturada não pode ser excluída")
	}

	if err := s.repo.DeleteInvoice(invoiceID); err != nil {
		s.logger.IPrintf(2, "Error deleting invoice %d: %v", invoiceID, err)
		return dto.NewInvoiceDeleteResponse(500, "Internal Server Error", err.Error())
	}

	s.logger.IPrintf(2, "Invoice with ID %d deleted (soft-delete) successfully", invoiceID)
	return dto.NewInvoiceDeleteResponse(200, "OK", "fatura excluída com sucesso")
}
