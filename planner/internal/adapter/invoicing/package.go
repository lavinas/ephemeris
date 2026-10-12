package invoicing

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// PackageStrategy handles invoicing for package plans (plan_type = 2).
type PackageStrategy struct {
	repo   port.Repository
	logger port.Logger
}

// NewPackageStrategy creates a new PackageStrategy adapter.
func NewPackageStrategy(repo port.Repository, logger port.Logger) *PackageStrategy {
	return &PackageStrategy{
		repo:   repo,
		logger: logger,
	}
}

// Process executes the package invoicing strategy.
func (s *PackageStrategy) Process(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error {
	if plan == nil {
		return errors.New("plano nulo")
	}
	if plan.Package == nil {
		return fmt.Errorf("plano %d: dados de pacote não encontrados", plan.ID)
	}
	if len(plan.Items) == 0 {
		return fmt.Errorf("plano %d: não possui itens cadastrados", plan.ID)
	}

	totalQuantity := plan.Package.Quantity
	if totalQuantity <= 0 {
		return fmt.Errorf("plano %d: quantidade de aulas do pacote deve ser maior que zero", plan.ID)
	}

	if len(plan.Items) > totalQuantity {
		return fmt.Errorf("plano %d: número de itens (%d) não pode ser maior que a quantidade de aulas (%d)",
			plan.ID, len(plan.Items), totalQuantity)
	}

	// Verificar se já existe fatura para o plano
	existingInvoices, err := s.repo.FindInvoices(map[string]interface{}{
		"plan_id = ?":            plan.ID,
		"deleted_at IS NULL": nil,
	})
	if err != nil {
		return fmt.Errorf("plano %d: erro ao verificar faturas existentes: %w", plan.ID, err)
	}

	if len(existingInvoices) > 0 {
		if !reprocess {
			s.logger.IPrintf(2, "Plano pacote %d já possui fatura gerada (id: %d). Pulando processamento.",
				plan.ID, existingInvoices[0].ID)
			return nil
		}

		// Reprocess: soft delete faturas existentes
		for _, inv := range existingInvoices {
			s.logger.IPrintf(2, "Reprocessando plano pacote %d: excluindo fatura anterior (id: %d).", plan.ID, inv.ID)
			if err := s.repo.DeleteInvoice(inv.ID); err != nil {
				return fmt.Errorf("plano %d: erro ao excluir fatura anterior %d: %w", plan.ID, inv.ID, err)
			}
		}
	}

	// Ordenar itens pelo order_index
	items := make([]domain.PlanItem, len(plan.Items))
	copy(items, plan.Items)
	sort.Slice(items, func(i, j int) bool {
		return items[i].OrderIndex < items[j].OrderIndex
	})

	// Distribuir quantidade proporcionalmente pelo order_index
	nItems := len(items)
	baseQty := totalQuantity / nItems
	remainder := totalQuantity % nItems

	now := time.Now()
	currentDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	dueDate := plan.Package.PaymentDate
	if dueDate.Before(currentDate) {
		dueDate = currentDate
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
			CreatedAt:        now,
			UpdatedAt:        now,
			DeletedAt:        nil,
		}

		itemsTotalAmount += price * float64(invoicedQty)
	}

	amount := itemsTotalAmount
	if plan.Price != nil {
		amount = *plan.Price
	}

	invoice := &domain.Invoice{
		PlanID:             plan.ID,
		InvoicingDate:      plan.PlanStart,
		DueDate:            dueDate,
		Amount:             amount,
		ReferenceInvoiceID: nil,
		CreatedAt:          now,
		UpdatedAt:          now,
		DeletedAt:          nil,
		Items:              invoiceItems,
	}

	if err := s.repo.SaveInvoice(invoice); err != nil {
		return fmt.Errorf("plano %d: erro ao salvar fatura pacote: %w", plan.ID, err)
	}

	s.logger.IPrintf(2, "Fatura gerada com sucesso para plano pacote %d (invoice id: %d, valor: %.2f)",
		plan.ID, invoice.ID, invoice.Amount)
	return nil
}
