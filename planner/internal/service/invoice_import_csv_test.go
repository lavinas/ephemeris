package service

import (
	"os"
	"strings"
	"testing"

	"planner/internal/domain"
	"planner/internal/dto"
)

func TestInvoiceImportCSV_Success(t *testing.T) {
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

	svc := NewInvoiceImportCSV(repo, logger)

	csvContent := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aula 1,"98,00",2,
ana_flavia,01/02/2024,05/02/2026,online/canto/60,01/01/2024,31/01/2024,aula 2,"98,00",2,
ana_flavia,01/03/2024,05/03/2026,online/canto/60,01/02/2024,29/02/2024,aula 3 item A,"110,00",2,520
ana_flavia,01/03/2024,05/03/2026,online/canto/60,01/02/2024,29/02/2024,aula 3 item B,"50,00",1,520`

	req := &dto.InvoiceImportCSVRequest{
		CSVContent: csvContent,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", out.GetStatusCode(), out.GetMessage())
	}

	resp, ok := out.(*dto.InvoiceImportCSVResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", out)
	}

	// 4 lines in CSV: 2 distinct dates, plus 2 lines grouped into 1 invoice -> 3 invoices total
	if resp.InvoicesImported != 3 {
		t.Errorf("expected 3 invoices imported, got %d", resp.InvoicesImported)
	}
	if resp.ItemsImported != 4 {
		t.Errorf("expected 4 items imported, got %d", resp.ItemsImported)
	}

	expectedTotal := (98.0 * 2) + (98.0 * 2) + (110.0*2 + 50.0*1)
	if resp.TotalAmount != expectedTotal {
		t.Errorf("expected total %.2f, got %.2f", expectedTotal, resp.TotalAmount)
	}

	if len(repo.invoices) != 3 {
		t.Fatalf("expected 3 invoices saved in repo, got %d", len(repo.invoices))
	}

	if len(repo.plans) != 1 {
		t.Fatalf("expected 1 manual plan created for customer ana_flavia, got %d", len(repo.plans))
	}

	var createdPlan *domain.Plan
	for _, p := range repo.plans {
		createdPlan = p
		break
	}

	if createdPlan.PlanType != int(domain.PlanTypeManual) {
		t.Errorf("expected plan type %d (Manual), got %d", domain.PlanTypeManual, createdPlan.PlanType)
	}
	if createdPlan.PlanStart.Format("2006-01-02") != "2024-01-02" {
		t.Errorf("expected plan start 2024-01-02, got %s", createdPlan.PlanStart.Format("2006-01-02"))
	}
	if createdPlan.PlanEnd == nil || createdPlan.PlanEnd.Format("2006-01-02") != "2024-03-01" {
		t.Errorf("expected plan end 2024-03-01, got %v", createdPlan.PlanEnd)
	}

	for _, inv := range repo.invoices {
		if inv.PlanID != createdPlan.ID {
			t.Errorf("expected invoice plan_id %d, got %d", createdPlan.ID, inv.PlanID)
		}
	}
}

func TestInvoiceImportCSV_MultipleCustomersManualPlans(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	repo.customers["cliente_a"] = &domain.Customer{ID: 101, Nickname: "cliente_a", VendorID: 1}
	repo.customers["cliente_b"] = &domain.Customer{ID: 102, Nickname: "cliente_b", VendorID: 1}
	repo.services[1] = &domain.Service{ID: 1, Name: "aula", VendorID: 1}

	svc := NewInvoiceImportCSV(repo, logger)

	csvContent := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
cliente_a,10/01/2026,15/01/2026,aula,01/01/2026,10/01/2026,item a1,"100,00",1,
cliente_a,20/02/2026,25/02/2026,aula,01/02/2026,20/02/2026,item a2,"100,00",1,
cliente_b,05/03/2026,10/03/2026,aula,01/03/2026,05/03/2026,item b1,"150,00",1,
cliente_b,15/05/2026,20/05/2026,aula,01/05/2026,15/05/2026,item b2,"150,00",1,`

	req := &dto.InvoiceImportCSVRequest{CSVContent: csvContent}
	out := svc.Run(req)
	if out.GetStatusCode() != 200 {
		t.Fatalf("import failed: %s", out.GetMessage())
	}

	if len(repo.plans) != 2 {
		t.Fatalf("expected 2 manual plans created (1 per customer), got %d", len(repo.plans))
	}

	// Verify plan for cliente_a: 2026-01-10 to 2026-02-20
	var planA, planB *domain.Plan
	for _, p := range repo.plans {
		if p.CustomerID == 101 {
			planA = p
		} else if p.CustomerID == 102 {
			planB = p
		}
	}

	if planA == nil {
		t.Fatalf("plan for cliente_a not found")
	}
	if planA.PlanType != int(domain.PlanTypeManual) {
		t.Errorf("planA expected Manual (5), got %d", planA.PlanType)
	}
	if planA.PlanStart.Format("2006-01-02") != "2026-01-10" || planA.PlanEnd.Format("2006-01-02") != "2026-02-20" {
		t.Errorf("planA unexpected dates: start=%s, end=%v", planA.PlanStart.Format("2006-01-02"), planA.PlanEnd)
	}

	if planB == nil {
		t.Fatalf("plan for cliente_b not found")
	}
	if planB.PlanType != int(domain.PlanTypeManual) {
		t.Errorf("planB expected Manual (5), got %d", planB.PlanType)
	}
	if planB.PlanStart.Format("2006-01-02") != "2026-03-05" || planB.PlanEnd.Format("2006-01-02") != "2026-05-15" {
		t.Errorf("planB unexpected dates: start=%s, end=%v", planB.PlanStart.Format("2006-01-02"), planB.PlanEnd)
	}
}

