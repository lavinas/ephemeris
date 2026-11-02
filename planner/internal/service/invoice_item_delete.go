package service

import (
	"context"
	"fmt"
	"time"

	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceItemDelete handles soft-deleting an item from an existing pre-invoice and updating its total and comptime.
type InvoiceItemDelete struct {
	Base
	billingClient port.BillingClient
	comptime      port.ComptimeService
}

// InvoicingItemDelete is an alias for InvoiceItemDelete.
type InvoicingItemDelete = InvoiceItemDelete

// NewInvoiceItemDelete creates a new instance of InvoiceItemDelete service.
func NewInvoiceItemDelete(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *InvoiceItemDelete {
	var client port.BillingClient
	if len(billingClients) > 0 && billingClients[0] != nil {
		client = billingClients[0]
	}
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceItemDelete{
		Base:          *NewBase(repo, logger),
		billingClient: client,
		comptime:      comptime,
	}
}

// NewInvoicingItemDelete creates a new instance of InvoiceItemDelete service using Invoicing nomenclature.
func NewInvoicingItemDelete(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *InvoiceItemDelete {
	return NewInvoiceItemDelete(repo, logger, billingClients...)
}

// SetComptimeService sets the comptime service.
func (s *InvoiceItemDelete) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// Run executes deleting an item from a pre-invoice.
func (s *InvoiceItemDelete) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing delete invoice item request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceItemDeleteResponse(400, "error", err.Error(), nil)
	}

	req, ok := input.(*dto.InvoiceItemDeleteRequest)
	if !ok {
		return dto.NewInvoiceItemDeleteResponse(400, "error", "payload de entrada inválido", nil)
	}

	invoiceID := req.InvoiceID
	if invoiceID <= 0 {
		if cr, ok := s.repo.(port.ComptimeRepository); ok {
			it, err := cr.GetInvoiceItemByID(context.Background(), req.ItemID)
			if err != nil {
				return dto.NewInvoiceItemDeleteResponse(500, "error", fmt.Sprintf("erro ao consultar item de fatura: %v", err), nil)
			}
			if it != nil {
				invoiceID = it.InvoiceID
			}
		}
	}
	if invoiceID <= 0 {
		return dto.NewInvoiceItemDeleteResponse(404, "error", "pré-fatura do item não encontrada", nil)
	}

	invoice, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil {
		s.logger.IPrintf(2, "Failed to get invoice #%d: %v", invoiceID, err)
		return dto.NewInvoiceItemDeleteResponse(500, "error", fmt.Sprintf("erro ao buscar fatura: %v", err), nil)
	}
	if invoice == nil || invoice.DeletedAt != nil {
		return dto.NewInvoiceItemDeleteResponse(404, "error", "pré-fatura não encontrada", nil)
	}

	now := time.Now()
	found := false
	for i := range invoice.Items {
		if invoice.Items[i].ID == req.ItemID && invoice.Items[i].DeletedAt == nil {
			invoice.Items[i].DeletedAt = &now
			invoice.Items[i].UpdatedAt = now
			found = true
			break
		}
	}
	if !found {
		return dto.NewInvoiceItemDeleteResponse(404, "error", "item de fatura não encontrado", nil)
	}

	totalAmount := 0.0
	for _, it := range invoice.Items {
		if it.DeletedAt == nil {
			totalAmount += it.Price * float64(it.Quantity)
		}
	}
	invoice.Amount = totalAmount

	// If the invoice was already billed in billing, cancel previous billing invoice and create replacement
	if invoice.ReferenceInvoiceID != nil && s.billingClient != nil {
		if err := syncUpdatedInvoiceWithBilling(s.repo, s.billingClient, s.logger, invoice); err != nil {
			return dto.NewInvoiceItemDeleteResponse(500, "error", err.Error(), nil)
		}
	}

	if err := s.repo.UpdateInvoice(invoice); err != nil {
		s.logger.IPrintf(2, "Failed to update invoice after item deletion in database: %v", err)
		return dto.NewInvoiceItemDeleteResponse(500, "error", fmt.Sprintf("erro ao remover item da pré-fatura: %v", err), nil)
	}

	if s.comptime != nil {
		if err := s.comptime.ProcessInvoiceItemDeleted(context.Background(), req.ItemID); err != nil {
			s.logger.IPrintf(2, "Aviso: falha ao estornar item %d no banco de horas: %v", req.ItemID, err)
		}
	}

	s.logger.IPrintf(2, "Item #%d deleted from invoice #%d. New amount: %.2f", req.ItemID, invoice.ID, invoice.Amount)

	updater := NewInvoiceUpdate(s.repo, s.logger, s.billingClient)
	invoiceDTO := updater.mapInvoiceToDTO(invoice)
	return dto.NewInvoiceItemDeleteResponse(200, "success", "Item de fatura excluído com sucesso", invoiceDTO)
}

