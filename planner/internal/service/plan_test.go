package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockLogger struct{}

func (m *mockLogger) IPrintf(level int, format string, v ...interface{}) {}
func (m *mockLogger) Close()                                             {}

type mockPlanRepoForService struct {
	customers    map[string]*domain.Customer
	services     map[int64]*domain.Service
	plans        map[int64]*domain.Plan
	nextPlanID   int64
	activePlans  []domain.Plan
}

func newMockPlanRepoForService() *mockPlanRepoForService {
	return &mockPlanRepoForService{
		customers: map[string]*domain.Customer{
			"cliente_1": {ID: 1, Nickname: "cliente_1", Name: "Cliente Um", VendorID: 1},
		},
		services: map[int64]*domain.Service{
			1: {ID: 1, Name: "Aula Canto 30", VendorID: 1},
		},
		plans:      make(map[int64]*domain.Plan),
		nextPlanID: 1,
	}
}

func (m *mockPlanRepoForService) BeginTransaction() error      { return nil }
func (m *mockPlanRepoForService) CommitTransaction() error     { return nil }
func (m *mockPlanRepoForService) RollbackTransaction() error   { return nil }
func (m *mockPlanRepoForService) Save(model interface{}) error { return nil }
func (m *mockPlanRepoForService) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockPlanRepoForService) FindCount(conditions map[string]interface{}) (int64, error) { return 0, nil }
func (m *mockPlanRepoForService) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockPlanRepoForService) Close() error { return nil }
func (m *mockPlanRepoForService) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	var list []domain.Customer
	for _, c := range m.customers {
		if nickname != nil && c.Nickname == *nickname {
			list = append(list, *c)
		}
	}
	return list, nil
}
func (m *mockPlanRepoForService) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if c, ok := m.customers[nickname]; ok {
		return c, nil
	}
	return nil, nil
}
func (m *mockPlanRepoForService) GetCustomerByID(id int64) (*domain.Customer, error) {
	for _, c := range m.customers {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockPlanRepoForService) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockPlanRepoForService) GetVendor(nickname string) (*domain.Vendor, error) { return nil, nil }
func (m *mockPlanRepoForService) FindServices(vendorID int64) ([]domain.Service, error) {
	var list []domain.Service
	for _, s := range m.services {
		list = append(list, *s)
	}
	return list, nil
}
func (m *mockPlanRepoForService) GetServiceByID(id int64) (*domain.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, nil
}
func (m *mockPlanRepoForService) SavePlan(plan *domain.Plan) error {
	if plan.ID == 0 {
		plan.ID = m.nextPlanID
		m.nextPlanID++
	}
	m.plans[plan.ID] = plan
	return nil
}
func (m *mockPlanRepoForService) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	var list []domain.Plan
	for _, p := range m.plans {
		if p.DeletedAt == nil {
			list = append(list, *p)
		}
	}
	return list, int64(len(list)), nil
}
func (m *mockPlanRepoForService) GetPlanByID(id int64) (*domain.Plan, error) {
	if p, ok := m.plans[id]; ok && p.DeletedAt == nil {
		return p, nil
	}
	return nil, nil
}
func (m *mockPlanRepoForService) DeletePlan(id int64) error {
	if p, ok := m.plans[id]; ok {
		now := time.Now()
		p.DeletedAt = &now
	}
	return nil
}
func (m *mockPlanRepoForService) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	var list []domain.Plan
	for _, p := range m.plans {
		if p.CustomerID == customerID && p.DeletedAt == nil && (excludePlanID == 0 || p.ID != excludePlanID) {
			list = append(list, *p)
		}
	}
	return list, nil
}

func TestPlanCRUD_Services(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	// 1. Create Plan
	createSvc := NewPlanCreate(repo, logger)
	initPrice := 100.0
	createReq := &dto.PlanCreateRequest{
		Nickname:  "cliente_1",
		PlanStart: "2026-01-01",
		PlanType:  1,
		Price:     &initPrice,
		Items:     []dto.PlanItemRequest{{ServiceID: 1, OrderIndex: 1}},
		Agenda: &dto.PlanAgendaRequest{
			Recurrence: 1,
			WeekDay:    2,
			TermType:   1,
			PaymentDay: 5,
		},
	}
	createResp := createSvc.Run(createReq)
	if createResp.GetStatusCode() != 200 {
		t.Fatalf("PlanCreate failed: %s", createResp.GetMessage())
	}
	createdPlanID := createResp.(*dto.PlanCreateResponse).PlanID
	if createdPlanID != 1 {
		t.Errorf("expected plan ID 1, got %d", createdPlanID)
	}

	// 2. List Plans
	listSvc := NewPlanList(repo, logger)
	listReq := &dto.PlanListRequest{Page: 1, PageSize: 10}
	listResp := listSvc.Run(listReq)
	if listResp.GetStatusCode() != 200 {
		t.Fatalf("PlanList failed: %s", listResp.GetMessage())
	}
	plans := listResp.(*dto.PlanListResponse).Plans
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan in list, got %d", len(plans))
	}
	if plans[0].Nickname != "cliente_1" || plans[0].PlanTypeName != "Agenda" {
		t.Errorf("unexpected plan details: %+v", plans[0])
	}

	// 3. Update Plan (change plan price)
	updateSvc := NewPlanUpdate(repo, logger)
	newPrice := 250.0
	updateReq := &dto.PlanUpdateRequest{
		ID:    createdPlanID,
		Price: &newPrice,
	}
	updateResp := updateSvc.Run(updateReq)
	if updateResp.GetStatusCode() != 200 {
		t.Fatalf("PlanUpdate failed: %s", updateResp.GetMessage())
	}
	updatedPlan, _ := repo.GetPlanByID(createdPlanID)
	if updatedPlan.Price == nil || *updatedPlan.Price != 250.0 {
		t.Errorf("expected price 250.0, got %v", updatedPlan.Price)
	}

	// 3b. Update Plan (switch from plan price to item price)
	itemPrice := 65.0
	updateReq2 := &dto.PlanUpdateRequest{
		ID:         createdPlanID,
		ClearPrice: true,
		Items: []dto.PlanItemRequest{
			{ServiceID: 1, OrderIndex: 1, Price: &itemPrice},
		},
	}
	updateResp2 := updateSvc.Run(updateReq2)
	if updateResp2.GetStatusCode() != 200 {
		t.Fatalf("PlanUpdate switch to item price failed: %s", updateResp2.GetMessage())
	}
	updatedPlan2, _ := repo.GetPlanByID(createdPlanID)
	if updatedPlan2.Price != nil {
		t.Errorf("expected plan price to be nil, got %v", updatedPlan2.Price)
	}
	if len(updatedPlan2.Items) != 1 || updatedPlan2.Items[0].Price == nil || *updatedPlan2.Items[0].Price != 65.0 {
		t.Errorf("expected item price 65.0, got %+v", updatedPlan2.Items)
	}

	// 4. Delete Plan
	delSvc := NewPlanDelete(repo, logger)
	delReq := &dto.PlanDeleteRequest{PlanID: createdPlanID}
	delResp := delSvc.Run(delReq)
	if delResp.GetStatusCode() != 200 {
		t.Fatalf("PlanDelete failed: %s", delResp.GetMessage())
	}

	// Check that plan is no longer returned
	deletedPlan, _ := repo.GetPlanByID(createdPlanID)
	if deletedPlan != nil {
		t.Errorf("expected plan to be deleted, but still found")
	}
}