func TestInvoiceImportCSV_WithExampleCSVFile(t *testing.T) {
	// Read example.csv from planner/tmp/example.csv
	data, err := os.ReadFile("../../tmp/example.csv")
	if err != nil {
		t.Skipf("skipping example.csv test: %v", err)
	}

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

	svc := NewInvoiceImportCSV(repo, logger)
	req := &dto.InvoiceImportCSVRequest{
		CSVContent: string(data),
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 200 {
		t.Fatalf("expected status 200 importing example.csv, got %d: %s", out.GetStatusCode(), out.GetMessage())
	}

	resp, ok := out.(*dto.InvoiceImportCSVResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", out)
	}

	// example.csv has 33 data lines (all with different months or distinct periods)
	if resp.InvoicesImported != 33 {
		t.Errorf("expected 33 invoices imported, got %d", resp.InvoicesImported)
	}
	if resp.ItemsImported != 33 {
		t.Errorf("expected 33 items imported, got %d", resp.ItemsImported)
	}
	if len(repo.invoices) != 33 {
		t.Errorf("expected 33 invoices saved in repo, got %d", len(repo.invoices))
	}
}

func TestInvoiceImportCSV_ValidationErrorAtomic(t *testing.T) {
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

	svc := NewInvoiceImportCSV(repo, logger)

	// Second line has negative price
	csvContent := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aula 1,"98,00",2,
ana_flavia,01/02/2024,05/02/2026,online/canto/60,01/01/2024,31/01/2024,aula 2,"-10,00",2,`

	req := &dto.InvoiceImportCSVRequest{
		CSVContent: csvContent,
	}

	out := svc.Run(req)
	if out.GetStatusCode() != 400 {
		t.Fatalf("expected status 400, got %d", out.GetStatusCode())
	}

	if !strings.Contains(out.GetMessage(), "linha 3") {
		t.Errorf("expected error message to mention line 3, got: %s", out.GetMessage())
	}

	// Atomic: NO invoices should have been saved!
	if len(repo.invoices) != 0 {
		t.Fatalf("expected 0 invoices saved in repo, got %d", len(repo.invoices))
	}
}
