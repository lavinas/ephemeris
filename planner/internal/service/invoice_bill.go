package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceBill is a service that bills pre-invoices by sending them to the external billing subsystem.
type InvoiceBill struct {
	Base
	billingClient port.BillingClient
}

// InvoicingBill is an alias for InvoiceBill to support both nomenclatures.
type InvoicingBill = InvoiceBill

// NewInvoiceBill creates a new instance of InvoiceBill service.
func NewInvoiceBill(repo port.Repository, logger port.Logger, billingClient port.BillingClient) *InvoiceBill {
	return &InvoiceBill{
		Base:          *NewBase(repo, logger),
		billingClient: billingClient,
	}
}

// NewInvoicingBill creates a new instance of InvoiceBill service using Invoicing nomenclature.
func NewInvoicingBill(repo port.Repository, logger port.Logger, billingClient port.BillingClient) *InvoiceBill {
	return NewInvoiceBill(repo, logger, billingClient)
}

// Run executes the invoice billing process using the provided input DTO.
func (s *InvoiceBill) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing bill invoices request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceBillResponse(400, "error", err.Error(), nil)
	}

	req := input.(*dto.InvoiceBillRequest)
	invoices, err := s.fetchInvoices(req)
	if err != nil {
		s.logger.IPrintf(2, "Failed to retrieve invoices: %v", err)
		return dto.NewInvoiceBillResponse(500, "error", fmt.Sprintf("erro ao consultar faturas: %v", err), nil)
	}

	if len(invoices) == 0 {
		s.logger.IPrintf(2, "No invoices found for billing")
		return dto.NewInvoiceBillResponse(400, "error", "nenhuma pré-fatura encontrada para faturamento", nil)
	}

	// Rule: invoices that already have ReferenceInvoiceID filled must NOT be sent, returning Bad Request (400)
	for _, inv := range invoices {
		if inv.ReferenceInvoiceID != nil {
			msg := fmt.Sprintf("a pré-fatura %d já foi enviada para faturamento anteriormente (referência: %d)", inv.ID, *inv.ReferenceInvoiceID)
			s.logger.IPrintf(2, "Invoice %d already has ReferenceInvoiceID %d, rejecting with Bad Request", inv.ID, *inv.ReferenceInvoiceID)
			return dto.NewInvoiceBillResponse(400, "error", msg, nil)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	results := make([]dto.InvoiceBillResultItem, 0, len(invoices))
	for _, inv := range invoices {
		resultItem := s.billSingleInvoice(ctx, inv)
		results = append(results, resultItem)
	}

	successCount := 0
	for _, r := range results {
		if r.Status == "success" {
			successCount++
		}
	}

	if successCount == 0 && len(results) > 0 {
		return dto.NewInvoiceBillResponse(500, "error", fmt.Sprintf("falha ao faturar: %s", results[0].Error), results)
	}

	msg := fmt.Sprintf("%d pré-faturas faturadas com sucesso", successCount)
	return dto.NewInvoiceBillResponse(200, "success", msg, results)
}

func (s *InvoiceBill) fetchInvoices(req *dto.InvoiceBillRequest) ([]domain.Invoice, error) {
	if req.InvoiceID > 0 {
		inv, err := s.repo.GetInvoiceByID(req.InvoiceID)
		if err != nil {
			return nil, err
		}
		if inv == nil {
			return nil, nil
		}
		return []domain.Invoice{*inv}, nil
	}

	conditions := map[string]interface{}{}
	conditions["planner.invoice.deleted_at IS NULL"] = nil

	if len(req.InvoiceIDs) > 0 {
		conditions["planner.invoice.id IN (?)"] = req.InvoiceIDs
	}
	if req.PlanID != 0 {
		conditions["planner.invoice.plan_id = ?"] = req.PlanID
	}
	if req.StartDate != "" {
		if dStart, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			conditions["planner.invoice.invoicing_date >= ?"] = dStart
		}
	}
	if req.EndDate != "" {
		if dEnd, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			conditions["planner.invoice.invoicing_date <= ?"] = dEnd
		}
	}

	if len(req.CustomerIDs) == 1 {
		if req.CustomerIDs[0] == -1 {
			conditions["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE customer_id = -1)"] = nil
		} else {
			conditions["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE customer_id = ? AND deleted_at IS NULL)"] = req.CustomerIDs[0]
		}
	} else if len(req.CustomerIDs) > 1 {
		conditions["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE customer_id IN (?) AND deleted_at IS NULL)"] = req.CustomerIDs
	}

	if req.PlanType != 0 {
		conditions["planner.invoice.plan_id IN (SELECT id FROM planner.plan WHERE plan_type = ? AND deleted_at IS NULL)"] = req.PlanType
	}
	if req.ServiceID != 0 {
		conditions["planner.invoice.id IN (SELECT invoice_id FROM planner.invoice_item WHERE service_id = ? AND deleted_at IS NULL)"] = req.ServiceID
	}

	return s.repo.FindInvoices(conditions, "invoicing_date desc", "id desc")
}

