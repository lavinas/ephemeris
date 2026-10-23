package http

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

type dummyLogger struct{}

func (d *dummyLogger) IPrintf(level int, format string, v ...interface{}) {}
func (d *dummyLogger) Close()                                             {}

type dummyRepo struct{}

func (d *dummyRepo) BeginTransaction() error      { return nil }
func (d *dummyRepo) CommitTransaction() error     { return nil }
func (d *dummyRepo) RollbackTransaction() error   { return nil }
func (d *dummyRepo) Save(model interface{}) error { return nil }
func (d *dummyRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (d *dummyRepo) FindCount(conditions map[string]interface{}) (int64, error) { return 0, nil }
func (d *dummyRepo) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (d *dummyRepo) Close() error { return nil }
func (d *dummyRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (d *dummyRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (d *dummyRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return nil, nil
}
func (d *dummyRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (d *dummyRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	return nil, nil
}
func (d *dummyRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return nil, nil
}
func (d *dummyRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	return nil, nil
}
func (d *dummyRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return nil, nil
}
func (d *dummyRepo) SavePlan(plan *domain.Plan) error { return nil }
func (d *dummyRepo) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return nil, 0, nil
}
func (d *dummyRepo) GetPlanByID(id int64) (*domain.Plan, error) { return nil, nil }
func (d *dummyRepo) DeletePlan(id int64) error                    { return nil }
func (d *dummyRepo) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}
func (d *dummyRepo) SaveInvoice(invoice *domain.Invoice) error { return nil }
func (d *dummyRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) { return nil, nil }
func (d *dummyRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) { return nil, nil }
func (d *dummyRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) { return nil, 0, 0, nil }
func (d *dummyRepo) DeleteInvoice(id int64) error { return nil }
func (d *dummyRepo) UpdateInvoice(invoice *domain.Invoice) error { return nil }
func (d *dummyRepo) DeleteInvoicesByIDs(ids []int64) error { return nil }
func (d *dummyRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) { return nil, nil }
func (d *dummyRepo) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) { return nil, nil }
func (d *dummyRepo) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) { return nil, nil }
func (d *dummyRepo) SaveInvoices(invoices []*domain.Invoice) error { return nil }
func (d *dummyRepo) CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error) { return false, nil }


func TestRootEndpointServesStaticIndex(t *testing.T) {
	// Change working directory to planner/backend for test to resolve web/static
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	templateData := []byte(`{{define "index"}}<html><body>Sessions</body></html>{{end}}`)
	routes, err := NewRoutes(&dummyRepo{}, &dummyLogger{}, templateData)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. Test GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	routes.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 on /, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "subframe-conteudo") {
		t.Errorf("expected index.html to contain 'subframe-conteudo'")
	}
	if !strings.Contains(body, `/html/sessoes`) {
		t.Errorf("expected index.html to reference '/html/sessoes'")
	}

	// 2. Test unknown route returns 404
	req404 := httptest.NewRequest(http.MethodGet, "/nao-existe", nil)
	rec404 := httptest.NewRecorder()
	routes.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Errorf("expected status 404 on /nao-existe, got %d", rec404.Code)
	}

	// 3. Test GET /html/ping
	reqPing := httptest.NewRequest(http.MethodGet, "/html/ping", nil)
	recPing := httptest.NewRecorder()
	routes.ServeHTTP(recPing, reqPing)

	if recPing.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/ping, got %d", recPing.Code)
	}

	// 4. Test GET /static/logo.png
	reqLogo := httptest.NewRequest(http.MethodGet, "/static/logo.png", nil)
	recLogo := httptest.NewRecorder()
	routes.ServeHTTP(recLogo, reqLogo)

	if recLogo.Code != http.StatusOK {
		t.Errorf("expected status 200 on /static/logo.png, got %d", recLogo.Code)
	}

	// 5. Test GET /html/sessoes (the subframe target)
	reqSessoes := httptest.NewRequest(http.MethodGet, "/html/sessoes", nil)
	recSessoes := httptest.NewRecorder()
	routes.ServeHTTP(recSessoes, reqSessoes)

	if recSessoes.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/sessoes, got %d", recSessoes.Code)
	}
	if !strings.Contains(recSessoes.Body.String(), "Sessions") {
		t.Errorf("expected /html/sessoes response to contain 'Sessions'")
	}

	// 6. Test index.html references /html/planos
	if !strings.Contains(body, `/html/planos`) {
		t.Errorf("expected index.html to reference '/html/planos'")
	}

	// 7. Test GET /html/planos
	reqPlanos := httptest.NewRequest(http.MethodGet, "/html/planos", nil)
	recPlanos := httptest.NewRecorder()
	routes.ServeHTTP(recPlanos, reqPlanos)
	if recPlanos.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/planos, got %d", recPlanos.Code)
	}

	// 8. Test API plan endpoints exist
	reqPlanList := httptest.NewRequest(http.MethodGet, "/api/plan/list", nil)
	recPlanList := httptest.NewRecorder()
	routes.ServeHTTP(recPlanList, reqPlanList)
	if recPlanList.Code != http.StatusOK {
		t.Errorf("expected status 200 on /api/plan/list, got %d", recPlanList.Code)
	}

	reqPlanCreate := httptest.NewRequest(http.MethodPost, "/api/plan/create", strings.NewReader(`{}`))
	recPlanCreate := httptest.NewRecorder()
	routes.ServeHTTP(recPlanCreate, reqPlanCreate)
	if recPlanCreate.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 on empty /api/plan/create, got %d", recPlanCreate.Code)
	}

	reqPlanUpdate := httptest.NewRequest(http.MethodPatch, "/api/plan/update", strings.NewReader(`{}`))
	recPlanUpdate := httptest.NewRecorder()
	routes.ServeHTTP(recPlanUpdate, reqPlanUpdate)
	if recPlanUpdate.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 on empty /api/plan/update, got %d", recPlanUpdate.Code)
	}

	reqPlanDelete := httptest.NewRequest(http.MethodDelete, "/api/plan/delete", strings.NewReader(`{}`))
	recPlanDelete := httptest.NewRecorder()
	routes.ServeHTTP(recPlanDelete, reqPlanDelete)
	if recPlanDelete.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 on empty /api/plan/delete, got %d", recPlanDelete.Code)
	}

	// 9. Test API invoicing endpoints exist
	reqInvoicing := httptest.NewRequest(http.MethodPost, "/api/invoicing", strings.NewReader(`{}`))
	recInvoicing := httptest.NewRecorder()
	routes.ServeHTTP(recInvoicing, reqInvoicing)
	if recInvoicing.Code != http.StatusOK {
		t.Errorf("expected status 200 on /api/invoicing, got %d: %s", recInvoicing.Code, recInvoicing.Body.String())
	}

	reqPlanInvoicing := httptest.NewRequest(http.MethodPost, "/api/plan/invoicing", strings.NewReader(`{}`))
	recPlanInvoicing := httptest.NewRecorder()
	routes.ServeHTTP(recPlanInvoicing, reqPlanInvoicing)
	if recPlanInvoicing.Code != http.StatusOK {
		t.Errorf("expected status 200 on /api/plan/invoicing, got %d: %s", recPlanInvoicing.Code, recPlanInvoicing.Body.String())
	}

	// 10. Test API invoicing list endpoints exist
	reqInvList := httptest.NewRequest(http.MethodGet, "/api/plan/invoicing/list", nil)
	recInvList := httptest.NewRecorder()
	routes.ServeHTTP(recInvList, reqInvList)
	if recInvList.Code != http.StatusOK {
		t.Errorf("expected status 200 on /api/plan/invoicing/list, got %d: %s", recInvList.Code, recInvList.Body.String())
	}

	reqInvInvoices := httptest.NewRequest(http.MethodGet, "/api/plan/invoicing/invoices", nil)
	recInvInvoices := httptest.NewRecorder()
	routes.ServeHTTP(recInvInvoices, reqInvInvoices)
	if recInvInvoices.Code != http.StatusOK {
		t.Errorf("expected status 200 on /api/plan/invoicing/invoices, got %d: %s", recInvInvoices.Code, recInvInvoices.Body.String())
	}

	// 11. Test index.html references /html/invoicing
	if !strings.Contains(body, `/html/invoicing`) {
		t.Errorf("expected index.html to reference '/html/invoicing'")
	}

	// 12. Test GET /html/invoicing
	reqInvoicingHTML := httptest.NewRequest(http.MethodGet, "/html/invoicing", nil)
	recInvoicingHTML := httptest.NewRecorder()
	routes.ServeHTTP(recInvoicingHTML, reqInvoicingHTML)
	if recInvoicingHTML.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/invoicing, got %d", recInvoicingHTML.Code)
	}

	// 13. Test POST /html/invoicing/tabela
	reqInvoicingTabela := httptest.NewRequest(http.MethodPost, "/html/invoicing/tabela", nil)
	recInvoicingTabela := httptest.NewRecorder()
	routes.ServeHTTP(recInvoicingTabela, reqInvoicingTabela)
	if recInvoicingTabela.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/invoicing/tabela, got %d", recInvoicingTabela.Code)
	}
}

