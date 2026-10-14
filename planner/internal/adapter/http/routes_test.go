package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"planner/internal/domain"
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
func (d *dummyRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) { return nil, nil }
func (d *dummyRepo) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) { return nil, nil }
func (d *dummyRepo) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) { return nil, nil }


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

type mockInvoiceRouteRepo struct {
	dummyRepo
	deletedInvoiceID int64
}

func (m *mockInvoiceRouteRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, Nickname: "cliente_teste"}, nil
}

func (m *mockInvoiceRouteRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return &domain.Invoice{
		ID:            id,
		PlanID:        1,
		Plan:          &domain.Plan{CustomerID: 10},
		InvoicingDate: time.Now(),
		DueDate:       time.Now().AddDate(0, 0, 5),
		Amount:        150.0,
	}, nil
}

func (m *mockInvoiceRouteRepo) DeleteInvoice(id int64) error {
	m.deletedInvoiceID = id
	return nil
}

func (m *mockInvoiceRouteRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	return []domain.Invoice{
		{
			ID:            1,
			PlanID:        1,
			Plan:          &domain.Plan{CustomerID: 10},
			InvoicingDate: time.Now(),
			DueDate:       time.Now().AddDate(0, 0, 5),
			Amount:        150.0,
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

func (m *mockInvoiceRouteRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	return []domain.Plan{
		{
			ID:         1,
			CustomerID: 10,
			PlanStart:  time.Now(),
			PlanType:   1,
		},
	}, nil
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



