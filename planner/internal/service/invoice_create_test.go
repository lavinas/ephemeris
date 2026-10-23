package service

import (
	"testing"

	"planner/internal/domain"
	"planner/internal/dto"
)

func TestInvoiceCreate_Success(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	repo.customers["ana_flavia"] = &domain.Customer{
		ID:       10,
		Nickname: "ana_flavia",
		Name:     "Ana Flavia",
		VendorID: 1,
	}
	repo.services[10] = &domain.Service{
		ID:       10,
		Name:     "online/canto/60",
		VendorID: 1,
	}

	svc := NewInvoiceCreate(repo, logger)

	refID := int64(520)
	req := &dto.InvoiceCreateRequest{
		Nickname:           "ana_flavia",
		InvoiceDate:        "2026-10-22",
		DueDate:            "2026-10-25",
		PeriodStart:        "2026-10-01",
		PeriodEnd:          "2026-10-31",
		ReferenceInvoiceID: &refID,
		Items: []dto.InvoiceItemCreateRequest{
			{
				ServiceID:   10,
				Price:       110.0,
				Quantity:    2,
				Description: "Aula de canto 60 min",
			},
			{
				ServiceID:   10,
				Price:       50.0,
				Quantity:    1,
				Description: "Material extra",
			},
		},
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 201 {
		t.Fatalf("expected status 201, got %d: %s", out.GetStatusCode(), out.GetMessage())
	}

	resp, ok := out.(*dto.InvoiceCreateResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", out)
	}

	if resp.Invoice == nil {
		t.Fatalf("expected non-nil invoice in response")
	}

	expectedAmount := 110.0*2 + 50.0*1 // 270.0
	if resp.Invoice.Amount != expectedAmount {
		t.Errorf("expected amount %.2f, got %.2f", expectedAmount, resp.Invoice.Amount)
	}

	if len(resp.Invoice.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(resp.Invoice.Items))
	}

	if len(repo.invoices) != 1 {
		t.Fatalf("expected 1 invoice saved in repo, got %d", len(repo.invoices))
	}
	saved := repo.invoices[0]
	if saved.Amount != expectedAmount {
		t.Errorf("expected saved invoice amount %.2f, got %.2f", expectedAmount, saved.Amount)
	}
}

func TestInvoiceCreate_ValidationError(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	svc := NewInvoiceCreate(repo, logger)
	req := &dto.InvoiceCreateRequest{
		Nickname: "",
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 400 {
		t.Fatalf("expected status 400, got %d", out.GetStatusCode())
	}
}