type mockPlanRepoWithItems struct {
	dummyRepo
	lastSavedPlan *domain.Plan
}

func (m *mockPlanRepoWithItems) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return &domain.Customer{ID: 100, Nickname: nickname}, nil
}

func (m *mockPlanRepoWithItems) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id, Name: "Serviço"}, nil
}

func (m *mockPlanRepoWithItems) FindServices(vendorID int64) ([]domain.Service, error) {
	return []domain.Service{{ID: 10, Name: "S10"}, {ID: 20, Name: "S20"}}, nil
}

func (m *mockPlanRepoWithItems) GetPlanByID(id int64) (*domain.Plan, error) {
	now := time.Now()
	price := 100.0
	return &domain.Plan{
		ID:         id,
		CustomerID: 100,
		PlanStart:  now,
		PlanType:   1,
		Price:      &price,
		Items:      []domain.PlanItem{{ServiceID: 10, OrderIndex: 1}},
	}, nil
}

func (m *mockPlanRepoWithItems) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	now := time.Now()
	price := 100.0
	return []domain.Plan{
		{
			ID:         1,
			CustomerID: 100,
			PlanStart:  now,
			PlanType:   1,
			Price:      &price,
			Items:      []domain.PlanItem{{ServiceID: 10, OrderIndex: 1}},
		},
	}, 1, nil
}

func (m *mockPlanRepoWithItems) SavePlan(plan *domain.Plan) error {
	m.lastSavedPlan = plan
	return nil
}

func (m *mockPlanRepoWithItems) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}

func TestPlansSaveAndMultipleItems(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockPlanRepoWithItems{}
	templateData := []byte(`{{define "index"}}<html><body>Sessions</body></html>{{end}}`)
	routes, err := NewRoutes(mockR, &dummyLogger{}, templateData)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. Test POST /html/planos/salvar with 2 items
	form := strings.NewReader("add_nickname=cliente_1&add_plan_start=2026-10-01&add_plan_type=1&add_agenda_recorrencia=1&add_agenda_dia_semana=2&add_agenda_vigencia=1&add_agenda_dia_pagamento=5&add_price=150.00&add_servico_id=10&add_order_index=1&add_servico_id=20&add_order_index=2")
	req := httptest.NewRequest(http.MethodPost, "/html/planos/salvar", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	routes.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/salvar, got %d: %s", rec.Code, rec.Body.String())
	}
	if mockR.lastSavedPlan == nil {
		t.Fatal("expected plan to be saved, got nil")
	}
	if len(mockR.lastSavedPlan.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(mockR.lastSavedPlan.Items))
	}
	if mockR.lastSavedPlan.Items[0].ServiceID != 10 || mockR.lastSavedPlan.Items[0].OrderIndex != 1 {
		t.Errorf("expected item 0 to be service 10 order 1, got %+v", mockR.lastSavedPlan.Items[0])
	}
	if mockR.lastSavedPlan.Items[1].ServiceID != 20 || mockR.lastSavedPlan.Items[1].OrderIndex != 2 {
		t.Errorf("expected item 1 to be service 20 order 2, got %+v", mockR.lastSavedPlan.Items[1])
	}

	// 2. Test POST /html/planos/atualizar?id=1 with 3 items
	editForm := strings.NewReader("edit_plan_start=2026-10-01&edit_price=200.00&edit_servico_id=10&edit_order_index=1&edit_servico_id=20&edit_order_index=2&edit_servico_id=30&edit_order_index=3")
	reqEdit := httptest.NewRequest(http.MethodPost, "/html/planos/atualizar?id=1", editForm)
	reqEdit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recEdit := httptest.NewRecorder()
	routes.ServeHTTP(recEdit, reqEdit)

	if recEdit.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/atualizar, got %d: %s", recEdit.Code, recEdit.Body.String())
	}
	if len(mockR.lastSavedPlan.Items) != 3 {
		t.Fatalf("expected 3 items on update, got %d", len(mockR.lastSavedPlan.Items))
	}
	if mockR.lastSavedPlan.Items[2].ServiceID != 30 || mockR.lastSavedPlan.Items[2].OrderIndex != 3 {
		t.Errorf("expected item 2 to be service 30 order 3, got %+v", mockR.lastSavedPlan.Items[2])
	}

	// 3. Test POST /html/planos/tabela with ativo=true
	tabForm := strings.NewReader("ativo=true&page=1")
	reqTab := httptest.NewRequest(http.MethodPost, "/html/planos/tabela", tabForm)
	reqTab.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recTab := httptest.NewRecorder()
	routes.ServeHTTP(recTab, reqTab)

	if recTab.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/tabela with ativo=true, got %d", recTab.Code)
	}

	// 4. Test POST /html/planos/tabela without ativo (show all)
	tabFormAll := strings.NewReader("page=1")
	reqTabAll := httptest.NewRequest(http.MethodPost, "/html/planos/tabela", tabFormAll)
	reqTabAll.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recTabAll := httptest.NewRecorder()
	routes.ServeHTTP(recTabAll, reqTabAll)

	if recTabAll.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/tabela without ativo, got %d", recTabAll.Code)
	}

	// 5. Test POST /html/planos/tabela/reset
	reqReset := httptest.NewRequest(http.MethodPost, "/html/planos/tabela/reset", nil)
	recReset := httptest.NewRecorder()
	routes.ServeHTTP(recReset, reqReset)

	if recReset.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/tabela/reset, got %d", recReset.Code)
	}

	// 6. Test rejection when saving plan with duplicate order
	dupForm := strings.NewReader("add_nickname=cliente_1&add_plan_start=2026-10-01&add_plan_type=1&add_agenda_recorrencia=1&add_agenda_dia_semana=2&add_agenda_vigencia=1&add_agenda_dia_pagamento=5&add_price=150.00&add_servico_id=10&add_order_index=1&add_servico_id=20&add_order_index=1")
	reqDup := httptest.NewRequest(http.MethodPost, "/html/planos/salvar", dupForm)
	reqDup.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recDup := httptest.NewRecorder()
	routes.ServeHTTP(recDup, reqDup)
	if !strings.Contains(recDup.Body.String(), "mesmo valor em order") {
		t.Errorf("expected duplicate order error message in response body, got: %s", recDup.Body.String())
	}

	// 7. Test rejection when updating plan with duplicate order
	dupEditForm := strings.NewReader("edit_plan_start=2026-10-01&edit_price=200.00&edit_servico_id=10&edit_order_index=2&edit_servico_id=20&edit_order_index=2")
	reqDupEdit := httptest.NewRequest(http.MethodPost, "/html/planos/atualizar?id=1", dupEditForm)
	reqDupEdit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recDupEdit := httptest.NewRecorder()
	routes.ServeHTTP(recDupEdit, reqDupEdit)
	if recDupEdit.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on /html/planos/atualizar with duplicate order, got %d", recDupEdit.Code)
	}
	if !strings.Contains(recDupEdit.Body.String(), "mesmo valor em order") {
		t.Errorf("expected duplicate order error message, got: %s", recDupEdit.Body.String())
	}

	// 8. Test GET /html/planos/encerrar-aviso?id=1
	reqEncerrarAviso := httptest.NewRequest(http.MethodGet, "/html/planos/encerrar-aviso?id=1", nil)
	recEncerrarAviso := httptest.NewRecorder()
	routes.ServeHTTP(recEncerrarAviso, reqEncerrarAviso)
	if recEncerrarAviso.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/encerrar-aviso, got %d: %s", recEncerrarAviso.Code, recEncerrarAviso.Body.String())
	}
	if !strings.Contains(recEncerrarAviso.Body.String(), "encerrar_plan_end") {
		t.Errorf("expected /html/planos/encerrar-aviso to render input encerrar_plan_end, got: %s", recEncerrarAviso.Body.String())
	}

	// 9. Test POST /html/planos/encerrar?id=1 with valid date
	encerrarForm := strings.NewReader("encerrar_plan_end=2026-10-25&page=1")
	reqEncerrar := httptest.NewRequest(http.MethodPost, "/html/planos/encerrar?id=1", encerrarForm)
	reqEncerrar.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recEncerrar := httptest.NewRecorder()
	routes.ServeHTTP(recEncerrar, reqEncerrar)
	if recEncerrar.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/encerrar, got %d: %s", recEncerrar.Code, recEncerrar.Body.String())
	}
	if mockR.lastSavedPlan == nil || mockR.lastSavedPlan.PlanEnd == nil {
		t.Fatalf("expected lastSavedPlan.PlanEnd to be set, got nil")
	}
	if mockR.lastSavedPlan.PlanEnd.Format("2006-01-02") != "2026-10-25" {
		t.Errorf("expected PlanEnd 2026-10-25, got %s", mockR.lastSavedPlan.PlanEnd.Format("2006-01-02"))
	}

	// 10. Test POST /html/planos/encerrar?id=1 with date before plan_start (error 400)
	invalidDateForm := strings.NewReader("encerrar_plan_end=2020-01-01")
	reqInvalid := httptest.NewRequest(http.MethodPost, "/html/planos/encerrar?id=1", invalidDateForm)
	reqInvalid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recInvalid := httptest.NewRecorder()
	routes.ServeHTTP(recInvalid, reqInvalid)
	if recInvalid.Code != http.StatusBadRequest {
		t.Errorf("expected 400 on /html/planos/encerrar with date before start, got %d", recInvalid.Code)
	}
}

