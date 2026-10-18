package service

import (
	"errors"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoiceUpdateRepoForService struct {
	mockPlanRepoForService
	invoices      map[int64]*domain.Invoice
	updatedInvoice *domain.Invoice
	failGet       bool
	failUpdate    bool
}

func newMockInvoiceUpdateRepoForService() *mockInvoiceUpdateRepoForService {
	return &mockInvoiceUpdateRepoForService{
		invoices: make(map[int64]*domain.Invoice),
	}
}

func (m *mockInvoiceUpdateRepoForService) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	if m.failGet {
		return nil, errors.New("not found")
	}
	if inv, ok := m.invoices[id]; ok && inv.DeletedAt == nil {
		return inv, nil
	}
	return nil, nil
}

func (m *mockInvoiceUpdateRepoForService) UpdateInvoice(invoice *domain.Invoice) error {
	if m.failUpdate {
		return errors.New("not found")
	}
	m.updatedInvoice = invoice
	m.invoices[invoice.ID] = invoice
	return nil
}

func TestInvoiceUpdate_Success(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	origDueDate := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	desc1 := "Consulta"
	desc2 := "Exame"
	inv := &domain.Invoice{
		ID:            1,
		DueDate:       origDueDate,
		InvoicingDate: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Amount:        250.0,
		Items: []domain.InvoiceItem{
			{ID: 10, InvoiceID: 1, ServiceID: 100, Description: &desc1, Price: 100.0, InvoicedQuantity: 1, RealizedQuantity: 1},
			{ID: 20, InvoiceID: 1, ServiceID: 200, Description: &desc2, Price: 75.0, InvoicedQuantity: 2, RealizedQuantity: 2},
		},
	}
	repo.invoices[1] = inv

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 1,
		DueDate:   "2026-11-05",
		Items: []dto.InvoiceItemUpdateRequest{
			{ID: 10, InvoicedQuantity: 3}, // 3 * 100 = 300
			{ID: 20, InvoicedQuantity: 1}, // 1 * 75 = 75
		},
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 200 {
		t.Fatalf("expected 200, got %d: %s", res.GetStatusCode(), res.GetMessage())
	}

	updateResp, ok := res.(*dto.InvoiceUpdateResponse)
	if !ok || updateResp.Invoice == nil {
		t.Fatalf("expected valid InvoiceUpdateResponse with InvoiceDTO")
	}

	// Verify total amount recalculated: 300 + 75 = 375
	if updateResp.Invoice.Amount != 375.0 {
		t.Errorf("expected recalculated amount 375.0, got %.2f", updateResp.Invoice.Amount)
	}

	// Verify DueDate updated
	if updateResp.Invoice.DueDate != "2026-11-05" {
		t.Errorf("expected due date 2026-11-05, got %s", updateResp.Invoice.DueDate)
	}

	// Verify repo received the updated model
	if repo.updatedInvoice == nil {
		t.Fatalf("repo did not receive updated invoice")
	}
	if repo.updatedInvoice.Amount != 375.0 {
		t.Errorf("repo invoice amount expected 375.0, got %.2f", repo.updatedInvoice.Amount)
	}
	if repo.updatedInvoice.DueDate.Format("2006-01-02") != "2026-11-05" {
		t.Errorf("repo invoice due date expected 2026-11-05, got %s", repo.updatedInvoice.DueDate.Format("2006-01-02"))
	}
	if repo.updatedInvoice.Items[0].InvoicedQuantity != 3 {
		t.Errorf("item 10 quantity expected 3, got %d", repo.updatedInvoice.Items[0].InvoicedQuantity)
	}
	if repo.updatedInvoice.Items[1].InvoicedQuantity != 1 {
		t.Errorf("item 20 quantity expected 1, got %d", repo.updatedInvoice.Items[1].InvoicedQuantity)
	}
}

func TestInvoiceUpdate_ValidationFailure(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 0,
		DueDate:   "invalid-date",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Errorf("expected 400 on validation failure, got %d", res.GetStatusCode())
	}
}

func TestInvoiceUpdate_NotFound(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 999,
		DueDate:   "2026-11-05",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 { // Validate checks repo and returns error
		t.Errorf("expected 400 when invoice not found in validation, got %d", res.GetStatusCode())
	}
}

func TestInvoiceUpdate_AlreadyInvoiced(t *testing.T) {
	repo := newMockInvoiceUpdateRepoForService()
	logger := &mockLogger{}

	billingID := int64(789)
	repo.invoices[1] = &domain.Invoice{
		ID:                 1,
		DueDate:            time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC),
		InvoicingDate:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Amount:             250.0,
		ReferenceInvoiceID: &billingID,
	}

	svc := NewInvoiceUpdate(repo, logger)
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: 1,
		DueDate:   "2026-11-05",
	}

	res := svc.Run(req)
	if res.GetStatusCode() != 400 {
		t.Errorf("expected 400 when invoice already invoiced, got %d with message %s", res.GetStatusCode(), res.GetMessage())
	}
	if repo.updatedInvoice != nil {
		t.Errorf("expected invoice not to be updated, but was updated")
	}
}
