package dto

import (
	"strings"
	"testing"
	"time"

	"planner/internal/domain"
)

type mockPlanRepo struct {
	customers    map[string]*domain.Customer
	services     map[int64]*domain.Service
	activePlans  []domain.Plan
	savedPlans   []*domain.Plan
	deletedPlan  int64
	existingPlan *domain.Plan
}

func newMockPlanRepo() *mockPlanRepo {
	return &mockPlanRepo{
		customers: map[string]*domain.Customer{
			"cliente_teste": {ID: 10, Nickname: "cliente_teste", Name: "Cliente Teste", VendorID: 1},
		},
		services: map[int64]*domain.Service{
			1: {ID: 1, Name: "Aula Canto 30", VendorID: 1},
			2: {ID: 2, Name: "Aula Piano 45", VendorID: 1},
		},
	}
}

func (m *mockPlanRepo) BeginTransaction() error      { return nil }
func (m *mockPlanRepo) CommitTransaction() error     { return nil }
func (m *mockPlanRepo) RollbackTransaction() error   { return nil }
func (m *mockPlanRepo) Save(model interface{}) error { return nil }
func (m *mockPlanRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockPlanRepo) FindCount(conditions map[string]interface{}) (int64, error) { return 0, nil }
func (m *mockPlanRepo) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockPlanRepo) Close() error { return nil }
func (m *mockPlanRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	var list []domain.Customer
	for _, c := range m.customers {
		if nickname != nil && c.Nickname == *nickname {
			list = append(list, *c)
		}
	}
	return list, nil
}
func (m *mockPlanRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if c, ok := m.customers[nickname]; ok {
		return c, nil
	}
	return nil, nil
}
func (m *mockPlanRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	for _, c := range m.customers {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockPlanRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockPlanRepo) GetVendor(nickname string) (*domain.Vendor, error) { return nil, nil }
func (m *mockPlanRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	var list []domain.Service
	for _, s := range m.services {
		list = append(list, *s)
	}
	return list, nil
}
func (m *mockPlanRepo) GetServiceByID(id int64) (*domain.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return nil, nil
}
func (m *mockPlanRepo) SavePlan(plan *domain.Plan) error {
	m.savedPlans = append(m.savedPlans, plan)
	if plan.ID == 0 {
		plan.ID = 100 + int64(len(m.savedPlans))
	}
	return nil
}
func (m *mockPlanRepo) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return m.activePlans, int64(len(m.activePlans)), nil
}
func (m *mockPlanRepo) GetPlanByID(id int64) (*domain.Plan, error) {
	if m.existingPlan != nil && m.existingPlan.ID == id {
		return m.existingPlan, nil
	}
	return nil, nil
}
func (m *mockPlanRepo) DeletePlan(id int64) error {
	m.deletedPlan = id
	return nil
}
func (m *mockPlanRepo) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	var result []domain.Plan
	for _, p := range m.activePlans {
		if p.CustomerID == customerID && (excludePlanID == 0 || p.ID != excludePlanID) {
			for _, item := range p.Items {
				for _, sid := range serviceIDs {
					if item.ServiceID == sid {
						result = append(result, p)
						break
					}
				}
			}
		}
	}
	return result, nil
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestPlanCreateRequest_ValidAgenda(t *testing.T) {
	repo := newMockPlanRepo()
	req := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanType:  1,
		Price:     floatPtr(120.0),
		Items:     []PlanItemRequest{{ServiceID: 1, OrderIndex: 1}},
		Agenda: &PlanAgendaRequest{
			Recurrence: 1,
			WeekDay:    2,
			TermType:   1,
			PaymentDay: 10,
		},
	}
	err := req.Validate(repo)
	if err != nil {
		t.Fatalf("expected valid agenda plan, got: %v", err)
	}
	if req.CustomerID != 10 {
		t.Errorf("expected CustomerID 10, got %d", req.CustomerID)
	}
}

