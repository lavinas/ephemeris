package invoicing

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"planner/internal/domain"
)

// PackageStrategy handles calculation of invoices for package plans (plan_type = 2).
type PackageStrategy struct{}

// NewPackageStrategy creates a new PackageStrategy.
func NewPackageStrategy() *PackageStrategy {
	return &PackageStrategy{}
}

// Calculate generates an Invoice for the given package plan.
func (s *PackageStrategy) Calculate(plan *domain.Plan, refDate time.Time) (*domain.Invoice, error) {
	if plan == nil {
		return nil, errors.New("plano nulo")
	}
	if plan.Package == nil {
		return nil, fmt.Errorf("plano %d: dados de pacote não encontrados", plan.ID)
	}
	if len(plan.Items) == 0 {
		return nil, fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	totalQuantity := plan.Package.Quantity
	if totalQuantity <= 0 {
		return nil, fmt.Errorf("plano %d: quantidade de aulas do pacote deve ser maior que zero", plan.ID)
	}

	if len(plan.Items) > totalQuantity {
		return nil, fmt.Errorf("plano %d: número de itens (%d) não pode ser maior que a quantidade de aulas (%d)",
			plan.ID, len(plan.Items), totalQuantity)
	}

	// Ordenar itens pelo order_index
	items := make([]domain.PlanItem, len(plan.Items))
	copy(items, plan.Items)
	sort.Slice(items, func(i, j int) bool {
		return items[i].OrderIndex < items[j].OrderIndex
	})

	nItems := len(items)
	baseQty := totalQuantity / nItems
	remainder := totalQuantity % nItems

	currentDate := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := plan.Package.PaymentDate
	if dueDate.Before(currentDate) {
		dueDate = currentDate
	}

	startDateStr := plan.PlanStart.Format("02/01/2006")
	var dateSuffix string
	if plan.PlanEnd == nil || plan.PlanEnd.IsZero() || plan.PlanStart.Format("02/01/2006") == plan.PlanEnd.Format("02/01/2006") {
		dateSuffix = startDateStr
	} else {
		dateSuffix = fmt.Sprintf("%s a %s", startDateStr, plan.PlanEnd.Format("02/01/2006"))
	}

	invoiceItems := make([]domain.InvoiceItem, nItems)
	var itemsTotalAmount float64

	for i, item := range items {
		invoicedQty := baseQty
		if i < remainder {
			invoicedQty++
		}

		price := 0.0
		if item.Price != nil {
			price = *item.Price
		}

		var desc *string
		if item.Service != nil && item.Service.Description != nil && strings.TrimSpace(*item.Service.Description) != "" {
			d := fmt.Sprintf("%s - %s", strings.TrimSpace(*item.Service.Description), dateSuffix)
			desc = &d
		} else {
			descCopy := dateSuffix
			desc = &descCopy
		}

		invoiceItems[i] = domain.InvoiceItem{
			ServiceID:        item.ServiceID,
			Price:            price,
			Description:      desc,
			InvoicedQuantity: invoicedQty,
			RealizedQuantity: 0,
			CreatedAt:        refDate,
			UpdatedAt:        refDate,
			DeletedAt:        nil,
		}

		itemsTotalAmount += price * float64(invoicedQty)
	}

	amount := itemsTotalAmount
	if plan.Price != nil {
		amount = *plan.Price
	}

	periodStart := plan.PlanStart
	periodEnd := plan.PlanStart
	if plan.PlanEnd != nil && !plan.PlanEnd.IsZero() {
		periodEnd = *plan.PlanEnd
	}

	invoice := &domain.Invoice{
		PlanID:             plan.ID,
		InvoicingDate:      plan.PlanStart,
		DueDate:            dueDate,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		Amount:             amount,
		ReferenceInvoiceID: nil,
		CreatedAt:          refDate,
		UpdatedAt:          refDate,
		DeletedAt:          nil,
		Items:              invoiceItems,
	}

	return invoice, nil
}
