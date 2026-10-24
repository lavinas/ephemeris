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

// InvoiceImportCSV is a service that handles importing invoices from a CSV file.
type InvoiceImportCSV struct {
	Base
	comptime port.ComptimeService
}

// NewInvoiceImportCSV creates a new instance of InvoiceImportCSV service.
func NewInvoiceImportCSV(repo port.Repository, logger port.Logger) *InvoiceImportCSV {
	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = NewComptime(cr, logger)
	}
	return &InvoiceImportCSV{
		Base:     *NewBase(repo, logger),
		comptime: comptime,
	}
}

// SetComptimeService sets the comptime service.
func (s *InvoiceImportCSV) SetComptimeService(cs port.ComptimeService) {
	s.comptime = cs
}

// Run executes the CSV import process.
func (s *InvoiceImportCSV) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing CSV invoice import request")
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "CSV invoice import validation error: %v", err)
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", err.Error(), 0, 0, 0)
	}

	req, ok := input.(*dto.InvoiceImportCSVRequest)
	if !ok {
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", "payload de entrada inválido", 0, 0, 0)
	}

	if len(req.GroupedInvoices) == 0 {
		return dto.NewInvoiceImportCSVResponse(400, "Bad Request", "nenhuma fatura para importar", 0, 0, 0)
	}

	// For each customer, determine the first (min) and last (max) invoice date in the CSV
	type customerDateRange struct {
		firstDate time.Time
		lastDate  time.Time
	}
	custDates := make(map[int64]*customerDateRange)
	for _, group := range req.GroupedInvoices {
		dr, exists := custDates[group.CustomerID]
		if !exists {
			custDates[group.CustomerID] = &customerDateRange{
				firstDate: group.InvoiceDate,
				lastDate:  group.InvoiceDate,
			}
		} else {
			if group.InvoiceDate.Before(dr.firstDate) {
				dr.firstDate = group.InvoiceDate
			}
			if group.InvoiceDate.After(dr.lastDate) {
				dr.lastDate = group.InvoiceDate
			}
		}
	}

	// Create a manual plan for each customer whose start date is the first invoice date
	// and end date is the last invoice date from the CSV
	customerPlans := make(map[int64]int64, len(custDates))
	for custID, dr := range custDates {
		plan := domain.NewPlan(custID, dr.firstDate, &dr.lastDate, int(domain.PlanTypeManual), nil)
		if err := s.repo.SavePlan(plan); err != nil {
			s.logger.IPrintf(2, "Failed to create manual plan for customer ID %d: %v", custID, err)
			return dto.NewInvoiceImportCSVResponse(500, "error", fmt.Sprintf("erro ao criar plano manual para cliente %d: %v", custID, err), 0, 0, 0)
		}
		customerPlans[custID] = plan.ID
		s.logger.IPrintf(2, "Created manual plan #%d for customer ID %d (period: %s to %s)",
			plan.ID, custID, dr.firstDate.Format("2006-01-02"), dr.lastDate.Format("2006-01-02"))
	}

	invoicesToSave := make([]*domain.Invoice, 0, len(req.GroupedInvoices))
	totalItems := 0
	totalAmount := 0.0

	for _, group := range req.GroupedInvoices {
		planID := customerPlans[group.CustomerID]

		invTotal := 0.0
		items := make([]domain.InvoiceItem, len(group.Items))
		for i, it := range group.Items {
			lineTotal := it.Price * float64(it.Quantity)
			invTotal += lineTotal
			var desc *string
			if strings.TrimSpace(it.Description) != "" {
				d := strings.TrimSpace(it.Description)
				desc = &d
			}
			items[i] = domain.InvoiceItem{
				ServiceID:   it.ServiceID,
				Price:       it.Price,
				Description: desc,
				Quantity:    it.Quantity,
			}
		}

		inv := domain.NewInvoice(planID, group.InvoiceDate, group.DueDate, group.PeriodStart, group.PeriodEnd, invTotal, group.ReferenceInvoiceID)
		inv.Items = items

		invoicesToSave = append(invoicesToSave, inv)
		totalItems += len(items)
		totalAmount += invTotal
	}

	// Persist all invoices and their items atomically in a single transaction
	if err := s.repo.SaveInvoices(invoicesToSave); err != nil {
		s.logger.IPrintf(2, "Failed to save imported invoices: %v", err)
		return dto.NewInvoiceImportCSVResponse(500, "error", fmt.Sprintf("erro ao persistir faturas na base de dados: %v", err), 0, 0, 0)
	}

	// Trigger comptime processing for created invoice items
	if s.comptime != nil {
		for _, inv := range invoicesToSave {
			for _, it := range inv.Items {
				if err := s.comptime.ProcessInvoiceItemCreated(context.Background(), it.ID); err != nil {
					s.logger.IPrintf(2, "Aviso: falha ao processar item importado %d no banco de horas: %v", it.ID, err)
				}
			}
		}
	}

	s.logger.IPrintf(2, "CSV import completed successfully: %d invoices, %d items, total %.2f", len(invoicesToSave), totalItems, totalAmount)

	msg := fmt.Sprintf("%d faturas e %d itens importados com sucesso", len(invoicesToSave), totalItems)
	return dto.NewInvoiceImportCSVResponse(200, "success", msg, len(invoicesToSave), totalItems, totalAmount)
}