func TestPlanCreateRequest_NotebookExclusiveValidation(t *testing.T) {
	repo := newMockPlanRepo()

	// 1. Both nil -> should error
	req1 := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanType:  3,
		Price:     floatPtr(100.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Notebook:  &PlanNotebookRequest{},
	}
	err1 := req1.Validate(repo)
	if err1 == nil || !strings.Contains(err1.Error(), "uma e apenas uma") {
		t.Errorf("expected error for notebook with both nil, got: %v", err1)
	}

	// 2. Both non-nil -> should error
	d := 15
	term := 30
	req2 := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanType:  3,
		Price:     floatPtr(100.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Notebook: &PlanNotebookRequest{
			PaymentDay:  &d,
			PaymentTerm: &term,
		},
	}
	err2 := req2.Validate(repo)
	if err2 == nil || !strings.Contains(err2.Error(), "não ambas") {
		t.Errorf("expected error for notebook with both non-nil, got: %v", err2)
	}

	// 3. Only PaymentDay -> should succeed
	req3 := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanType:  3,
		Price:     floatPtr(100.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Notebook: &PlanNotebookRequest{
			PaymentDay: &d,
		},
	}
	if err := req3.Validate(repo); err != nil {
		t.Errorf("expected valid notebook with payment_day only, got: %v", err)
	}

	// 4. Only PaymentTerm -> should succeed
	req4 := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanType:  3,
		Price:     floatPtr(100.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Notebook: &PlanNotebookRequest{
			PaymentTerm: &term,
		},
	}
	if err := req4.Validate(repo); err != nil {
		t.Errorf("expected valid notebook with payment_term only, got: %v", err)
	}
}

func TestPlanCreateRequest_ActivePeriodOverlapConflict(t *testing.T) {
	repo := newMockPlanRepo()

	// Existing plan: Customer 10, Service 1, from 2026-01-01 to nil (infinite)
	pStart, _ := time.Parse("2006-01-02", "2026-01-01")
	repo.activePlans = []domain.Plan{
		{
			ID:         1,
			CustomerID: 10,
			PlanStart:  pStart,
			PlanEnd:    nil,
			PlanType:   1,
			Price:      floatPtr(100.0),
			Items: []domain.PlanItem{
				{ServiceID: 1, Service: &domain.Service{ID: 1, Name: "Aula Canto 30"}},
			},
		},
	}

	// New plan for same customer with Service 1 starting 2026-05-01 -> conflict!
	reqConflict := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-05-01",
		PlanType:  2,
		Price:     floatPtr(200.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Package: &PlanPackageRequest{
			Quantity:    5,
			PaymentDate: "2026-05-01",
		},
	}
	err := reqConflict.Validate(repo)
	if err == nil || !strings.Contains(err.Error(), "conflito de período") {
		t.Fatalf("expected period conflict error, got: %v", err)
	}

	// New plan for same customer with DIFFERENT service (Service 2) -> NO conflict!
	reqDifferentService := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-05-01",
		PlanType:  2,
		Price:     floatPtr(200.0),
		Items:     []PlanItemRequest{{ServiceID: 2}},
		Package: &PlanPackageRequest{
			Quantity:    5,
			PaymentDate: "2026-05-01",
		},
	}
	if err := reqDifferentService.Validate(repo); err != nil {
		t.Errorf("expected different service to succeed without conflict, got: %v", err)
	}
}

func TestPlanCreateRequest_InvalidDates(t *testing.T) {
	repo := newMockPlanRepo()
	end := "2025-12-31"
	req := &PlanCreateRequest{
		Nickname:  "cliente_teste",
		PlanStart: "2026-01-01",
		PlanEnd:   &end,
		PlanType:  1,
		Price:     floatPtr(100.0),
		Items:     []PlanItemRequest{{ServiceID: 1}},
		Agenda: &PlanAgendaRequest{
			Recurrence: 1,
			WeekDay:    2,
			TermType:   1,
			PaymentDay: 10,
		},
	}
	err := req.Validate(repo)
	if err == nil || !strings.Contains(err.Error(), "anterior à data de início") {
		t.Errorf("expected error when end date is before start date, got: %v", err)
	}
}

