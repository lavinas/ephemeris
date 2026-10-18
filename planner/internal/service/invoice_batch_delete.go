package service

import (
	"fmt"
	"time"

	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceBatchDelete handles deleting all invoices matching current filter criteria.
type InvoiceBatchDelete struct {
	Base
}

// InvoicingBatchDelete is an alias for InvoiceBatchDelete to support both nomenclatures.
type InvoicingBatchDelete = InvoiceBatchDelete

// NewInvoiceBatchDelete creates a new instance of InvoiceBatchDelete service.
func NewInvoiceBatchDelete(repo port.Repository, logger port.Logger) *InvoiceBatchDelete {
	return &InvoiceBatchDelete{
		Base: *NewBase(repo, logger),
	}
}

// NewInvoicingBatchDelete creates a new instance of InvoiceBatchDelete service using Invoicing nomenclature.
func NewInvoicingBatchDelete(repo port.Repository, logger port.Logger) *InvoiceBatchDelete {
	return NewInvoiceBatchDelete(repo, logger)
}

// Run executes the batch deletion process for invoices matching filter criteria.
func (s *InvoiceBatchDelete) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing batch delete invoices request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceBatchDeleteResponse(400, "error", err.Error(), 0)
	}

	req := input.(*dto.InvoiceBatchDeleteRequest)
	conditions := map[string]interface{}{}
	conditions["planner.invoice.deleted_at IS NULL"] = nil

	if req.InvoiceID != 0 {
		conditions["planner.invoice.id = ?"] = req.InvoiceID
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
	conditions["planner.invoice.reference_invoice_id IS NULL"] = nil

	invoices, err := s.repo.FindInvoices(conditions)
	if err != nil {
		s.logger.IPrintf(2, "Failed to query invoices for batch deletion: %v", err)
		return dto.NewInvoiceBatchDeleteResponse(500, "error", fmt.Sprintf("erro ao buscar faturas para exclusão: %v", err), 0)
	}

	if len(invoices) == 0 {
		s.logger.IPrintf(2, "No invoices matched the filters for batch deletion")
		return dto.NewInvoiceBatchDeleteResponse(200, "success", "Nenhuma fatura encontrada com os filtros informados", 0)
	}

	ids := make([]int64, len(invoices))
	for i, inv := range invoices {
		ids[i] = inv.ID
	}

	if err := s.repo.DeleteInvoicesByIDs(ids); err != nil {
		s.logger.IPrintf(2, "Failed to delete invoices in batch: %v", err)
		return dto.NewInvoiceBatchDeleteResponse(500, "error", fmt.Sprintf("erro ao excluir faturas: %v", err), 0)
	}

	s.logger.IPrintf(2, "Successfully batch-deleted %d invoices", len(ids))
	return dto.NewInvoiceBatchDeleteResponse(200, "success", fmt.Sprintf("%d faturas excluídas com sucesso", len(ids)), int64(len(ids)))
}
