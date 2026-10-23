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
