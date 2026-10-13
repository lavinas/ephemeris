package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoiceList is a service that handles querying and retrieving a paginated list of invoices.
type InvoiceList struct {
	Base
}

// InvoicingList is an alias for InvoiceList to support both nomenclatures.
type InvoicingList = InvoiceList

// NewInvoiceList creates a new instance of InvoiceList service.
func NewInvoiceList(repo port.Repository, logger port.Logger) *InvoiceList {
	return &InvoiceList{
		Base: *NewBase(repo, logger),
	}
}

// NewInvoicingList creates a new instance of InvoiceList service using Invoicing nomenclature.
func NewInvoicingList(repo port.Repository, logger port.Logger) *InvoiceList {
	return NewInvoiceList(repo, logger)
}

// Run executes the invoice list query process using the provided input DTO.
func (s *InvoiceList) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list invoices request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewInvoiceListResponse(400, "error", err.Error(), nil, 0, 0, 0, 0, 0)
	}

	req := input.(*dto.InvoiceListRequest)
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
			// Nickname was provided but no customer matched
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

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	invoices, total, totalAmount, err := s.repo.FindInvoicesPaged(page, pageSize, conditions, "invoicing_date desc", "id desc")
	if err != nil {
		s.logger.IPrintf(2, "Failed to retrieve invoices: %v", err)
		return dto.NewInvoiceListResponse(500, "error", err.Error(), nil, 0, 0, 0, 0, 0)
	}

	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	dtoInvoices := s.mapInvoicesToDTO(invoices)
	msg := fmt.Sprintf("retrieved %d records", len(dtoInvoices))
	return dto.NewInvoiceListResponse(200, "success", msg, dtoInvoices, page, pageSize, totalPages, total, totalAmount)
}

func (s *InvoiceList) mapInvoicesToDTO(invoices []domain.Invoice) []dto.InvoiceDTO {
	result := make([]dto.InvoiceDTO, len(invoices))
	for i, inv := range invoices {
		nickname := ""
		if inv.Plan != nil && inv.Plan.Customer != nil {
			nickname = inv.Plan.Customer.Nickname
		} else if inv.Plan != nil && inv.Plan.CustomerID != 0 && s.repo != nil {
			c, err := s.repo.GetCustomerByID(inv.Plan.CustomerID)
			if err == nil && c != nil {
				nickname = c.Nickname
			}
		} else if inv.PlanID != 0 && s.repo != nil {
			p, err := s.repo.GetPlanByID(inv.PlanID)
			if err == nil && p != nil {
				if p.Customer != nil {
					nickname = p.Customer.Nickname
				} else if p.CustomerID != 0 {
					c, err := s.repo.GetCustomerByID(p.CustomerID)
					if err == nil && c != nil {
						nickname = c.Nickname
					}
				}
			}
		}

		items := make([]dto.InvoiceItemDTO, len(inv.Items))
		for j, it := range inv.Items {
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

			items[j] = dto.InvoiceItemDTO{
				ID:               it.ID,
				ServiceID:        it.ServiceID,
				Description:      desc,
				Price:            it.Price,
				InvoicedQuantity: it.InvoicedQuantity,
				RealizedQuantity: it.RealizedQuantity,
			}
		}

		result[i] = dto.InvoiceDTO{
			ID:          inv.ID,
			InvoiceDate: inv.InvoicingDate.Format("2006-01-02"),
			Nickname:    nickname,
			DueDate:     inv.DueDate.Format("2006-01-02"),
			Amount:      inv.Amount,
			Items:       items,
		}
	}
	return result
}