func TestPlanCreateRequest_PricingValidation(t *testing.T) {
	repo := newMockPlanRepo()

	baseReq := func() *PlanCreateRequest {
		return &PlanCreateRequest{
			Nickname:  "cliente_teste",
			PlanStart: "2026-01-01",
			PlanType:  1,
			Agenda: &PlanAgendaRequest{
				Recurrence: 1,
				WeekDay:    2,
				TermType:   1,
				PaymentDay: 10,
			},
		}
	}

	// 1. Both plan price and item price set -> Error (mutually exclusive)
	r1 := baseReq()
	r1.Price = floatPtr(200.0)
	r1.Items = []PlanItemRequest{
		{ServiceID: 1, Price: floatPtr(50.0)},
	}
	err1 := r1.Validate(repo)
	if err1 == nil || !strings.Contains(err1.Error(), "mutuamente exclusivos") {
		t.Errorf("expected error for both plan and item price set, got: %v", err1)
	}

	// 2. Neither plan price nor item price set -> Error (at least one must have price)
	r2 := baseReq()
	r2.Price = nil
	r2.Items = []PlanItemRequest{
		{ServiceID: 1, Price: nil},
	}
	err2 := r2.Validate(repo)
	if err2 == nil || !strings.Contains(err2.Error(), "valor definido") {
		t.Errorf("expected error for neither plan nor item price set, got: %v", err2)
	}

	// 3. Plan price nil, but only 1 of 2 items has price -> Error (all items must have price)
	r3 := baseReq()
	r3.Price = nil
	r3.Items = []PlanItemRequest{
		{ServiceID: 1, Price: floatPtr(50.0)},
		{ServiceID: 2, Price: nil},
	}
	err3 := r3.Validate(repo)
	if err3 == nil || !strings.Contains(err3.Error(), "todos os itens do plano devem ter valor") {
		t.Errorf("expected error for partial item pricing, got: %v", err3)
	}

	// 4. Valid plan price with nil item prices -> Success
	r4 := baseReq()
	r4.Price = floatPtr(300.0)
	r4.Items = []PlanItemRequest{
		{ServiceID: 1, Price: nil},
		{ServiceID: 2, Price: nil},
	}
	if err := r4.Validate(repo); err != nil {
		t.Errorf("expected valid plan price, got error: %v", err)
	}

	// 5. Valid item prices with nil plan price -> Success
	r5 := baseReq()
	r5.Price = nil
	r5.Items = []PlanItemRequest{
		{ServiceID: 1, Price: floatPtr(75.0)},
		{ServiceID: 2, Price: floatPtr(85.0)},
	}
	if err := r5.Validate(repo); err != nil {
		t.Errorf("expected valid item prices, got error: %v", err)
	}
}

func TestPlanUpdateRequest_PricingValidation(t *testing.T) {
	repo := newMockPlanRepo()

	// Existing plan in repo has Plan.Price = 200, item price = nil
	pStart, _ := time.Parse("2006-01-02", "2026-01-01")
	existingPlan := domain.Plan{
		ID:         99,
		CustomerID: 10,
		PlanStart:  pStart,
		PlanType:   1,
		Price:      floatPtr(200.0),
		Items: []domain.PlanItem{
			{ID: 1, PlanID: 99, ServiceID: 1, OrderIndex: 1, Price: nil},
		},
	}
	repo.existingPlan = &existingPlan

	// 1. Update Price to new valid price (items remain nil from existing) -> Success
	u1 := &PlanUpdateRequest{
		ID:    99,
		Price: floatPtr(250.0),
	}
	if err := u1.Validate(repo); err != nil {
		t.Errorf("expected valid plan price update, got: %v", err)
	}

	// 2. Update Items with prices without clearing Plan price -> Error (mutually exclusive)
	u2 := &PlanUpdateRequest{
		ID: 99,
		Items: []PlanItemRequest{
			{ServiceID: 1, Price: floatPtr(60.0)},
		},
	}
	err2 := u2.Validate(repo)
	if err2 == nil || !strings.Contains(err2.Error(), "mutuamente exclusivos") {
		t.Errorf("expected error updating items with price while plan has price, got: %v", err2)
	}

	// 3. Clear plan price and provide items with prices -> Success
	u3 := &PlanUpdateRequest{
		ID:         99,
		ClearPrice: true,
		Items: []PlanItemRequest{
			{ServiceID: 1, Price: floatPtr(60.0)},
		},
	}
	if err := u3.Validate(repo); err != nil {
		t.Errorf("expected valid switch to item prices, got: %v", err)
	}

	// 4. Clear plan price and provide items with only partial price -> Error
	u4 := &PlanUpdateRequest{
		ID:         99,
		ClearPrice: true,
		Items: []PlanItemRequest{
			{ServiceID: 1, Price: floatPtr(60.0)},
			{ServiceID: 2, Price: nil},
		},
	}
	err4 := u4.Validate(repo)
	if err4 == nil || !strings.Contains(err4.Error(), "todos os itens do plano devem ter valor") {
		t.Errorf("expected error for partial item prices on update, got: %v", err4)
	}
}

