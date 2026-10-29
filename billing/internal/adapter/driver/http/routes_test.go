package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
)

type dummyLogger struct{}

func (d *dummyLogger) IPrintf(level int, format string, v ...interface{}) {}
func (d *dummyLogger) Close()                                             {}

type dummyRepo struct {
	invoices map[int64]*domain.Invoice
}

func newDummyRepo() *dummyRepo {
	now := time.Now()
	email := "test@example.com"
	extCode := "EXT-99"
	return &dummyRepo{
		invoices: map[int64]*domain.Invoice{
			1: {
				ID:          1,
				VendorID:    1,
				CustomerID:  1,
				Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_teste", Email: &email},
				Amount:      250.0,
				InvoiceDate: now,
				DueDate:     now.AddDate(0, 0, 15),
				InvoiceItems: []domain.InvoiceItem{
					{ID: 1, Description: "Aula de Canto", Quantity: 2, Price: 125.0},
				},
			},
			5: {
				ID:           5,
				VendorID:     1,
				CustomerID:   1,
				Customer:     domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_teste", Email: &email},
				Amount:       300.0,
				InvoiceDate:  now,
				DueDate:      now.AddDate(0, 0, 15),
				ExternalCode: &extCode,
				InvoiceItems: []domain.InvoiceItem{
					{ID: 2, Description: "Aula de Piano", Quantity: 1, Price: 300.0},
				},
			},
		},
	}
}

func (d *dummyRepo) Save(model interface{}) error {
	if inv, ok := model.(*domain.Invoice); ok {
		copied := *inv
		d.invoices[inv.ID] = &copied
	}
	return nil
}
func (d *dummyRepo) BeginTransaction() error    { return nil }
func (d *dummyRepo) CommitTransaction() error   { return nil }
func (d *dummyRepo) RollbackTransaction() error { return nil }
func (d *dummyRepo) Close() error               { return nil }

func (d *dummyRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return []domain.Customer{
		{ID: 1, VendorID: 1, Nickname: "cliente_teste", Name: "Cliente Teste"},
	}, nil
}

func (d *dummyRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return &domain.Customer{ID: 1, VendorID: 1, Nickname: nickname, Name: "Cliente Teste"}, nil
}

func (d *dummyRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return []domain.Vendor{
		{ID: 1, Nickname: "estudio_amelia", LegalName: "Estudio Amelia"},
	}, nil
}

func (d *dummyRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	return &domain.Vendor{ID: 1, Nickname: nickname, LegalName: "Estudio Amelia"}, nil
}

func (d *dummyRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "estudio_amelia", LegalName: "Estudio Amelia"}, nil
}

func (d *dummyRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, sentDate, receiptDate,
	taxDate, cancellationDate, externalCode *string) ([]domain.Invoice, error) {
	now := time.Now()
	return []domain.Invoice{
		{
			ID:          1,
			CustomerID:  1,
			Customer:    domain.Customer{ID: 1, Nickname: "cliente_teste"},
			Amount:      250.0,
			InvoiceDate: now,
			DueDate:     now.AddDate(0, 0, 15),
			InvoiceItems: []domain.InvoiceItem{
				{ID: 1, Description: "Aula de Canto", Quantity: 2, Price: 125.0},
			},
		},
	}, nil
}