type mockBillingClientForRoutes struct {
	cancelledInvoices []port.BillingCancelInput
	createdInvoices   []port.BillingInvoiceInput
}

func (m *mockBillingClientForRoutes) CreateInvoice(ctx context.Context, input port.BillingInvoiceInput) (int64, error) {
	m.createdInvoices = append(m.createdInvoices, input)
	return 9999, nil
}

func (m *mockBillingClientForRoutes) GetInvoiceByExternalCode(ctx context.Context, vendor, externalCode string) (*int64, error) {
	return nil, nil
}

func (m *mockBillingClientForRoutes) CancelInvoice(ctx context.Context, input port.BillingCancelInput) error {
	m.cancelledInvoices = append(m.cancelledInvoices, input)
	return nil
}

type mockInvoiceRouteRepo struct {
	dummyRepo
	deletedInvoiceID  int64
	updatedInvoice    *domain.Invoice
	deletedBatchIDs   []int64
	invoicedInvoiceID int64
}

func (m *mockInvoiceRouteRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vendedor_teste"}, nil
}

func (m *mockInvoiceRouteRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, Nickname: "cliente_teste", VendorID: 1}, nil
}

func (m *mockInvoiceRouteRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if nickname == "inexistente" {
		return nil, nil
	}
	return &domain.Customer{ID: 10, Nickname: nickname, VendorID: vendorID}, nil
}

func (m *mockInvoiceRouteRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	return []domain.Service{
		{ID: 1, Name: "Aula Canto 30", VendorID: vendorID},
		{ID: 10, Name: "Servico Teste", VendorID: vendorID},
	}, nil
}

func (m *mockInvoiceRouteRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	var refID *int64
	if m.invoicedInvoiceID != 0 && id == m.invoicedInvoiceID {
		rid := int64(8888)
		refID = &rid
	}
	return &domain.Invoice{
		ID:                 id,
		PlanID:             1,
		Plan:               &domain.Plan{CustomerID: 10},
		InvoicingDate:      time.Now(),
		DueDate:            time.Now().AddDate(0, 0, 5),
		Amount:             150.0,
		ReferenceInvoiceID: refID,
		Items: []domain.InvoiceItem{
			{ID: 1, InvoiceID: id, ServiceID: 10, Price: 50.0, InvoicedQuantity: 3},
		},
	}, nil
}

func (m *mockInvoiceRouteRepo) DeleteInvoice(id int64) error {
	m.deletedInvoiceID = id
	return nil
}

func (m *mockInvoiceRouteRepo) UpdateInvoice(invoice *domain.Invoice) error {
	m.updatedInvoice = invoice
	return nil
}

func (m *mockInvoiceRouteRepo) DeleteInvoicesByIDs(ids []int64) error {
	m.deletedBatchIDs = append(m.deletedBatchIDs, ids...)
	return nil
}

func (m *mockInvoiceRouteRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	return []domain.Invoice{
		{
			ID:            1,
			PlanID:        1,
			Plan:          &domain.Plan{CustomerID: 10},
			InvoicingDate: time.Now(),
			DueDate:       time.Now().AddDate(0, 0, 5),
			Amount:        150.0,
			Items: []domain.InvoiceItem{
				{ID: 1, InvoiceID: 1, ServiceID: 10, Price: 50.0, InvoicedQuantity: 3},
			},
		},
	}, nil
}

func (m *mockInvoiceRouteRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	var refID *int64
	if m.invoicedInvoiceID != 0 {
		rid := int64(8888)
		refID = &rid
	}
	return []domain.Invoice{
		{
			ID:                 1,
			PlanID:             1,
			Plan:               &domain.Plan{CustomerID: 10},
			InvoicingDate:      time.Now(),
			DueDate:            time.Now().AddDate(0, 0, 5),
			Amount:             150.0,
			ReferenceInvoiceID: refID,
		},
	}, 1, 150.0, nil
}

func (m *mockInvoiceRouteRepo) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return []domain.Plan{
		{
			ID:         1,
			CustomerID: 10,
			PlanStart:  time.Now(),
			PlanType:   1,
		},
	}, 1, nil
}

func (m *mockInvoiceRouteRepo) GetPlanByID(id int64) (*domain.Plan, error) {
	price := 100.0
	return &domain.Plan{
		ID:         id,
		CustomerID: 10,
		PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PlanType:   1,
		Price:      &price,
		Agenda: &domain.PlanAgenda{
			ID:         id,
			Recurrence: 1,
			WeekDay:    1,
			TermType:   1,
			PaymentDay: 5,
		},
		Items: []domain.PlanItem{
			{ID: 1, PlanID: id, ServiceID: 1, Price: &price, OrderIndex: 1},
		},
	}, nil
}

func (m *mockInvoiceRouteRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id, Name: "Servico Teste"}, nil
}

func (m *mockInvoiceRouteRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	price := 100.0
	return []domain.Plan{
		{
			ID:         1,
			CustomerID: 10,
			PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PlanType:   1,
			Price:      &price,
			Agenda: &domain.PlanAgenda{
				ID:         1,
				Recurrence: 1,
				WeekDay:    1,
				TermType:   1,
				PaymentDay: 5,
			},
			Items: []domain.PlanItem{
				{ID: 1, PlanID: 1, ServiceID: 1, Price: &price, OrderIndex: 1},
			},
		},
	}, nil
}

func (m *mockInvoiceRouteRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	if nickname != nil && *nickname == "inexistente" {
		return nil, nil
	}
	return []domain.Customer{{ID: 10, Nickname: "cliente_teste", VendorID: 1}}, nil
}

