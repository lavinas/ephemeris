package invoicing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
)

// NotebookStrategy handles calculation of invoices for notebook plans (plan_type = 3).
type NotebookStrategy struct{}

// NewNotebookStrategy creates a new NotebookStrategy.
func NewNotebookStrategy() *NotebookStrategy {
	return &NotebookStrategy{}
}

// IsActiveInMonth checks if the plan is active for invoicing in the given calculation month.
// Since notebook plans are always post-paid, the active period evaluated is the previous month (where sessions occurred).
func (s *NotebookStrategy) IsActiveInMonth(plan *domain.Plan, monthStart time.Time) bool {
	prevMonthStart, prevMonthEnd := s.PreviousMonthPeriod(monthStart)
	return !plan.PlanStart.After(prevMonthEnd) && (plan.PlanEnd == nil || !plan.PlanEnd.Before(prevMonthStart))
}

// PreviousMonthPeriod returns the start and end of the previous month for billing sessions.
func (s *NotebookStrategy) PreviousMonthPeriod(monthStart time.Time) (time.Time, time.Time) {
	prevMonthStart := monthStart.AddDate(0, -1, 0)
	prevMonthEnd := monthStart.AddDate(0, 0, -1)
	return prevMonthStart, prevMonthEnd
}

// CalculateMonth calculates the invoice for a notebook plan based on sessions of the previous month.
// sessionsByServiceID maps serviceID to the slice of billable sessions realized in the previous month.
// Returns (nil, nil) if the plan is inactive in that month, or if there are no sessions and the plan has no fixed price.
func (s *NotebookStrategy) CalculateMonth(
	plan *domain.Plan,
	monthStart time.Time,
	sessionsByServiceID map[int64][]domain.Session,
	refDate time.Time,
) (*domain.Invoice, error) {
	if plan == nil {
		return nil, errors.New("plano nulo")
	}
	if len(plan.Items) == 0 {
		return nil, fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	if !s.IsActiveInMonth(plan, monthStart) {
		return nil, nil
	}

	currentDate := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)
	invoiceItems := make([]domain.InvoiceItem, 0, len(plan.Items))
	var itemsTotalAmount float64
	var totalSessionsCount int

	prevMonthStart, prevMonthEnd := s.PreviousMonthPeriod(monthStart)
	monthYearSuffix := fmt.Sprintf("%s de %d", MonthNameInPortuguese(prevMonthStart.Month()), prevMonthStart.Year())

	for _, item := range plan.Items {
		sessions := sessionsByServiceID[item.ServiceID]
		var billableSessions []domain.Session
		for _, sess := range sessions {
			if isCancelledNotCharge(sess.SessionStatus) {
				continue
			}
			billableSessions = append(billableSessions, sess)
		}

		invoicedQty := len(billableSessions)
		totalSessionsCount += invoicedQty

		price := 0.0
		if item.Price != nil {
			price = *item.Price
		}

		var desc *string
		if item.Service != nil && item.Service.Description != nil && strings.TrimSpace(*item.Service.Description) != "" {
			d := fmt.Sprintf("%s - %s", strings.TrimSpace(*item.Service.Description), monthYearSuffix)
			desc = &d
		} else {
			descCopy := monthYearSuffix
			desc = &descCopy
		}

		itemSessions := make([]domain.InvoiceItemSession, invoicedQty)
		for idx, sess := range billableSessions {
			itemSessions[idx] = domain.InvoiceItemSession{
				SessionID: sess.ID,
				CreatedAt: refDate,
			}
		}

		invItem := domain.InvoiceItem{
			ServiceID:        item.ServiceID,
			Price:            price,
			Description:      desc,
			Quantity:         invoicedQty,
			InvoicedQuantity: invoicedQty,
			RealizedQuantity: invoicedQty,
			CreatedAt:        refDate,
			UpdatedAt:        refDate,
			DeletedAt:        nil,
			ItemSessions:     itemSessions,
		}

		itemsTotalAmount += price * float64(invoicedQty)
		invoiceItems = append(invoiceItems, invItem)
	}

	// Se não há sessões e o plano não tem preço fixo, não gera fatura
	if totalSessionsCount == 0 && (plan.Price == nil || *plan.Price == 0) {
		return nil, nil
	}

	// Definir vencimento
	dueDate := currentDate
	if plan.Package != nil {
		dueDate = plan.Package.PaymentDate
		if dueDate.Before(currentDate) {
			dueDate = currentDate
		}
	} else if plan.Notebook != nil && plan.Notebook.PaymentDay > 0 {
		dim := DaysInMonth(monthStart.Year(), monthStart.Month())
		payDay := plan.Notebook.PaymentDay
		if payDay > dim {
			payDay = dim
		}
		dueDate = time.Date(monthStart.Year(), monthStart.Month(), payDay, 0, 0, 0, 0, time.UTC)
		if dueDate.Before(currentDate) {
			dueDate = currentDate
		}
	}

	invoicingDate := plan.PlanStart
	if monthStart.After(plan.PlanStart) {
		invoicingDate = monthStart
	}

	amount := itemsTotalAmount
	if plan.Price != nil {
		amount = *plan.Price
	}

	invoice := &domain.Invoice{
		PlanID:             plan.ID,
		InvoicingDate:      invoicingDate,
		DueDate:            dueDate,
		PeriodStart:        prevMonthStart,
		PeriodEnd:          prevMonthEnd,
		Amount:             amount,
		ReferenceInvoiceID: nil,
		CreatedAt:          refDate,
		UpdatedAt:          refDate,
		DeletedAt:          nil,
		Items:              invoiceItems,
	}

	return invoice, nil
}

func isCancelledNotCharge(status string) bool {
	trimmed := strings.TrimSpace(status)
	return trimmed == string(domain.StatusCancelNotCharge) ||
		trimmed == "Cancelada/Não Cobrar" ||
		strings.EqualFold(trimmed, "cancelada_nao_cobrar") ||
		strings.EqualFold(trimmed, "cancelada/não cobrar") ||
		strings.EqualFold(trimmed, "cancelada/nao cobrar")
}
