package invoicing

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"planner/internal/domain"
)

// AgendaStrategy handles calculation of invoices for recurring agenda plans (plan_type = 1).
type AgendaStrategy struct{}

// NewAgendaStrategy creates a new AgendaStrategy.
func NewAgendaStrategy() *AgendaStrategy {
	return &AgendaStrategy{}
}

// IsActiveInMonth checks if the plan is active in the given month.
func (s *AgendaStrategy) IsActiveInMonth(plan *domain.Plan, monthStart time.Time) bool {
	monthEnd := monthStart.AddDate(0, 1, -1)
	return !plan.PlanStart.After(monthEnd) && (plan.PlanEnd == nil || !plan.PlanEnd.Before(monthStart))
}

// CalculateMonth calculates the invoice for a recurring agenda plan in a given month.
// Returns (nil, nil) if the plan is not active in that month.
func (s *AgendaStrategy) CalculateMonth(plan *domain.Plan, monthStart, refDate time.Time) (*domain.Invoice, error) {
	if plan == nil {
		return nil, errors.New("plano nulo")
	}
	if plan.Agenda == nil {
		return nil, fmt.Errorf("plano %d: dados de agenda não encontrados", plan.ID)
	}
	if len(plan.Items) == 0 {
		return nil, fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	if !s.IsActiveInMonth(plan, monthStart) {
		return nil, nil
	}

	// Calcular período alvo de sessões de acordo com plan_agenda.term_type:
	// se 1 (pre-paid): mês corrente ao processamento
	// se 2 (post-paid): mês anterior ao processamento
	targetMonth := monthStart
	if plan.Agenda.TermType == int(domain.TermTypePostPaid) {
		targetMonth = monthStart.AddDate(0, -1, 0)
	}

	targetStart := time.Date(targetMonth.Year(), targetMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	targetEnd := targetStart.AddDate(0, 1, -1)

	// Ajustar intervalo efetivo com vigência do plano
	effStart := targetStart
	if effStart.Before(plan.PlanStart) {
		effStart = plan.PlanStart
	}
	effEnd := targetEnd
	if plan.PlanEnd != nil && effEnd.After(*plan.PlanEnd) {
		effEnd = *plan.PlanEnd
	}

	calculatedQty := 0
	if !effStart.After(effEnd) {
		calculatedQty = s.CalculateSessionsCount(plan.PlanStart, effStart, effEnd, plan.Agenda.Recurrence, plan.Agenda.WeekDay)
	}

	// Limitar por monthly_service_limit se informado
	if plan.Agenda.MonthlyServiceLimit != nil && *plan.Agenda.MonthlyServiceLimit > 0 {
		if calculatedQty > *plan.Agenda.MonthlyServiceLimit {
			calculatedQty = *plan.Agenda.MonthlyServiceLimit
		}
	}

	// Ordenar itens pelo order_index
	items := make([]domain.PlanItem, len(plan.Items))
	copy(items, plan.Items)
	sort.Slice(items, func(i, j int) bool {
		return items[i].OrderIndex < items[j].OrderIndex
	})

	nItems := len(items)
	invoiceItems := make([]domain.InvoiceItem, nItems)
	var itemsTotalAmount float64

	baseQty := calculatedQty / nItems
	remainder := calculatedQty % nItems

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
		if item.Service != nil && item.Service.Description != nil {
			descCopy := *item.Service.Description
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

	// invoice.invoice_date = o primeiro dia do mês do período especificado ou a data de início, o que for menor
	invoicingDate := monthStart
	if plan.PlanStart.Before(invoicingDate) {
		invoicingDate = plan.PlanStart
	}

	// invoice.due_date = mês e ano do período de processamento, dia de plan_agenda.payment_day
	dim := DaysInMonth(monthStart.Year(), monthStart.Month())
	payDay := plan.Agenda.PaymentDay
	if payDay > dim {
		payDay = dim
	}
	dueDate := time.Date(monthStart.Year(), monthStart.Month(), payDay, 0, 0, 0, 0, time.UTC)

	invoice := &domain.Invoice{
		PlanID:             plan.ID,
		InvoicingDate:      invoicingDate,
		DueDate:            dueDate,
		Amount:             amount,
		ReferenceInvoiceID: nil,
		CreatedAt:          refDate,
		UpdatedAt:          refDate,
		DeletedAt:          nil,
		Items:              invoiceItems,
	}

	return invoice, nil
}

// CalculateSessionsCount calculates the number of sessions according to recurrence and weekday in [effStart, effEnd].
func (s *AgendaStrategy) CalculateSessionsCount(planStart, effStart, effEnd time.Time, recurrence, weekDay int) int {
	targetWeekday := time.Weekday((weekDay - 1) % 7) // 1: Sunday = 0, 2: Monday = 1 ...

	switch domain.Recurrence(recurrence) {
	case domain.RecurrenceWeekly: // 1: weekly (7 days interval)
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		count := 0
		d := firstOccurrence
		for !d.After(effEnd) {
			count++
			d = d.AddDate(0, 0, 7)
		}
		return count

	case domain.RecurrenceBiWeekly: // 2: bi-weekly (14 days interval anchored from plan start)
		offsetPlanStart := (int(targetWeekday) - int(planStart.Weekday()) + 7) % 7
		firstPlanStart := planStart.AddDate(0, 0, offsetPlanStart)
		count := 0
		d := firstPlanStart
		for !d.After(effEnd) {
			if !d.Before(effStart) {
				count++
			}
			d = d.AddDate(0, 0, 14)
		}
		return count

	case domain.RecurrenceMonthly: // 3: monthly
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		if !firstOccurrence.After(effEnd) {
			return 1
		}
		return 0

	default:
		// Default to weekly if unknown
		offset := (int(targetWeekday) - int(effStart.Weekday()) + 7) % 7
		firstOccurrence := effStart.AddDate(0, 0, offset)
		count := 0
		d := firstOccurrence
		for !d.After(effEnd) {
			count++
			d = d.AddDate(0, 0, 7)
		}
		return count
	}
}
