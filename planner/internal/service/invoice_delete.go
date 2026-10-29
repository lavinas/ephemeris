package service

import (
	"context"
	"fmt"
	"time"

	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceDelete is a service that handles the soft-deletion of invoices.
type InvoiceDelete struct {
	Base
	billingClient port.BillingClient
	comptime      port.ComptimeService
}

// NewInvoiceDelete creates a new instance of InvoiceDelete.
func NewInvoiceDelete(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *InvoiceDelete {
	var client port.BillingClient
	if len(billingClients) > 0 && billingClients[0] != nil {
		client = billingClients[0]
	}
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceDelete{
		Base:          *NewBase(repo, logger),
		billingClient: client,
		comptime:      comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *InvoiceDelete) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
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
		if s.billingClient != nil {
			vendorNickname, _, err := resolveInvoiceVendorAndCustomer(s.repo, existing)
			if err != nil {
				s.logger.IPrintf(2, "Failed to resolve vendor for invoice %d cancellation: %v", invoiceID, err)
				return dto.NewInvoiceDeleteResponse(500, "Internal Server Error", fmt.Sprintf("erro ao resolver vendedor para cancelamento no billing: %v", err))
			}

			todayStr := time.Now().Format("2006-01-02")
			extCode := fmt.Sprintf("%d", existing.ID)
			cancelInput := port.BillingCancelInput{
				Vendor:           vendorNickname,
				ID:               *existing.ReferenceInvoiceID,
				ExternalCode:     &extCode,
				CancellationDate: &todayStr,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := s.billingClient.CancelInvoice(ctx, cancelInput); err != nil {
				s.logger.IPrintf(2, "Failed to cancel invoice %d in billing (ref #%d): %v", invoiceID, *existing.ReferenceInvoiceID, err)
				return dto.NewInvoiceDeleteResponse(500, "Internal Server Error", fmt.Sprintf("erro ao cancelar fatura no billing: %v", err))
			}
			s.logger.IPrintf(2, "Billing invoice #%d cancelled successfully for pre-invoice %d", *existing.ReferenceInvoiceID, invoiceID)
		} else {
			s.logger.IPrintf(1, "Warning: Invoice %d has ReferenceInvoiceID #%d, but billingClient is not configured", invoiceID, *existing.ReferenceInvoiceID)
		}
	}

	if err := s.repo.DeleteInvoice(invoiceID); err != nil {
		s.logger.IPrintf(2, "Error deleting invoice %d: %v", invoiceID, err)
		return dto.NewInvoiceDeleteResponse(500, "Internal Server Error", err.Error())
	}

	if s.comptime != nil {
		if err := s.comptime.ProcessInvoiceDeleted(context.Background(), invoiceID); err != nil {
			s.logger.IPrintf(2, "Aviso: falha ao estornar fatura %d no banco de horas: %v", invoiceID, err)
		}
	}

	s.logger.IPrintf(2, "Invoice with ID %d deleted (soft-delete) successfully", invoiceID)
	return dto.NewInvoiceDeleteResponse(200, "OK", "fatura excluída com sucesso")
}