func TestPlanGenerateInvoiceModalAndAction(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockInvoiceRouteRepo{}
	routes, err := NewRoutes(mockR, &dummyLogger{}, nil)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. GET /html/planos/gerar-invoice-modal?id=1
	reqModal := httptest.NewRequest(http.MethodGet, "/html/planos/gerar-invoice-modal?id=1", nil)
	recModal := httptest.NewRecorder()
	routes.ServeHTTP(recModal, reqModal)
	if recModal.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-invoice-modal, got %d", recModal.Code)
	}
	if !strings.Contains(recModal.Body.String(), "Gerar Pré-Fatura") {
		t.Errorf("expected modal to contain 'Gerar Pré-Fatura', got: %s", recModal.Body.String())
	}

	// 2. POST /html/planos/gerar-invoice?id=1
	form := strings.NewReader("start_date=2026-10-01&end_date=2026-10-31")
	reqGen := httptest.NewRequest(http.MethodPost, "/html/planos/gerar-invoice?id=1", form)
	reqGen.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recGen := httptest.NewRecorder()
	routes.ServeHTTP(recGen, reqGen)
	if recGen.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-invoice, got %d", recGen.Code)
	}
	if !strings.Contains(recGen.Body.String(), "Processamento Concluído") {
		t.Errorf("expected modal to contain 'Processamento Concluído', got: %s", recGen.Body.String())
	}
	if !strings.Contains(recGen.Body.String(), "navegarParaPreFaturas") {
		t.Errorf("expected modal to contain 'navegarParaPreFaturas', got: %s", recGen.Body.String())
	}

	// 3. POST /html/planos/gerar-invoices (batch by active filters)
	batchForm := strings.NewReader("tipo_filtro=1&servico_filtro=1&data_inicio=2026-10-01&data_fim=2026-10-31&ativo=true")
	reqBatch := httptest.NewRequest(http.MethodPost, "/html/planos/gerar-invoices", batchForm)
	reqBatch.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recBatch := httptest.NewRecorder()
	routes.ServeHTTP(recBatch, reqBatch)
	if recBatch.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-invoices, got %d", recBatch.Code)
	}
	if !strings.Contains(recBatch.Body.String(), "Processamento Concluído") {
		t.Errorf("expected batch response to contain 'Processamento Concluído', got: %s", recBatch.Body.String())
	}

	// 4. POST /html/planos/gerar-pre-faturas (alias) with nickname filter
	aliasForm := strings.NewReader("nickname=cliente_teste&tipo_filtro=0&data_inicio=2026-10-01&data_fim=2026-10-31")
	reqAlias := httptest.NewRequest(http.MethodPost, "/html/planos/gerar-pre-faturas", aliasForm)
	reqAlias.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recAlias := httptest.NewRecorder()
	routes.ServeHTTP(recAlias, reqAlias)
	if recAlias.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-pre-faturas, got %d", recAlias.Code)
	}
	if !strings.Contains(recAlias.Body.String(), "Processamento Concluído") {
		t.Errorf("expected alias response to contain 'Processamento Concluído', got: %s", recAlias.Body.String())
	}

	// 5. GET /html/planos/gerar-invoices-modal
	reqBatchModal := httptest.NewRequest(http.MethodGet, "/html/planos/gerar-invoices-modal?nickname=cliente_teste&tipo_filtro=1", nil)
	recBatchModal := httptest.NewRecorder()
	routes.ServeHTTP(recBatchModal, reqBatchModal)
	if recBatchModal.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-invoices-modal, got %d", recBatchModal.Code)
	}
	if !strings.Contains(recBatchModal.Body.String(), "Gerar Pré-Faturas") {
		t.Errorf("expected batch modal to contain 'Gerar Pré-Faturas', got: %s", recBatchModal.Body.String())
	}
	if !strings.Contains(recBatchModal.Body.String(), "name=\"start_date\"") || !strings.Contains(recBatchModal.Body.String(), "name=\"end_date\"") {
		t.Errorf("expected batch modal to contain start_date and end_date inputs, got: %s", recBatchModal.Body.String())
	}

	// 6. GET /html/planos/gerar-pre-faturas-modal (alias)
	reqAliasModal := httptest.NewRequest(http.MethodGet, "/html/planos/gerar-pre-faturas-modal", nil)
	recAliasModal := httptest.NewRecorder()
	routes.ServeHTTP(recAliasModal, reqAliasModal)
	if recAliasModal.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos/gerar-pre-faturas-modal, got %d", recAliasModal.Code)
	}

	// 7. GET /html/planos should contain 'Gerar' button with hx-get to modal
	reqPlans := httptest.NewRequest(http.MethodGet, "/html/planos", nil)
	recPlans := httptest.NewRecorder()
	routes.ServeHTTP(recPlans, reqPlans)
	if recPlans.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/planos, got %d", recPlans.Code)
	}
	if !strings.Contains(recPlans.Body.String(), "Gerar") {
		t.Errorf("expected /html/planos to contain 'Gerar' button, got: %s", recPlans.Body.String())
	}
	if !strings.Contains(recPlans.Body.String(), "hx-get=\"/html/planos/gerar-invoices-modal\"") {
		t.Errorf("expected 'Gerar' button to have hx-get to modal, got: %s", recPlans.Body.String())
	}
}

func TestInvoiceDeleteRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockInvoiceRouteRepo{}
	routes, err := NewRoutes(mockR, &dummyLogger{}, nil)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. DELETE /api/invoice/delete?id=1
	reqDelAPI := httptest.NewRequest(http.MethodDelete, "/api/invoice/delete?id=1", nil)
	recDelAPI := httptest.NewRecorder()
	routes.ServeHTTP(recDelAPI, reqDelAPI)
	if recDelAPI.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/invoice/delete, got %d: %s", recDelAPI.Code, recDelAPI.Body.String())
	}
	if mockR.deletedInvoiceID != 1 {
		t.Errorf("expected deletedInvoiceID 1, got %d", mockR.deletedInvoiceID)
	}

	// 2. POST /api/invoicing/delete (JSON body)
	reqDelJSON := httptest.NewRequest(http.MethodPost, "/api/invoicing/delete", strings.NewReader(`{"id": 2}`))
	reqDelJSON.Header.Set("Content-Type", "application/json")
	recDelJSON := httptest.NewRecorder()
	routes.ServeHTTP(recDelJSON, reqDelJSON)
	if recDelJSON.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/invoicing/delete, got %d: %s", recDelJSON.Code, recDelJSON.Body.String())
	}
	if mockR.deletedInvoiceID != 2 {
		t.Errorf("expected deletedInvoiceID 2, got %d", mockR.deletedInvoiceID)
	}

	// 3. GET /html/invoicing/deletar-aviso?id=1
	reqAviso := httptest.NewRequest(http.MethodGet, "/html/invoicing/deletar-aviso?id=1", nil)
	recAviso := httptest.NewRecorder()
	routes.ServeHTTP(recAviso, reqAviso)
	if recAviso.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar-aviso, got %d", recAviso.Code)
	}
	if !strings.Contains(recAviso.Body.String(), "Sim, Excluir") {
		t.Errorf("expected warning to contain 'Sim, Excluir', got: %s", recAviso.Body.String())
	}

	// 4. GET /html/invoicing/cancelar-delete?id=1
	reqCancel := httptest.NewRequest(http.MethodGet, "/html/invoicing/cancelar-delete?id=1", nil)
	recCancel := httptest.NewRecorder()
	routes.ServeHTTP(recCancel, reqCancel)
	if recCancel.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/cancelar-delete, got %d", recCancel.Code)
	}

	// 5. POST /html/invoicing/deletar?id=1
	reqDelHTML := httptest.NewRequest(http.MethodPost, "/html/invoicing/deletar?id=1", nil)
	recDelHTML := httptest.NewRecorder()
	routes.ServeHTTP(recDelHTML, reqDelHTML)
	if recDelHTML.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar, got %d", recDelHTML.Code)
	}
}

func TestInvoiceEditAndBatchDeleteRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockInvoiceRouteRepo{}
	routes, err := NewRoutes(mockR, &dummyLogger{}, nil)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. GET /html/invoicing/editar?id=1
	reqEdit := httptest.NewRequest(http.MethodGet, "/html/invoicing/editar?id=1", nil)
	recEdit := httptest.NewRecorder()
	routes.ServeHTTP(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/editar, got %d: %s", recEdit.Code, recEdit.Body.String())
	}
	if !strings.Contains(recEdit.Body.String(), "Salvar") && !strings.Contains(recEdit.Body.String(), "due_date") {
		t.Errorf("expected edit row to contain 'Salvar' or 'due_date', got: %s", recEdit.Body.String())
	}

	// 2. GET /html/invoicing/cancelar-edicao?id=1
	reqCancelEdit := httptest.NewRequest(http.MethodGet, "/html/invoicing/cancelar-edicao?id=1", nil)
	recCancelEdit := httptest.NewRecorder()
	routes.ServeHTTP(recCancelEdit, reqCancelEdit)
	if recCancelEdit.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/cancelar-edicao, got %d", recCancelEdit.Code)
	}

	// 3. POST /html/invoicing/atualizar?id=1
	updateForm := strings.NewReader("due_date=2026-11-20&item_qty_1=4")
	reqUpdateHTML := httptest.NewRequest(http.MethodPost, "/html/invoicing/atualizar?id=1", updateForm)
	reqUpdateHTML.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recUpdateHTML := httptest.NewRecorder()
	routes.ServeHTTP(recUpdateHTML, reqUpdateHTML)
	if recUpdateHTML.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/atualizar, got %d: %s", recUpdateHTML.Code, recUpdateHTML.Body.String())
	}
	if mockR.updatedInvoice == nil {
		t.Errorf("expected repo to have updatedInvoice")
	}

	// 4. PUT /api/invoice/update (JSON API)
	jsonUpdateBody := `{"id": 1, "due_date": "2026-11-25", "items": [{"id": 1, "invoiced_quantity": 5}]}`
	reqUpdateAPI := httptest.NewRequest(http.MethodPut, "/api/invoice/update", strings.NewReader(jsonUpdateBody))
	reqUpdateAPI.Header.Set("Content-Type", "application/json")
	recUpdateAPI := httptest.NewRecorder()
	routes.ServeHTTP(recUpdateAPI, reqUpdateAPI)
	if recUpdateAPI.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/invoice/update, got %d: %s", recUpdateAPI.Code, recUpdateAPI.Body.String())
	}

	// 5. POST /html/invoicing/deletar-lote-aviso
	filterForm := strings.NewReader("nickname=cliente_teste&plan_type=1")
	reqBatchWarn := httptest.NewRequest(http.MethodPost, "/html/invoicing/deletar-lote-aviso", filterForm)
	reqBatchWarn.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recBatchWarn := httptest.NewRecorder()
	routes.ServeHTTP(recBatchWarn, reqBatchWarn)
	if recBatchWarn.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar-lote-aviso, got %d: %s", recBatchWarn.Code, recBatchWarn.Body.String())
	}
	if !strings.Contains(recBatchWarn.Body.String(), "Excluir Pré-Faturas em Lote") && !strings.Contains(recBatchWarn.Body.String(), "modal-deletar-lote") {
		t.Errorf("expected modal to be rendered, got: %s", recBatchWarn.Body.String())
	}
	if !strings.Contains(recBatchWarn.Body.String(), `hx-post="/html/invoicing/deletar-lote"`) {
		t.Errorf("expected modal to contain hx-post form for deletar-lote")
	}

	// 6. POST /html/invoicing/deletar-lote
	reqBatchDel := httptest.NewRequest(http.MethodPost, "/html/invoicing/deletar-lote", strings.NewReader("nickname=cliente_teste"))
	reqBatchDel.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recBatchDel := httptest.NewRecorder()
	routes.ServeHTTP(recBatchDel, reqBatchDel)
	if recBatchDel.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar-lote, got %d: %s", recBatchDel.Code, recBatchDel.Body.String())
	}
	if len(mockR.deletedBatchIDs) == 0 {
		t.Errorf("expected mockR.deletedBatchIDs to be populated")
	}
	if !strings.Contains(recBatchDel.Body.String(), `id="modal-container-invoicing" hx-swap-oob="true"`) {
		t.Errorf("expected response to include oob clear for modal-container-invoicing, got: %s", recBatchDel.Body.String())
	}

	// 7. DELETE /api/invoice/batch-delete (JSON API)
	reqBatchDelAPI := httptest.NewRequest(http.MethodDelete, "/api/invoice/batch-delete?nickname=cliente_teste", nil)
	recBatchDelAPI := httptest.NewRecorder()
	routes.ServeHTTP(recBatchDelAPI, reqBatchDelAPI)
	if recBatchDelAPI.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/invoice/batch-delete, got %d: %s", recBatchDelAPI.Code, recBatchDelAPI.Body.String())
	}
}

func TestInvoicing_AlreadyInvoicedBlockedAndPendingFilter(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockInvoiceRouteRepo{invoicedInvoiceID: 1}
	billingClient := &mockBillingClientForRoutes{}
	routes, err := NewRoutesWithBilling(mockR, &dummyLogger{}, billingClient, nil)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. GET /html/invoicing/editar?id=1 must return 200 for invoiced invoice
	reqEdit := httptest.NewRequest(http.MethodGet, "/html/invoicing/editar?id=1", nil)
	recEdit := httptest.NewRecorder()
	routes.ServeHTTP(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/editar for invoiced pre-invoice, got %d", recEdit.Code)
	}

	// 2. GET /html/invoicing/deletar-aviso?id=1 must return 200 for invoiced invoice
	reqDelWarn := httptest.NewRequest(http.MethodGet, "/html/invoicing/deletar-aviso?id=1", nil)
	recDelWarn := httptest.NewRecorder()
	routes.ServeHTTP(recDelWarn, reqDelWarn)
	if recDelWarn.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar-aviso for invoiced pre-invoice, got %d", recDelWarn.Code)
	}

	// 3. POST /html/invoicing/deletar?id=1 must succeed with 200 and cancel in billing
	reqDel := httptest.NewRequest(http.MethodPost, "/html/invoicing/deletar?id=1", nil)
	recDel := httptest.NewRecorder()
	routes.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/deletar for invoiced pre-invoice, got %d", recDel.Code)
	}
	if len(billingClient.cancelledInvoices) != 1 {
		t.Errorf("expected 1 cancelled invoice in billing, got %d", len(billingClient.cancelledInvoices))
	}

	// 4. POST /html/invoicing/atualizar?id=1 must succeed with 200, cancel old invoice and create replacement
	updateForm := strings.NewReader("due_date=2026-11-20")
	reqUpd := httptest.NewRequest(http.MethodPost, "/html/invoicing/atualizar?id=1", updateForm)
	reqUpd.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recUpd := httptest.NewRecorder()
	routes.ServeHTTP(recUpd, reqUpd)
	if recUpd.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/atualizar for invoiced pre-invoice, got %d", recUpd.Code)
	}
	if len(billingClient.cancelledInvoices) != 2 {
		t.Errorf("expected 2 cancelled invoices in billing, got %d", len(billingClient.cancelledInvoices))
	}
	if len(billingClient.createdInvoices) != 1 {
		t.Errorf("expected 1 created replacement invoice in billing, got %d", len(billingClient.createdInvoices))
	}

	// 5. InvoicingTable when invoiced: should show "✓ Faturada #8888" and should show edit/delete buttons, but NOT faturar button
	reqTable := httptest.NewRequest(http.MethodPost, "/html/invoicing/tabela", nil)
	recTable := httptest.NewRecorder()
	routes.ServeHTTP(recTable, reqTable)
	if recTable.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/tabela, got %d", recTable.Code)
	}
	bodyTable := recTable.Body.String()
	if !strings.Contains(bodyTable, "Faturada #8888") {
		t.Errorf("expected invoiced row to show 'Faturada #8888', got: %s", bodyTable)
	}
	if !strings.Contains(bodyTable, `/html/invoicing/editar?id=1`) {
		t.Errorf("expected invoiced row to have edit button")
	}
	if !strings.Contains(bodyTable, `/html/invoicing/deletar-aviso?id=1`) {
		t.Errorf("expected invoiced row to have delete button")
	}
	if strings.Contains(bodyTable, `/html/invoicing/faturar?id=1`) {
		t.Errorf("expected invoiced row to NOT have faturar button")
	}

	// 6. Test when NOT invoiced: should show "⏳ Pendente", edit button, delete button, and pendente toggle switch
	mockR2 := &mockInvoiceRouteRepo{invoicedInvoiceID: 0}
	routes2, err := NewRoutes(mockR2, &dummyLogger{}, nil)
	if err != nil {
		t.Fatalf("failed to create routes2: %v", err)
	}
	reqHTML := httptest.NewRequest(http.MethodGet, "/html/invoicing", nil)
	recHTML := httptest.NewRecorder()
	routes2.ServeHTTP(recHTML, reqHTML)
	if recHTML.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing, got %d", recHTML.Code)
	}
	bodyHTML := recHTML.Body.String()
	if !strings.Contains(bodyHTML, "filtro-pendente") {
		t.Errorf("expected /html/invoicing to contain 'filtro-pendente'")
	}
	if !strings.Contains(bodyHTML, "Pendente") {
		t.Errorf("expected /html/invoicing to contain 'Pendente'")
	}
	if !strings.Contains(bodyHTML, `/html/invoicing/editar?id=1`) {
		t.Errorf("expected pending row to have edit button")
	}
	if !strings.Contains(bodyHTML, `/html/invoicing/deletar-aviso?id=1`) {
		t.Errorf("expected pending row to have delete button")
	}
}

