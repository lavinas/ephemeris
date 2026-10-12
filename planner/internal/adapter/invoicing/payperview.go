package invoicing

import (
	"errors"
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// PayPerViewStrategy handles invoicing for pay-per-view plans (plan_type = 4).
type PayPerViewStrategy struct {
	repo   port.Repository
	logger port.Logger
}

// NewPayPerViewStrategy creates a new PayPerViewStrategy adapter.
func NewPayPerViewStrategy(repo port.Repository, logger port.Logger) *PayPerViewStrategy {
	return &PayPerViewStrategy{
		repo:   repo,
		logger: logger,
	}
}

// Process executes the pay-per-view invoicing strategy.
func (s *PayPerViewStrategy) Process(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	if plan == nil {
		return errors.New("plano nulo")
	}
	if len(plan.Items) == 0 {
		return fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Faturamento ocorre para cada sessão do customer e do serviço correspondente dentro do período informado
	for _, item := range plan.Items {
		sessions, err := s.repo.FindBillableSessions(plan.CustomerID, item.ServiceID, startDate, endDate)
		if err != nil {
			return fmt.Errorf("plano %d: erro ao buscar sessões do serviço %d: %w", plan.ID, item.ServiceID, err)
		}

		for _, session := range sessions {
			existingInvoice, err := s.repo.GetInvoiceForSession(session.ID)
			if err != nil {
				return fmt.Errorf("plano %d: erro ao verificar fatura para a sessão %d: %w", plan.ID, session.ID, err)
			}

			if existingInvoice != nil {
				if !reprocess {
					s.logger.IPrintf(2, "Plano pay-per-view %d: sessão %d já faturada (invoice id: %d). Pulando sessão.",
						plan.ID, session.ID, existingInvoice.ID)
					continue
				}

				// Reprocess: soft delete fatura existente
				s.logger.IPrintf(2, "Reprocessando plano pay-per-view %d para sessão %d: excluindo fatura anterior (id: %d).",
					plan.ID, session.ID, existingInvoice.ID)
				if err := s.repo.DeleteInvoice(existingInvoice.ID); err != nil {
					return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d da sessão %d: %w",
						plan.ID, existingInvoice.ID, session.ID, err)
				}
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

			itemSession := domain.InvoiceItemSession{
				SessionID: session.ID,
				CreatedAt: now,
			}

			invItem := domain.InvoiceItem{
				ServiceID:        item.ServiceID,
				Price:            price,
				Description:      desc,
				InvoicedQuantity: 1,
				RealizedQuantity: 1,
				CreatedAt:        now,
				UpdatedAt:        now,
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
				Amount:             amount,
				ReferenceInvoiceID: nil,
				CreatedAt:          now,
				UpdatedAt:          now,
				DeletedAt:          nil,
				Items:              []domain.InvoiceItem{invItem},
			}

			if err := s.repo.SaveInvoice(invoice); err != nil {
				return fmt.Errorf("plano %d: erro ao salvar fatura pay-per-view para sessão %d: %w",
					plan.ID, session.ID, err)
			}

			s.logger.IPrintf(2, "Fatura gerada com sucesso para plano pay-per-view %d, sessão %d (invoice id: %d, valor: %.2f)",
				plan.ID, session.ID, invoice.ID, invoice.Amount)
		}
	}

	return nil
}
