package invoicing

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
)

// PayPerViewStrategy handles calculation of invoices for pay-per-view plans (plan_type = 4).
type PayPerViewStrategy struct{}

// NewPayPerViewStrategy creates a new PayPerViewStrategy.
func NewPayPerViewStrategy() *PayPerViewStrategy {
	return &PayPerViewStrategy{}
}

// CalculateSession generates an invoice for a specific session of a pay-per-view plan.
func (s *PayPerViewStrategy) CalculateSession(
	plan *domain.Plan,
	item domain.PlanItem,
	session domain.Session,
	refDate time.Time,
) (*domain.Invoice, error) {
	if plan == nil {
		return nil, errors.New("plano nulo")
	}

	currentDate := time.Date(refDate.Year(), refDate.Month(), refDate.Day(), 0, 0, 0, 0, time.UTC)

	price := 0.0
	if item.Price != nil {
		price = *item.Price
	}

	sessionDateStr := session.SessionDate.Format("02/01/2006")
	var desc *string
	if item.Service != nil && item.Service.Description != nil && strings.TrimSpace(*item.Service.Description) != "" {
		d := fmt.Sprintf("%s - %s", strings.TrimSpace(*item.Service.Description), sessionDateStr)
		desc = &d
	} else {
		descCopy := sessionDateStr
		desc = &descCopy
	}

	itemSession := domain.InvoiceItemSession{
		SessionID: session.ID,
		CreatedAt: refDate,
	}

	invItem := domain.InvoiceItem{
		ServiceID:        item.ServiceID,
		Price:            price,
		Description:      desc,
		Quantity:         1,
		InvoicedQuantity: 1,
		RealizedQuantity: 1,
		CreatedAt:        refDate,
		UpdatedAt:        refDate,
		DeletedAt:        nil,
		ItemSessions:     []domain.InvoiceItemSession{itemSession},
	}

	amount := price
	if plan.Price != nil {
		amount = *plan.Price
	}

	dueDate := currentDate
	if plan.Package != nil {
		dueDate = plan.Package.PaymentDate
		if dueDate.Before(currentDate) {
			dueDate = currentDate
		}
	} else if plan.PayPerView != nil {
		dueDate = session.SessionDate.AddDate(0, 0, plan.PayPerView.PaymentTerm)
		if dueDate.Before(currentDate) {
			dueDate = currentDate
		}
	}

	invoicingDate := plan.PlanStart
	if session.SessionDate.After(plan.PlanStart) {
		invoicingDate = session.SessionDate
	}

	invoice := &domain.Invoice{
		PlanID:             plan.ID,
		InvoicingDate:      invoicingDate,
		DueDate:            dueDate,
		PeriodStart:        session.SessionDate,
		PeriodEnd:          session.SessionDate,
		Amount:             amount,
		ReferenceInvoiceID: nil,
		CreatedAt:          refDate,
		UpdatedAt:          refDate,
		DeletedAt:          nil,
		Items:              []domain.InvoiceItem{invItem},
	}

	return invoice, nil
}