func (d *dummyRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (d *dummyRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (d *dummyRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	if inv, ok := d.invoices[id]; ok {
		copied := *inv
		return &copied, nil
	}
	now := time.Now()
	email := "test@example.com"
	newInv := &domain.Invoice{
		ID:          id,
		CustomerID:  1,
		Customer:    domain.Customer{ID: 1, VendorID: 1, Nickname: "cliente_teste", Email: &email},
		Amount:      250.0,
		InvoiceDate: now,
		DueDate:     now.AddDate(0, 0, 15),
		InvoiceItems: []domain.InvoiceItem{
			{ID: 1, Description: "Aula de Canto", Quantity: 2, Price: 125.0},
		},
	}
	d.invoices[id] = newInv
	copied := *newInv
	return &copied, nil
}

func (d *dummyRepo) GetInvoiceByExternalCode(vendorID int64, externalCode string) (*domain.Invoice, error) {
	for _, inv := range d.invoices {
		if inv.ExternalCode != nil && *inv.ExternalCode == externalCode {
			copied := *inv
			return &copied, nil
		}
	}
	return nil, nil
}

func (d *dummyRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}

func (d *dummyRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}

func (d *dummyRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

type dummyTaxer struct{}

func (d *dummyTaxer) GetEmission(emission *domain.Emission, builder *strings.Builder) error {
	return nil
}
func (d *dummyTaxer) ClearEmission(source string) (map[int64]*domain.EmissionItem, error) {
	return nil, nil
}

type dummyPixer struct{}

func (d *dummyPixer) Get(request port.InDTO) (string, string, error) {
	return "pix_code", "pix_base64", nil
}

type dummyIssuer struct{}

func (d *dummyIssuer) GetBase64(data port.InDTO, html_pdf string) ([]byte, error) {
	return []byte("pdf"), nil
}
func (d *dummyIssuer) SendMail(data port.InDTO, subject, filename, html_pdf, html_email string) error {
	return nil
}

func TestBillingRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../../.."); err != nil {
		t.Fatalf("could not chdir to billing root: %v", err)
	}

	templateData := []byte(`
{{define "index"}}<html><body>Invoices</body></html>{{end}}
{{define "fragment"}}<div>Fragment</div>{{end}}
{{define "tabela"}}<div>Tabela</div>{{end}}
{{define "formulario_cadastro"}}<form>Cadastro</form>{{end}}
{{define "linha_invoice"}}<div>Linha</div>{{end}}
{{define "linha_invoice_edit"}}<form>Edit</form>{{end}}
{{define "linha_invoice_deletar_aviso"}}<div>Aviso Deletar</div>{{end}}
{{define "linha_invoice_pagamento_aviso"}}<div>Aviso Pagar</div>{{end}}
`)
	routes, err := NewRoutes(newDummyRepo(), &dummyLogger{}, &dummyTaxer{}, &dummyPixer{}, &dummyIssuer{}, templateData)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		url        string
		body       url.Values
		jsonBody   string
		headers    map[string]string
		wantStatus int
	}{
		{name: "GET /", method: http.MethodGet, url: "/", wantStatus: http.StatusOK},
		{name: "GET /html/ping", method: http.MethodGet, url: "/html/ping", wantStatus: http.StatusOK},
		{name: "GET /ping", method: http.MethodGet, url: "/ping", wantStatus: http.StatusOK},
		{name: "GET /html/invoices", method: http.MethodGet, url: "/html/invoices", wantStatus: http.StatusOK},
		{name: "GET /html/invoices HTMX", method: http.MethodGet, url: "/html/invoices", headers: map[string]string{"HX-Request": "true"}, wantStatus: http.StatusOK},
		{name: "GET /html/invoices/novo", method: http.MethodGet, url: "/html/invoices/novo", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/tabela", method: http.MethodPost, url: "/html/invoices/tabela", body: url.Values{"page": {"1"}}, wantStatus: http.StatusOK},
		{name: "POST /html/invoices/tabela with overdue=true", method: http.MethodPost, url: "/html/invoices/tabela", body: url.Values{"page": {"1"}, "overdue": {"true"}}, wantStatus: http.StatusOK},
		{name: "POST /html/invoices/tabela with overdue=false", method: http.MethodPost, url: "/html/invoices/tabela", body: url.Values{"page": {"1"}, "overdue": {"false"}}, wantStatus: http.StatusOK},
		{name: "GET /invoice/api/list with overdue", method: http.MethodGet, url: "/invoice/api/list?vendor=estudio_amelia&page=1&page_size=10&overdue=true", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/tabela/reset", method: http.MethodPost, url: "/html/invoices/tabela/reset", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/bloco/limpar", method: http.MethodGet, url: "/html/invoices/bloco/limpar", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/editar", method: http.MethodGet, url: "/html/invoices/editar?id=1", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/editar external forbidden", method: http.MethodGet, url: "/html/invoices/editar?id=5", wantStatus: http.StatusForbidden},
		{name: "POST /html/invoices/atualizar", method: http.MethodPost, url: "/html/invoices/atualizar?id=1", body: url.Values{"edit_invoicing": {"2026-09-01"}, "edit_due": {"2026-09-15"}}, wantStatus: http.StatusOK},
		{name: "POST /html/invoices/atualizar external blocked", method: http.MethodPost, url: "/html/invoices/atualizar?id=5", body: url.Values{"edit_invoicing": {"2026-09-01"}}, wantStatus: http.StatusOK},
		{name: "GET /html/invoices/cancelar-edicao", method: http.MethodGet, url: "/html/invoices/cancelar-edicao?id=1", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/deletar-aviso", method: http.MethodGet, url: "/html/invoices/deletar-aviso?id=1", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/deletar-aviso external forbidden", method: http.MethodGet, url: "/html/invoices/deletar-aviso?id=5", wantStatus: http.StatusForbidden},
		{name: "POST /html/invoices/deletar", method: http.MethodPost, url: "/html/invoices/deletar?id=1", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/cancelar", method: http.MethodPost, url: "/html/invoices/cancelar?id=1", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/deletar external blocked", method: http.MethodPost, url: "/html/invoices/deletar?id=5", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/cancelar external blocked", method: http.MethodPost, url: "/html/invoices/cancelar?id=5", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/pagamento-aviso", method: http.MethodGet, url: "/html/invoices/pagamento-aviso?id=2", wantStatus: http.StatusOK},
		{name: "POST /html/invoices/pagar", method: http.MethodPost, url: "/html/invoices/pagar?id=2", body: url.Values{"payment_date": {"2026-09-28"}}, wantStatus: http.StatusOK},
		{name: "POST /api/invoice/payment", method: http.MethodPost, url: "/api/invoice/payment", jsonBody: `{"vendor":"estudio_amelia","id":3,"payment_date":"2026-09-28"}`, wantStatus: http.StatusOK},
		{name: "POST /invoice/api/payment", method: http.MethodPost, url: "/invoice/api/payment", jsonBody: `{"vendor":"estudio_amelia","id":4,"payment_date":"2026-09-28"}`, wantStatus: http.StatusOK},
		{name: "POST /api/invoice/cancel", method: http.MethodPost, url: "/api/invoice/cancel", jsonBody: `{"vendor":"estudio_amelia","id":21}`, wantStatus: http.StatusOK},
		{name: "POST /invoice/api/cancel by external_code", method: http.MethodPost, url: "/invoice/api/cancel", jsonBody: `{"vendor":"estudio_amelia","external_code":"EXT-99"}`, wantStatus: http.StatusOK},
		{name: "POST /invoice/cancel", method: http.MethodPost, url: "/invoice/cancel", jsonBody: `{"vendor":"estudio_amelia","id":22}`, wantStatus: http.StatusOK},
		{name: "POST /html/invoices/salvar", method: http.MethodPost, url: "/html/invoices/salvar", body: url.Values{
			"add_customer":      {"cliente_teste"},
			"add_invoicing":     {"2026-09-28"},
			"add_due":           {"2026-10-15"},
			"item_descricao[]":  {"Aula de Canto"},
			"item_quantidade[]": {"1"},
			"item_preco[]":      {"150.00"},
		}, wantStatus: http.StatusOK},
		{name: "GET /html/invoices/download doc=0 (invoice)", method: http.MethodGet, url: "/html/invoices/download?id=10&doc=0", wantStatus: http.StatusOK},
		{name: "GET /html/invoices/download doc=1 (receipt)", method: http.MethodGet, url: "/html/invoices/download?id=2&doc=1", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.jsonBody != "" {
				req = httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.jsonBody))
				req.Header.Set("Content-Type", "application/json")
			} else if tt.body != nil {
				req = httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req = httptest.NewRequest(tt.method, tt.url, nil)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			routes.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("endpoint %s returned status %d, want %d (body: %s)", tt.url, rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestInvoicesRealTemplate(t *testing.T) {
	paths := []string{
		"../../../../web/templates/invoices.html",
		"web/templates/invoices.html",
		"billing/web/templates/invoices.html",
	}
	var templateData []byte
	var err error
	wd, _ := os.Getwd()
	for _, p := range paths {
		if data, errRead := os.ReadFile(p); errRead == nil && len(data) > 0 {
			templateData = data
			break
		}
	}
	if len(templateData) == 0 {
		t.Fatalf("could not find real invoices.html template (wd: %s)", wd)
	}

	routes, err := NewRoutes(newDummyRepo(), &dummyLogger{}, &dummyTaxer{}, &dummyPixer{}, &dummyIssuer{}, templateData)
	if err != nil {
		t.Fatalf("failed to parse real template into routes: %v", err)
	}

	// Test GET /html/invoices
	req := httptest.NewRequest(http.MethodGet, "/html/invoices", nil)
	rec := httptest.NewRecorder()
	routes.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Envio Nota") {
		t.Errorf("expected body to contain 'Envio Nota'")
	}
	if !strings.Contains(body, "Envio Recibo") {
		t.Errorf("expected body to contain 'Envio Recibo'")
	}
	if !strings.Contains(body, `name="tax"`) {
		t.Errorf("expected filter form to contain input name=\"tax\"")
	}

	// Test POST /html/invoices/tabela with filter
	formData := url.Values{
		"page":          {"1"},
		"tax":           {"2026-10"},
		"email_receipt": {"2026-10"},
	}
	reqTable := httptest.NewRequest(http.MethodPost, "/html/invoices/tabela", strings.NewReader(formData.Encode()))
	reqTable.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recTable := httptest.NewRecorder()
	routes.ServeHTTP(recTable, reqTable)
	if recTable.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recTable.Code, recTable.Body.String())
	}
	bodyTable := recTable.Body.String()
	if !strings.Contains(bodyTable, "Envio Nota") {
		t.Errorf("expected table body to contain 'Envio Nota'")
	}
	if !strings.Contains(bodyTable, "cliente_teste") {
		t.Errorf("expected table body to contain customer nickname 'cliente_teste'")
	}
	if !strings.Contains(bodyTable, "Cliente Teste") {
		t.Errorf("expected table body to contain customer name 'Cliente Teste'")
	}

	// Test GET /html/invoices/editar?id=1
	reqEdit := httptest.NewRequest(http.MethodGet, "/html/invoices/editar?id=1", nil)
	recEdit := httptest.NewRecorder()
	routes.ServeHTTP(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recEdit.Code, recEdit.Body.String())
	}
	bodyEdit := recEdit.Body.String()
	if !strings.Contains(bodyEdit, `name="edit_tax"`) {
		t.Errorf("expected edit row to contain 'name=\"edit_tax\"'")
	}
	if !strings.Contains(bodyEdit, "Cliente Teste") {
		t.Errorf("expected edit row to contain customer name 'Cliente Teste'")
	}
}
