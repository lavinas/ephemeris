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
	if saved.PlanID == 0 {
		t.Fatalf("expected invoice to have a PlanID associated")
	}

	plan := repo.plans[saved.PlanID]
	if plan == nil {
		t.Fatalf("expected manual plan to be created with ID %d", saved.PlanID)
	}
	if plan.PlanType != int(domain.PlanTypeManual) {
		t.Errorf("expected plan type %d (Manual), got %d", domain.PlanTypeManual, plan.PlanType)
	}
	if plan.PlanStart.Format("2006-01-02") != "2026-10-22" {
		t.Errorf("expected plan start 2026-10-22, got %s", plan.PlanStart.Format("2006-01-02"))
	}
	if plan.PlanEnd == nil || plan.PlanEnd.Format("2006-01-02") != "2026-10-22" {
		t.Errorf("expected plan end 2026-10-22, got %v", plan.PlanEnd)
	}
	if len(plan.Items) != 0 {
		t.Errorf("expected manual plan to have 0 items, got %d", len(plan.Items))
	}
}

func TestInvoiceCreate_ManualPlanResolution(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	repo.customers["joao"] = &domain.Customer{
		ID:       20,
		Nickname: "joao",
		Name:     "Joao Silva",
		VendorID: 1,
	}
	repo.services[1] = &domain.Service{
		ID:       1,
		Name:     "canto",
		VendorID: 1,
	}

	svc := NewInvoiceCreate(repo, logger)

	// 1. Create first invoice on 2026-10-15: no plan exists, creates manual plan [2026-10-15, 2026-10-15]
	req1 := &dto.InvoiceCreateRequest{
		Nickname:    "joao",
		InvoiceDate: "2026-10-15",
		DueDate:     "2026-10-20",
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Items: []dto.InvoiceItemCreateRequest{
			{ServiceID: 1, Price: 100, Quantity: 1},
		},
	}
	res1 := svc.Run(req1)
	if res1.GetStatusCode() != 201 {
		t.Fatalf("first invoice creation failed: %s", res1.GetMessage())
	}
	if len(repo.plans) != 1 {
		t.Fatalf("expected 1 manual plan created, got %d", len(repo.plans))
	}
	inv1 := repo.invoices[0]
	firstPlanID := inv1.PlanID

	// 2. Create second invoice on same date (2026-10-15): encompasses invoiceDate, should reuse firstPlanID
	req2 := &dto.InvoiceCreateRequest{
		Nickname:    "joao",
		InvoiceDate: "2026-10-15",
		DueDate:     "2026-10-22",
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Items: []dto.InvoiceItemCreateRequest{
			{ServiceID: 1, Price: 100, Quantity: 1},
		},
	}
	res2 := svc.Run(req2)
	if res2.GetStatusCode() != 201 {
		t.Fatalf("second invoice creation failed: %s", res2.GetMessage())
	}
	if len(repo.plans) != 1 {
		t.Errorf("expected plan to be reused (count=1), but plan count is %d", len(repo.plans))
	}
	inv2 := repo.invoices[1]
	if inv2.PlanID != firstPlanID {
		t.Errorf("expected inv2 to reuse plan %d, got %d", firstPlanID, inv2.PlanID)
	}

	// 3. Create third invoice on 2026-11-01: outside [2026-10-15, 2026-10-15], creates a new manual plan
	req3 := &dto.InvoiceCreateRequest{
		Nickname:    "joao",
		InvoiceDate: "2026-11-01",
		DueDate:     "2026-11-05",
		PeriodStart: "2026-11-01",
		PeriodEnd:   "2026-11-30",
		Items: []dto.InvoiceItemCreateRequest{
			{ServiceID: 1, Price: 120, Quantity: 1},
		},
	}
	res3 := svc.Run(req3)
	if res3.GetStatusCode() != 201 {
		t.Fatalf("third invoice creation failed: %s", res3.GetMessage())
	}
	if len(repo.plans) != 2 {
		t.Fatalf("expected 2 manual plans created, got %d", len(repo.plans))
	}
	inv3 := repo.invoices[2]
	if inv3.PlanID == firstPlanID {
		t.Errorf("expected inv3 to get a new plan, but got %d", inv3.PlanID)
	}
	plan2 := repo.plans[inv3.PlanID]
	if plan2.PlanStart.Format("2006-01-02") != "2026-11-01" || plan2.PlanEnd.Format("2006-01-02") != "2026-11-01" {
		t.Errorf("expected plan2 dates 2026-11-01, got start=%s, end=%v", plan2.PlanStart, plan2.PlanEnd)
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