type mockReconcileRouteRepo struct {
	dummyRepo
	reconciledSessions map[int64]bool
	sessions           []domain.Session
	invoices           []domain.Invoice
}

func (m *mockReconcileRouteRepo) GetSessionByID(id int64) (*domain.Session, error) {
	for _, s := range m.sessions {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockReconcileRouteRepo) FindMatchingInvoiceItem(customerID, serviceID int64, sessionDate time.Time) (*domain.InvoiceItem, *domain.Invoice, error) {
	return nil, nil, nil
}

func (m *mockReconcileRouteRepo) IsSessionReconciled(sessionID int64) (bool, error) {
	return m.reconciledSessions[sessionID], nil
}

func (m *mockReconcileRouteRepo) FindUnreconciledSessions(customerID, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	return nil, nil
}

func (m *mockReconcileRouteRepo) AssociateSessionToInvoiceItem(invoiceItemID, sessionID int64) error {
	return nil
}

func (m *mockReconcileRouteRepo) DissociateSessionFromInvoiceItem(sessionID int64) (*domain.InvoiceItem, error) {
	return nil, nil
}

func (m *mockReconcileRouteRepo) FindActivePayPerViewPlan(customerID, serviceID int64, sessionDate time.Time) (*domain.Plan, *domain.PlanItem, error) {
	return nil, nil, nil
}

func (m *mockReconcileRouteRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	for _, inv := range m.invoices {
		if inv.ID == id {
			return &inv, nil
		}
	}
	return nil, nil
}

func (m *mockReconcileRouteRepo) SaveInvoice(invoice *domain.Invoice) error {
	return nil
}

func (m *mockReconcileRouteRepo) FindCount(conditions map[string]interface{}) (int64, error) {
	return int64(len(m.sessions)), nil
}

func (m *mockReconcileRouteRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	res := make([]interface{}, len(m.sessions))
	for i, s := range m.sessions {
		res[i] = s
	}
	return res, nil
}

func (m *mockReconcileRouteRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, Nickname: "cliente_teste"}, nil
}

func (m *mockReconcileRouteRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id, Name: "Serviço Teste"}, nil
}

func (m *mockReconcileRouteRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	total := 0.0
	for _, inv := range m.invoices {
		total += inv.Amount
	}
	return m.invoices, int64(len(m.invoices)), total, nil
}

type mockReconciliationTracker struct {
	sessionReconciledID int64
	invoiceReconciledID int64
}

func (m *mockReconciliationTracker) ReconcileSession(sessionID int64) error {
	m.sessionReconciledID = sessionID
	return nil
}

func (m *mockReconciliationTracker) DeconcileSession(sessionID int64) error {
	return nil
}

func (m *mockReconciliationTracker) ReconcileInvoice(invoiceID int64) error {
	m.invoiceReconciledID = invoiceID
	return nil
}

func TestHTMLReconciliationRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	sessionsTmpl, err := os.ReadFile("web/templates/sessions.html")
	if err != nil {
		t.Fatalf("failed reading sessions template: %v", err)
	}
	plansTmpl, _ := os.ReadFile("web/templates/plans.html")
	invTmpl, _ := os.ReadFile("web/templates/invoicing.html")

	mockRepo := &mockReconcileRouteRepo{
		reconciledSessions: make(map[int64]bool),
	}
	mux, _, err := NewHTMLRoutes(mockRepo, &dummyLogger{}, sessionsTmpl, plansTmpl, invTmpl)
	if err != nil {
		t.Fatalf("failed creating HTML routes: %v", err)
	}

	// 1. Test POST /sessoes/reconciliar?id=42
	reqSess := httptest.NewRequest(http.MethodPost, "/sessoes/reconciliar?id=42", nil)
	recSess := httptest.NewRecorder()
	mux.ServeHTTP(recSess, reqSess)
	if recSess.Code != http.StatusOK {
		t.Errorf("expected 200 on /sessoes/reconciliar, got %d", recSess.Code)
	}

	// 2. Test POST /invoicing/reconciliar?id=99
	reqInv := httptest.NewRequest(http.MethodPost, "/invoicing/reconciliar?id=99", nil)
	recInv := httptest.NewRecorder()
	mux.ServeHTTP(recInv, reqInv)
	if recInv.Code != http.StatusOK {
		t.Errorf("expected 200 on /invoicing/reconciliar, got %d", recInv.Code)
	}

	// 3. Test POST /pre-faturas/reconciliar?id=101
	reqPre := httptest.NewRequest(http.MethodPost, "/pre-faturas/reconciliar?id=101", nil)
	recPre := httptest.NewRecorder()
	mux.ServeHTTP(recPre, reqPre)
	if recPre.Code != http.StatusOK {
		t.Errorf("expected 200 on /pre-faturas/reconciliar, got %d", recPre.Code)
	}
}