func TestPlanCreateRequest_DuplicateOrderValidation(t *testing.T) {
	repo := newMockPlanRepo()

	baseReq := func() *PlanCreateRequest {
		return &PlanCreateRequest{
			Nickname:  "cliente_teste",
			PlanStart: "2026-01-01",
			PlanType:  1,
			Price:     floatPtr(100.0),
			Agenda: &PlanAgendaRequest{
				Recurrence: 1,
				WeekDay:    2,
				TermType:   1,
				PaymentDay: 10,
			},
		}
	}

	// 1. Two items with same order (OrderIndex = 1) -> Error
	r1 := baseReq()
	r1.Items = []PlanItemRequest{
		{ServiceID: 1, OrderIndex: 1},
		{ServiceID: 2, OrderIndex: 1},
	}
	err1 := r1.Validate(repo)
	if err1 == nil || !strings.Contains(err1.Error(), "mesmo valor em order") {
		t.Fatalf("expected error for duplicate order in plan create, got: %v", err1)
	}

	// 2. Two items with different orders (OrderIndex = 1 and 2) -> Success
	r2 := baseReq()
	r2.Items = []PlanItemRequest{
		{ServiceID: 1, OrderIndex: 1},
		{ServiceID: 2, OrderIndex: 2},
	}
	if err := r2.Validate(repo); err != nil {
		t.Fatalf("expected valid distinct orders, got: %v", err)
	}

	// 3. Single item -> Success
	r3 := baseReq()
	r3.Items = []PlanItemRequest{
		{ServiceID: 1, OrderIndex: 1},
	}
	if err := r3.Validate(repo); err != nil {
		t.Fatalf("expected valid single item, got: %v", err)
	}
}

func TestPlanUpdateRequest_DuplicateOrderValidation(t *testing.T) {
	repo := newMockPlanRepo()

	pStart, _ := time.Parse("2006-01-02", "2026-01-01")
	existingPlan := domain.Plan{
		ID:         99,
		CustomerID: 10,
		PlanStart:  pStart,
		PlanType:   1,
		Price:      floatPtr(200.0),
		Items: []domain.PlanItem{
			{ID: 1, PlanID: 99, ServiceID: 1, OrderIndex: 1, Price: nil},
		},
	}
	repo.existingPlan = &existingPlan

	// 1. Update with 2 items having same order -> Error
	u1 := &PlanUpdateRequest{
		ID: 99,
		Items: []PlanItemRequest{
			{ServiceID: 1, OrderIndex: 2},
			{ServiceID: 2, OrderIndex: 2},
		},
	}
	err1 := u1.Validate(repo)
	if err1 == nil || !strings.Contains(err1.Error(), "mesmo valor em order") {
		t.Fatalf("expected error for duplicate order in plan update, got: %v", err1)
	}

	// 2. Update with 2 items having different orders -> Success
	u2 := &PlanUpdateRequest{
		ID: 99,
		Items: []PlanItemRequest{
			{ServiceID: 1, OrderIndex: 1},
			{ServiceID: 2, OrderIndex: 2},
		},
	}
	if err := u2.Validate(repo); err != nil {
		t.Fatalf("expected valid distinct orders in update, got: %v", err)
	}
}


