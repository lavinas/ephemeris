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
	return &domain.Plan{
		ID:         id,
		CustomerID: 100,
		PlanStart:  now,
		PlanType:   1,
		Items:      []domain.PlanItem{{ServiceID: 10, OrderIndex: 1}},
	}, nil
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
}