func (s *InvoiceBill) billSingleInvoice(ctx context.Context, inv domain.Invoice) dto.InvoiceBillResultItem {
	vendorNickname, customerNickname, err := s.resolveVendorAndCustomer(inv)
	if err != nil {
		s.logger.IPrintf(2, "Error resolving vendor/customer for invoice %d: %v", inv.ID, err)
		return dto.InvoiceBillResultItem{
			InvoiceID: inv.ID,
			Status:    "error",
			Error:     err.Error(),
		}
	}

	items := make([]port.BillingItemInput, len(inv.Items))
	for i, it := range inv.Items {
		desc := ""
		if it.Description != nil && *it.Description != "" {
			desc = *it.Description
		} else if it.Service != nil && it.Service.Name != "" {
			desc = it.Service.Name
		} else if it.ServiceID != 0 && s.repo != nil {
			svc, err := s.repo.GetServiceByID(it.ServiceID)
			if err == nil && svc != nil {
				desc = svc.Name
			}
		}
		if desc == "" {
			desc = fmt.Sprintf("Serviço #%d", it.ServiceID)
		}

		qty := it.InvoicedQuantity
		if qty <= 0 {
			qty = 1
		}

		items[i] = port.BillingItemInput{
			Description: desc,
			Price:       it.Price,
			Quantity:    qty,
		}
	}

	extCode := fmt.Sprintf("%d", inv.ID)
	billingInput := port.BillingInvoiceInput{
		Vendor:       vendorNickname,
		Customer:     customerNickname,
		Invoicing:    inv.InvoicingDate.Format("2006-01-02"),
		Due:          inv.DueDate.Format("2006-01-02"),
		ExternalCode: &extCode,
		Items:        items,
	}

	refID, err := s.billingClient.CreateInvoice(ctx, billingInput)
	if err != nil {
		// Check for synchronization race duplicate
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "duplicidade") {
			s.logger.IPrintf(1, "Duplicate detected for invoice %d, querying billing by external code %s", inv.ID, extCode)
			existingID, getErr := s.billingClient.GetInvoiceByExternalCode(ctx, vendorNickname, extCode)
			if getErr == nil && existingID != nil && *existingID > 0 {
				refID = *existingID
				err = nil
			}
		}
	}

	if err != nil {
		s.logger.IPrintf(2, "Failed to bill invoice %d: %v", inv.ID, err)
		return dto.InvoiceBillResultItem{
			InvoiceID: inv.ID,
			Status:    "error",
			Error:     err.Error(),
		}
	}

	// Update ReferenceInvoiceID on the pre-invoice
	inv.ReferenceInvoiceID = &refID
	if updateErr := s.repo.UpdateInvoice(&inv); updateErr != nil {
		s.logger.IPrintf(2, "Failed to persist ReferenceInvoiceID %d on invoice %d: %v", refID, inv.ID, updateErr)
		return dto.InvoiceBillResultItem{
			InvoiceID:          inv.ID,
			ReferenceInvoiceID: refID,
			Status:             "error",
			Error:              fmt.Sprintf("fatura gerada no billing (#%d), mas erro ao atualizar pre-fatura local: %v", refID, updateErr),
		}
	}

	s.logger.IPrintf(2, "Successfully billed invoice %d with ReferenceInvoiceID %d", inv.ID, refID)
	return dto.InvoiceBillResultItem{
		InvoiceID:          inv.ID,
		ReferenceInvoiceID: refID,
		Status:             "success",
	}
}

func (s *InvoiceBill) resolveVendorAndCustomer(inv domain.Invoice) (string, string, error) {
	return resolveInvoiceVendorAndCustomer(s.repo, &inv)
}