func TestHTMLSessionReconciliationButtonStates(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	sessionsTmpl, err := os.ReadFile("web/templates/sessions.html")
	if err != nil {
		t.Fatalf("failed reading sessions template: %v", err)
	}

	now := time.Now()
	comment := "Sessão de teste"
	mockRepo := &mockReconcileRouteRepo{
		reconciledSessions: map[int64]bool{
			1: true,
			2: false,
			3: false,
		},
		sessions: []domain.Session{
			{
				ID:             1,
				CustomerID:     10,
				ServiceID:      20,
				SessionDate:    now,
				SessionMinutes: 60,
				SessionStatus:  "realizada",
				Comments:       &comment,
			},
			{
				ID:             2,
				CustomerID:     10,
				ServiceID:      20,
				SessionDate:    now,
				SessionMinutes: 60,
				SessionStatus:  "realizada",
				Comments:       &comment,
			},
			{
				ID:             3,
				CustomerID:     10,
				ServiceID:      20,
				SessionDate:    now,
				SessionMinutes: 60,
				SessionStatus:  "cancelada_nao_cobrar",
				Comments:       &comment,
			},
		},
	}

	mux, _, err := NewHTMLRoutes(mockRepo, &dummyLogger{}, sessionsTmpl)
	if err != nil {
		t.Fatalf("failed creating HTML routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/sessoes/tabela", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on /sessoes/tabela, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()

	// Verify reconciliation buttons are removed from session rows
	if strings.Contains(body, "Sessão conciliada com pré-fatura") {
		t.Errorf("expected session table NOT to have reconciled tooltip")
	}
	if strings.Contains(body, "Sessão não conciliada - Clique para conciliar") {
		t.Errorf("expected session table NOT to have unreconciled tooltip")
	}
	if strings.Contains(body, `hx-post="/html/sessoes/reconciliar`) {
		t.Errorf("expected session row NOT to contain hx-post to /html/sessoes/reconciliar")
	}
}

func TestHTMLInvoicingReconciliationButton(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	sessionsTmpl, _ := os.ReadFile("web/templates/sessions.html")
	plansTmpl, _ := os.ReadFile("web/templates/plans.html")
	invTmpl, err := os.ReadFile("web/templates/invoicing.html")
	if err != nil {
		t.Fatalf("failed reading invoicing template: %v", err)
	}

	now := time.Now()
	desc := "Sessão Terapia"
	mockRepo := &mockReconcileRouteRepo{
		invoices: []domain.Invoice{
			{
				ID:            5,
				PlanID:        1,
				Plan:          &domain.Plan{CustomerID: 10, Customer: &domain.Customer{ID: 10, Nickname: "cliente_teste"}},
				InvoicingDate: now,
				DueDate:       now.AddDate(0, 0, 5),
				PeriodStart:   now.AddDate(0, -1, 0),
				PeriodEnd:     now,
				Amount:        200.0,
				Items: []domain.InvoiceItem{
					{
						ID:               1,
						InvoiceID:        5,
						ServiceID:        10,
						Price:            100.0,
						Description:      &desc,
						InvoicedQuantity: 2,
						RealizedQuantity: 1,
					},
				},
			},
		},
	}

	mux, _, err := NewHTMLRoutes(mockRepo, &dummyLogger{}, sessionsTmpl, plansTmpl, invTmpl)
	if err != nil {
		t.Fatalf("failed creating HTML routes: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/invoicing/tabela", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on /invoicing/tabela, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()

	// Verify that the reconcile button is NOT present on the invoice row
	if strings.Contains(body, `hx-post="/html/invoicing/reconciliar`) {
		t.Errorf("expected invoicing table NOT to contain hx-post to /html/invoicing/reconciliar, got: %s", body)
	}
	if strings.Contains(body, "Conciliar pré-fatura com sessões pendentes") {
		t.Errorf("expected invoicing table NOT to contain reconciliation tooltip")
	}
	if strings.Contains(body, "Qtd Real.") {
		t.Errorf("expected invoicing table NOT to contain Qtd Real. column")
	}
}

type mockComptimeRouteRepo struct {
	dummyRepo
	records []port.ComptimeView
}

func (m *mockComptimeRouteRepo) LockCustomerService(ctx context.Context, customerID, serviceID int64) error {
	return nil
}
func (m *mockComptimeRouteRepo) FindOrCreateComptime(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetComptime(ctx context.Context, id int64) (*domain.Comptime, error) {
	for _, r := range m.records {
		if r.ID == id {
			return &domain.Comptime{
				ID:              r.ID,
				VendorID:        r.VendorID,
				CustomerID:      r.CustomerID,
				ServiceID:       r.ServiceID,
				Month:           r.Month,
				InvoicedMinutes: r.InvoicedMinutes,
				ConsumedMinutes: r.ConsumedMinutes,
				Balance:         r.Balance,
				UpdatedAt:       r.UpdatedAt,
			}, nil
		}
	}
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetPreviousComptime(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetSubsequentComptimes(ctx context.Context, customerID, serviceID int64, month time.Time) ([]domain.Comptime, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) SaveComptime(ctx context.Context, comptime *domain.Comptime) error {
	return nil
}
func (m *mockComptimeRouteRepo) ListComptime(ctx context.Context, filter port.ComptimeFilter) ([]port.ComptimeView, int64, error) {
	for i := range m.records {
		if m.records[i].MonthBalance == 0 && (m.records[i].InvoicedMinutes != 0 || m.records[i].ConsumedMinutes != 0) {
			m.records[i].MonthBalance = m.records[i].InvoicedMinutes - m.records[i].ConsumedMinutes
		}
	}
	return m.records, int64(len(m.records)), nil
}
func (m *mockComptimeRouteRepo) FindComptimeSessionBySessionID(ctx context.Context, sessionID int64) (*domain.ComptimeSession, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) SaveComptimeSession(ctx context.Context, cs *domain.ComptimeSession) error {
	return nil
}
func (m *mockComptimeRouteRepo) DeleteComptimeSession(ctx context.Context, sessionID int64) error {
	return nil
}
func (m *mockComptimeRouteRepo) FindComptimeInvoiceItemByInvoiceItemID(ctx context.Context, invoiceItemID int64) (*domain.ComptimeInvoiceItem, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) FindComptimeInvoiceItemsByInvoiceID(ctx context.Context, invoiceID int64) ([]domain.ComptimeInvoiceItem, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) SaveComptimeInvoiceItem(ctx context.Context, cii *domain.ComptimeInvoiceItem) error {
	return nil
}
func (m *mockComptimeRouteRepo) DeleteComptimeInvoiceItem(ctx context.Context, invoiceItemID int64) error {
	return nil
}
func (m *mockComptimeRouteRepo) SaveComptimeExtra(ctx context.Context, ce *domain.ComptimeExtra) error {
	return nil
}
func (m *mockComptimeRouteRepo) DeleteComptimeExtra(ctx context.Context, id int64) error {
	return nil
}
func (m *mockComptimeRouteRepo) FindActiveSessionsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.Session, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) FindActiveInvoiceItemsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.InvoiceItem, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetSessionByID(id int64) (*domain.Session, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) GetInvoiceItemByID(ctx context.Context, id int64) (*domain.InvoiceItem, error) {
	return nil, nil
}
func (m *mockComptimeRouteRepo) DeleteComptime(ctx context.Context, id int64) error {
	return nil
}

func TestHTMLComptimeRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	sessionsTmpl, err := os.ReadFile("web/templates/sessions.html")
	if err != nil {
		t.Fatalf("failed reading sessions template: %v", err)
	}
	plansTmpl, _ := os.ReadFile("web/templates/plans.html")
	invTmpl, _ := os.ReadFile("web/templates/invoicing.html")
	compTmpl, err := os.ReadFile("web/templates/comptime.html")
	if err != nil {
		t.Fatalf("failed reading comptime template: %v", err)
	}

	mockRepo := &mockComptimeRouteRepo{
		records: []port.ComptimeView{
			{
				ID:               1,
				VendorID:         1,
				CustomerID:       10,
				CustomerNickname: "cteste",
				CustomerName:     "Cliente Teste",
				ServiceID:        20,
				ServiceName:      "Psicoterapia",
				Month:            time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
				InvoicedMinutes:  240,
				ConsumedMinutes:  180,
				MonthBalance:     60,
				Balance:          60,
				UpdatedAt:        time.Now(),
			},
		},
	}

	mux, _, err := NewHTMLRoutes(mockRepo, &dummyLogger{}, sessionsTmpl, plansTmpl, invTmpl, compTmpl)
	if err != nil {
		t.Fatalf("failed creating HTML routes: %v", err)
	}

	// 1. Test GET /comptime
	reqPage := httptest.NewRequest(http.MethodGet, "/comptime", nil)
	recPage := httptest.NewRecorder()
	mux.ServeHTTP(recPage, reqPage)
	if recPage.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime, got %d", recPage.Code)
	}
	bodyPage := recPage.Body.String()
	if !strings.Contains(bodyPage, "Banco de Horas") {
		t.Errorf("expected body to contain 'Banco de Horas'")
	}
	if !strings.Contains(bodyPage, "cteste") {
		t.Errorf("expected body to contain nickname 'cteste'")
	}
	if !strings.Contains(bodyPage, "Cliente Teste") {
		t.Errorf("expected body to contain 'Cliente Teste'")
	}
	if !strings.Contains(bodyPage, "SALDO MÊS") {
		t.Errorf("expected body to contain 'SALDO MÊS'")
	}
	if !strings.Contains(bodyPage, "SALDO TOTAL") {
		t.Errorf("expected body to contain 'SALDO TOTAL'")
	}
	if !strings.Contains(bodyPage, "240 min") {
		t.Errorf("expected body to contain '240 min'")
	}

	// 2. Test POST /comptime/tabela
	reqTable := httptest.NewRequest(http.MethodPost, "/comptime/tabela", nil)
	recTable := httptest.NewRecorder()
	mux.ServeHTTP(recTable, reqTable)
	if recTable.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/tabela, got %d", recTable.Code)
	}
	bodyTable := recTable.Body.String()
	if !strings.Contains(bodyTable, "Psicoterapia") {
		t.Errorf("expected table to contain 'Psicoterapia'")
	}
	if !strings.Contains(bodyTable, "+60 min") {
		t.Errorf("expected table to contain '+60 min'")
	}

	// 3. Test GET /comptime/resync-modal
	reqModal := httptest.NewRequest(http.MethodGet, "/comptime/resync-modal", nil)
	recModal := httptest.NewRecorder()
	mux.ServeHTTP(recModal, reqModal)
	if recModal.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/resync-modal, got %d", recModal.Code)
	}
	bodyModal := recModal.Body.String()
	if !strings.Contains(bodyModal, "Ressincronizar Banco de Horas") {
		t.Errorf("expected modal to contain 'Ressincronizar Banco de Horas'")
	}
	if !strings.Contains(bodyModal, "Executar Sincronização") {
		t.Errorf("expected modal to contain 'Executar Sincronização'")
	}

	// 4. Test POST /comptime/resync
	formData := url.Values{
		"start_date":  {"2026-01-01"},
		"end_date":    {"2026-06-30"},
		"data_type":   {"todos"},
		"customer_id": {"1"},
		"service_id":  {"2"},
	}
	reqResync := httptest.NewRequest(http.MethodPost, "/comptime/resync", strings.NewReader(formData.Encode()))
	reqResync.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recResync := httptest.NewRecorder()
	mux.ServeHTTP(recResync, reqResync)
	if recResync.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/resync, got %d", recResync.Code)
	}
	bodyResync := recResync.Body.String()
	if !strings.Contains(bodyResync, "Sincronização Concluída") {
		t.Errorf("expected resync response to contain 'Sincronização Concluída'")
	}

	// 5. Test GET /comptime/deletar-aviso?id=1
	reqDelAviso := httptest.NewRequest(http.MethodGet, "/comptime/deletar-aviso?id=1", nil)
	recDelAviso := httptest.NewRecorder()
	mux.ServeHTTP(recDelAviso, reqDelAviso)
	if recDelAviso.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/deletar-aviso, got %d", recDelAviso.Code)
	}
	bodyDelAviso := recDelAviso.Body.String()
	if !strings.Contains(bodyDelAviso, "Deseja realmente apagar o banco de horas") {
		t.Errorf("expected delete warning banner in response, got: %s", bodyDelAviso)
	}

	// 6. Test GET /comptime/cancelar-delete?id=1
	reqCancelDel := httptest.NewRequest(http.MethodGet, "/comptime/cancelar-delete?id=1", nil)
	recCancelDel := httptest.NewRecorder()
	mux.ServeHTTP(recCancelDel, reqCancelDel)
	if recCancelDel.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/cancelar-delete, got %d", recCancelDel.Code)
	}
	bodyCancelDel := recCancelDel.Body.String()
	if !strings.Contains(bodyCancelDel, "linha-comptime-1") {
		t.Errorf("expected restored row in response, got: %s", bodyCancelDel)
	}

	// 7. Test POST /comptime/deletar?id=1
	reqDoDel := httptest.NewRequest(http.MethodPost, "/comptime/deletar?id=1", nil)
	recDoDel := httptest.NewRecorder()
	mux.ServeHTTP(recDoDel, reqDoDel)
	if recDoDel.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/deletar, got %d", recDoDel.Code)
	}
}

