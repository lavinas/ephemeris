package dto

import (
	"strings"
	"testing"
	"time"

	"planner/internal/domain"
)

type mockInvoiceImportCSVRepo struct {
	mockPlanRepo
	invoiceExists bool
}

func (m *mockInvoiceImportCSVRepo) CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error) {
	return m.invoiceExists, nil
}

func TestInvoiceImportCSVRequest_Validate(t *testing.T) {
	baseRepo := newMockPlanRepo()
	baseRepo.customers["ana_flavia"] = &domain.Customer{ID: 101, Nickname: "ana_flavia", Name: "Ana Flávia", VendorID: 1}
	baseRepo.services[10] = &domain.Service{ID: 10, Name: "online/canto/60", VendorID: 1}

	repo := &mockInvoiceImportCSVRepo{
		mockPlanRepo: *baseRepo,
	}

	t.Run("valid single row", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aulas de canto,"98,00",2,`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		if err := req.Validate(repo); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}

		if len(req.GroupedInvoices) != 1 {
			t.Fatalf("expected 1 invoice, got %d", len(req.GroupedInvoices))
		}
		inv := req.GroupedInvoices[0]
		if inv.CustomerNickname != "ana_flavia" || len(inv.Items) != 1 {
			t.Errorf("unexpected invoice: %+v", inv)
		}
		if inv.Items[0].Price != 98.00 || inv.Items[0].Quantity != 2 {
			t.Errorf("unexpected item: %+v", inv.Items[0])
		}
	})

	t.Run("grouping multiple rows into 1 invoice with multiple items", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,01/03/2026,09/03/2026,online/canto/60,01/02/2026,28/02/2026,item 1,"110,00",2,520
ana_flavia,01/03/2026,09/03/2026,online/canto/60,01/02/2026,28/02/2026,item 2,"50,00",1,520`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		if err := req.Validate(repo); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}

		if len(req.GroupedInvoices) != 1 {
			t.Fatalf("expected 1 grouped invoice, got %d", len(req.GroupedInvoices))
		}
		inv := req.GroupedInvoices[0]
		if len(inv.Items) != 2 {
			t.Fatalf("expected 2 items in grouped invoice, got %d", len(inv.Items))
		}
		if inv.ReferenceInvoiceID == nil || *inv.ReferenceInvoiceID != 520 {
			t.Errorf("expected reference 520, got %v", inv.ReferenceInvoiceID)
		}
	})

	t.Run("divergent reference on same invoice group", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,01/03/2026,09/03/2026,online/canto/60,01/02/2026,28/02/2026,item 1,"110,00",2,520
ana_flavia,01/03/2026,09/03/2026,online/canto/60,01/02/2026,28/02/2026,item 2,"50,00",1,521`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "divergente") {
			t.Fatalf("expected divergent reference error, got: %v", err)
		}
	})

	t.Run("missing required column in header", func(t *testing.T) {
		csvData := `cliente,faturamento,serviço,inicio,fim,descrição,preço,quantidade
ana_flavia,02/01/2024,online/canto/60,01/12/2023,31/12/2023,aulas,"98,00",2`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "colunas obrigatórias") {
			t.Fatalf("expected header error, got: %v", err)
		}
	})

	t.Run("invalid customer in row returns first error", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
desconhecido,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aulas,"98,00",2,`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "não encontrado") {
			t.Fatalf("expected customer not found error, got: %v", err)
		}
	})

	t.Run("invalid price in row", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aulas,"0",2,`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "preço inválido") {
			t.Fatalf("expected price error, got: %v", err)
		}
	})

	t.Run("due date before invoice date", func(t *testing.T) {
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,05/01/2026,02/01/2026,online/canto/60,01/12/2025,31/12/2025,aulas,"100,00",1,`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(repo)
		if err == nil || !strings.Contains(err.Error(), "vencimento") {
			t.Fatalf("expected due date error, got: %v", err)
		}
	})

	t.Run("duplicate invoice in database", func(t *testing.T) {
		dupRepo := &mockInvoiceImportCSVRepo{
			mockPlanRepo:  *baseRepo,
			invoiceExists: true,
		}
		csvData := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
ana_flavia,02/01/2024,05/01/2026,online/canto/60,01/12/2023,31/12/2023,aulas,"98,00",2,`

		req := &InvoiceImportCSVRequest{CSVContent: csvData}
		err := req.Validate(dupRepo)
		if err == nil || !strings.Contains(err.Error(), "já existe uma fatura cadastrada") {
			t.Fatalf("expected duplicate error, got: %v", err)
		}
	})
}