func TestAPIComptimeRoutes(t *testing.T) {
	mockRepo := &mockComptimeRouteRepo{
		records: []port.ComptimeView{
			{
				ID:       1,
				VendorID: 1,
			},
		},
	}
	mux := NewAPIRoutes(mockRepo, &dummyLogger{})

	// 1. DELETE /comptime/delete?id=1
	reqDel := httptest.NewRequest(http.MethodDelete, "/comptime/delete?id=1", nil)
	recDel := httptest.NewRecorder()
	mux.ServeHTTP(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/delete, got %d: %s", recDel.Code, recDel.Body.String())
	}

	// 2. POST /comptime/delete with JSON body
	reqDelJSON := httptest.NewRequest(http.MethodPost, "/comptime/delete", strings.NewReader(`{"id": 1}`))
	recDelJSON := httptest.NewRecorder()
	mux.ServeHTTP(recDelJSON, reqDelJSON)
	if recDelJSON.Code != http.StatusOK {
		t.Fatalf("expected 200 on /comptime/delete (JSON), got %d: %s", recDelJSON.Code, recDelJSON.Body.String())
	}

	// 3. Validation error on id <= 0
	reqDelInvalid := httptest.NewRequest(http.MethodDelete, "/comptime/delete?id=0", nil)
	recDelInvalid := httptest.NewRecorder()
	mux.ServeHTTP(recDelInvalid, reqDelInvalid)
	if recDelInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on /comptime/delete with invalid id, got %d", recDelInvalid.Code)
	}
}

func TestInvoiceCreateAndImportCSVRoutes(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	mockR := &mockInvoiceRouteRepo{}
	routes, err := NewRoutes(mockR, &dummyLogger{}, nil)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. GET /html/invoicing/novo (Modal)
	reqModal := httptest.NewRequest(http.MethodGet, "/html/invoicing/novo", nil)
	recModal := httptest.NewRecorder()
	routes.ServeHTTP(recModal, reqModal)
	if recModal.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/novo, got %d", recModal.Code)
	}
	if !strings.Contains(recModal.Body.String(), "Adicionar Pré-Fatura") {
		t.Errorf("expected modal to contain 'Adicionar Pré-Fatura', got: %s", recModal.Body.String())
	}

	// 2. GET /html/invoicing/item-linha (Dynamic item row)
	reqRow := httptest.NewRequest(http.MethodGet, "/html/invoicing/item-linha", nil)
	recRow := httptest.NewRecorder()
	routes.ServeHTTP(recRow, reqRow)
	if recRow.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/item-linha, got %d", recRow.Code)
	}
	if !strings.Contains(recRow.Body.String(), "linha-item-modal") {
		t.Errorf("expected row to contain 'linha-item-modal', got: %s", recRow.Body.String())
	}

	// 3. POST /html/invoicing/salvar (HTML manual save)
	formData := strings.NewReader("nickname=cliente_teste&invoice_date=2026-10-22&due_date=2026-10-25&period_start=2026-10-01&period_end=2026-10-31&item_service_id=1&item_price=100.00&item_quantity=2&item_description=teste")
	reqSave := httptest.NewRequest(http.MethodPost, "/html/invoicing/salvar", formData)
	reqSave.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recSave := httptest.NewRecorder()
	routes.ServeHTTP(recSave, reqSave)
	if recSave.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/salvar, got %d", recSave.Code)
	}
	if !strings.Contains(recSave.Body.String(), "Fatura criada com sucesso") {
		t.Errorf("expected table to contain success message, got: %s", recSave.Body.String())
	}

	// 4. POST /html/invoicing/importar-csv (HTML CSV import via multipart/form-data)
	csvContent := `cliente,faturamento,vencimento,serviço,inicio,fim,descrição,preço,quantidade,referencia
cliente_teste,02/01/2024,05/01/2026,Aula Canto 30,01/12/2023,31/12/2023,aula 1,"98,00",2,`

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("file", "test.csv")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte(csvContent))
	writer.Close()

	reqImportHTML := httptest.NewRequest(http.MethodPost, "/html/invoicing/importar-csv", buf)
	reqImportHTML.Header.Set("Content-Type", writer.FormDataContentType())
	recImportHTML := httptest.NewRecorder()
	routes.ServeHTTP(recImportHTML, reqImportHTML)
	if recImportHTML.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/invoicing/importar-csv, got %d", recImportHTML.Code)
	}
	if !strings.Contains(recImportHTML.Body.String(), "importados com sucesso") {
		t.Errorf("expected table to contain success message, got: %s", recImportHTML.Body.String())
	}

	// 5. POST /api/invoice/create (API JSON manual save)
	jsonBody := `{
		"nickname": "cliente_teste",
		"invoice_date": "2026-10-22",
		"due_date": "2026-10-25",
		"period_start": "2026-10-01",
		"period_end": "2026-10-31",
		"items": [
			{"service_id": 1, "price": 100.0, "quantity": 1, "description": "api item"}
		]
	}`
	reqAPICreate := httptest.NewRequest(http.MethodPost, "/api/invoice/create", strings.NewReader(jsonBody))
	reqAPICreate.Header.Set("Content-Type", "application/json")
	recAPICreate := httptest.NewRecorder()
	routes.ServeHTTP(recAPICreate, reqAPICreate)
	if recAPICreate.Code != http.StatusCreated {
		t.Errorf("expected 201 on /api/invoice/create, got %d: %s", recAPICreate.Code, recAPICreate.Body.String())
	}

	// 6. POST /api/invoicing/import-csv (API CSV import via multipart/form-data)
	bufAPI := &bytes.Buffer{}
	writerAPI := multipart.NewWriter(bufAPI)
	partAPI, err := writerAPI.CreateFormFile("file", "api_test.csv")
	if err != nil {
		t.Fatalf("failed to create api form file: %v", err)
	}
	partAPI.Write([]byte(csvContent))
	writerAPI.Close()

	reqAPIImport := httptest.NewRequest(http.MethodPost, "/api/invoicing/import-csv", bufAPI)
	reqAPIImport.Header.Set("Content-Type", writerAPI.FormDataContentType())
	recAPIImport := httptest.NewRecorder()
	routes.ServeHTTP(recAPIImport, reqAPIImport)
	if recAPIImport.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/invoicing/import-csv, got %d: %s", recAPIImport.Code, recAPIImport.Body.String())
	}
}





