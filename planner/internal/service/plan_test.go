package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
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
	invoices     []domain.Invoice
}

func newMockPlanRepoForService() *mockPlanRepoForService {
	desc1 := "Aula Canto 30"
	return &mockPlanRepoForService{
		customers: map[string]*domain.Customer{
			"cliente_1": {ID: 1, Nickname: "cliente_1", Name: "Cliente Um", VendorID: 1},
			"cliente_2": {ID: 2, Nickname: "cliente_2", Name: "Cliente Dois", VendorID: 1},
		},
		services: map[int64]*domain.Service{
			1: {ID: 1, Name: "Aula Canto 30", Description: &desc1, VendorID: 1},
		},
		plans:      make(map[int64]*domain.Plan),
		nextPlanID: 1,
		invoices:   make([]domain.Invoice, 0),
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
func (m *mockPlanRepoForService) GetVendorByID(id int64) (*domain.Vendor, error)   { return nil, nil }
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
	if plan.Package != nil {
		plan.Package.ID = plan.ID
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
func (m *mockPlanRepoForService) SaveInvoice(invoice *domain.Invoice) error {
	m.invoices = append(m.invoices, *invoice)
	return nil
}
func (m *mockPlanRepoForService) GetInvoiceByID(id int64) (*domain.Invoice, error) { return nil, nil }
func (m *mockPlanRepoForService) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *mockPlanRepoForService) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) { return nil, 0, 0, nil }
func (m *mockPlanRepoForService) DeleteInvoice(id int64) error { return nil }
func (m *mockPlanRepoForService) UpdateInvoice(invoice *domain.Invoice) error { return nil }
func (m *mockPlanRepoForService) DeleteInvoicesByIDs(ids []int64) error { return nil }
func (m *mockPlanRepoForService) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	var list []domain.Plan
	for _, p := range m.plans {
		if p.DeletedAt != nil {
			continue
		}
		if planID != nil && *planID != 0 && p.ID != *planID {
			continue
		}
		if planType != nil && *planType != 0 && p.PlanType != *planType {
			continue
		}
		planCopy := *p
		itemsCopy := make([]domain.PlanItem, len(p.Items))
		for i, it := range p.Items {
			itemsCopy[i] = it
			if it.Service == nil && m.services != nil {
				if s, ok := m.services[it.ServiceID]; ok {
					itemsCopy[i].Service = s
				}
			}
		}
		planCopy.Items = itemsCopy
		list = append(list, planCopy)
	}
	return list, nil
}
func (m *mockPlanRepoForService) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) { return nil, nil }
func (m *mockPlanRepoForService) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) { return nil, nil }
func (m *mockPlanRepoForService) SaveInvoices(invoices []*domain.Invoice) error {
	for _, inv := range invoices {
		m.invoices = append(m.invoices, *inv)
	}
	return nil
}
func (m *mockPlanRepoForService) CheckInvoiceExists(customerID int64, invoiceDate, periodStart, periodEnd time.Time, excludeInvoiceID int64) (bool, error) { return false, nil }

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

func TestPlanList_ActiveFilter(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	pastDate := time.Now().AddDate(-1, 0, 0)
	futureDate := time.Now().AddDate(1, 0, 0)
	yesterday := time.Now().AddDate(0, 0, -1)
	tomorrow := time.Now().AddDate(0, 0, 1)

	// Plan 1: Active (past start, nil end)
	p1 := domain.NewPlan(1, pastDate, nil, 1, nil)
	p1.ID = 1
	repo.plans[1] = p1

	// Plan 2: Active (past start, future end)
	p2 := domain.NewPlan(1, pastDate, &tomorrow, 1, nil)
	p2.ID = 2
	repo.plans[2] = p2

	// Plan 3: Inactive (past start, past end)
	p3 := domain.NewPlan(1, pastDate, &yesterday, 1, nil)
	p3.ID = 3
	repo.plans[3] = p3

	// Plan 4: Inactive (future start, future end)
	p4 := domain.NewPlan(1, tomorrow, &futureDate, 1, nil)
	p4.ID = 4
	repo.plans[4] = p4

	listSvc := NewPlanList(repo, logger)

	// List all (Active is nil)
	reqAll := &dto.PlanListRequest{Page: 1, PageSize: 10}
	respAll := listSvc.Run(reqAll)
	if respAll.GetStatusCode() != 200 {
		t.Fatalf("PlanList failed: %s", respAll.GetMessage())
	}
	plans := respAll.(*dto.PlanListResponse).Plans
	if len(plans) != 4 {
		t.Fatalf("expected 4 plans, got %d", len(plans))
	}

	planActiveMap := map[int64]bool{}
	for _, p := range plans {
		planActiveMap[p.ID] = p.Active
	}

	if !planActiveMap[1] {
		t.Errorf("expected plan 1 to be active")
	}
	if !planActiveMap[2] {
		t.Errorf("expected plan 2 to be active")
	}
	if planActiveMap[3] {
		t.Errorf("expected plan 3 to be inactive")
	}
	if planActiveMap[4] {
		t.Errorf("expected plan 4 to be inactive")
	}

	// Test with Active = true filter
	isTrue := true
	reqActive := &dto.PlanListRequest{Page: 1, PageSize: 10, Active: &isTrue}
	respActive := listSvc.Run(reqActive)
	if respActive.GetStatusCode() != 200 {
		t.Fatalf("PlanList with active=true failed: %s", respActive.GetMessage())
	}

	// Test with Active = false filter
	isFalse := false
	reqInactive := &dto.PlanListRequest{Page: 1, PageSize: 10, Active: &isFalse}
	respInactive := listSvc.Run(reqInactive)
	if respInactive.GetStatusCode() != 200 {
		t.Fatalf("PlanList with active=false failed: %s", respInactive.GetMessage())
	}
}

func TestPlanService_NotebookAndPayPerView(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	// 1. Create Notebook Plan (PlanType = 3)
	createSvc := NewPlanCreate(repo, logger)
	price3 := 150.0
	createReq3 := &dto.PlanCreateRequest{
		Nickname:  "cliente_1",
		PlanStart: "2026-01-01",
		PlanType:  3,
		Price:     &price3,
		Items:     []dto.PlanItemRequest{{ServiceID: 1, OrderIndex: 1}},
		Notebook: &dto.PlanNotebookRequest{
			PaymentDay: 10,
		},
	}
	resp3 := createSvc.Run(createReq3)
	if resp3.GetStatusCode() != 200 {
		t.Fatalf("PlanCreate Notebook failed: %s", resp3.GetMessage())
	}
	nbID := resp3.(*dto.PlanCreateResponse).PlanID

	// 2. Create PayPerView Plan (PlanType = 4)
	price4 := 200.0
	createReq4 := &dto.PlanCreateRequest{
		Nickname:  "cliente_2",
		PlanStart: "2026-01-01",
		PlanType:  4,
		Price:     &price4,
		Items:     []dto.PlanItemRequest{{ServiceID: 1, OrderIndex: 1}},
		PayPerView: &dto.PlanPayPerViewRequest{
			PaymentTerm: 45,
		},
	}
	resp4 := createSvc.Run(createReq4)
	if resp4.GetStatusCode() != 200 {
		t.Fatalf("PlanCreate PayPerView failed: %s", resp4.GetMessage())
	}
	ppvID := resp4.(*dto.PlanCreateResponse).PlanID

	// 3. List plans and verify
	listSvc := NewPlanList(repo, logger)
	listResp := listSvc.Run(&dto.PlanListRequest{Page: 1, PageSize: 10})
	if listResp.GetStatusCode() != 200 {
		t.Fatalf("PlanList failed: %s", listResp.GetMessage())
	}
	plans := listResp.(*dto.PlanListResponse).Plans
	if len(plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(plans))
	}

	var nbPlan, ppvPlan *dto.PlanDTO
	for i := range plans {
		switch plans[i].ID {
		case nbID:
			nbPlan = &plans[i]
		case ppvID:
			ppvPlan = &plans[i]
		}
	}
	if nbPlan == nil || nbPlan.PlanTypeName != "Caderneta" || nbPlan.Notebook == nil || nbPlan.Notebook.PaymentDay != 10 {
		t.Errorf("unexpected notebook plan: %+v", nbPlan)
	}
	if ppvPlan == nil || ppvPlan.PlanTypeName != "Pay-per-view" || ppvPlan.PayPerView == nil || ppvPlan.PayPerView.PaymentTerm != 45 {
		t.Errorf("unexpected payperview plan: %+v", ppvPlan)
	}

	// 4. Update PayPerView plan
	updateSvc := NewPlanUpdate(repo, logger)
	updateReq4 := &dto.PlanUpdateRequest{
		ID: ppvID,
		PayPerView: &dto.PlanPayPerViewRequest{
			PaymentTerm: 60,
		},
	}
	upResp4 := updateSvc.Run(updateReq4)
	if upResp4.GetStatusCode() != 200 {
		t.Fatalf("PlanUpdate PayPerView failed: %s", upResp4.GetMessage())
	}
	p4Saved, _ := repo.GetPlanByID(ppvID)
	if p4Saved.PayPerView == nil || p4Saved.PayPerView.PaymentTerm != 60 {
		t.Errorf("expected updated payment term 60, got %+v", p4Saved.PayPerView)
	}
}

type spyInvoicingService struct {
	called bool
	req    *dto.InvoicingRequest
}

func (s *spyInvoicingService) Run(input port.InDTO) port.OutDTO {
	s.called = true
	if r, ok := input.(*dto.InvoicingRequest); ok {
		s.req = r
	}
	return dto.NewInvoicingResponse(200, "OK", "fatura gerada com sucesso")
}

func TestPlanCreate_AutomaticPackageInvoicing(t *testing.T) {
	logger := &mockLogger{}

	t.Run("Creating package plan automatically creates invoice for package plan", func(t *testing.T) {
		repo := newMockPlanRepoForService()
		createSvc := NewPlanCreate(repo, logger)

		itemPrice := 50.0
		planEnd := "2027-10-01"
		createReq := &dto.PlanCreateRequest{
			Nickname:  "cliente_1",
			PlanStart: "2027-09-01",
			PlanEnd:   &planEnd,
			PlanType:  int(domain.PlanTypePackage),
			Items: []dto.PlanItemRequest{
				{ServiceID: 1, OrderIndex: 1, Price: &itemPrice},
			},
			Package: &dto.PlanPackageRequest{
				Quantity:    4,
				PaymentDate: "2027-09-15",
			},
		}

		resp := createSvc.Run(createReq)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("PlanCreate failed: %s", resp.GetMessage())
		}
		createdPlanID := resp.(*dto.PlanCreateResponse).PlanID
		if createdPlanID <= 0 {
			t.Fatalf("expected positive plan ID, got %d", createdPlanID)
		}

		// Verify that an invoice was generated
		if len(repo.invoices) != 1 {
			t.Fatalf("expected 1 invoice created, got %d", len(repo.invoices))
		}
		inv := repo.invoices[0]
		if inv.PlanID != createdPlanID {
			t.Errorf("expected invoice PlanID %d, got %d", createdPlanID, inv.PlanID)
		}
		expectedAmount := 200.0 // 4 * 50
		if inv.Amount != expectedAmount {
			t.Errorf("expected invoice Amount %.2f, got %.2f", expectedAmount, inv.Amount)
		}
		expectedInvoicingDate := time.Date(2027, 9, 1, 0, 0, 0, 0, time.UTC)
		if !inv.InvoicingDate.Equal(expectedInvoicingDate) {
			t.Errorf("expected InvoicingDate %v, got %v", expectedInvoicingDate, inv.InvoicingDate)
		}
		expectedDueDate := time.Date(2027, 9, 15, 0, 0, 0, 0, time.UTC)
		if !inv.DueDate.Equal(expectedDueDate) {
			t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
		}
		expectedPeriodStart := time.Date(2027, 9, 1, 0, 0, 0, 0, time.UTC)
		if !inv.PeriodStart.Equal(expectedPeriodStart) {
			t.Errorf("expected PeriodStart %v, got %v", expectedPeriodStart, inv.PeriodStart)
		}
		expectedPeriodEnd := time.Date(2027, 10, 1, 0, 0, 0, 0, time.UTC)
		if !inv.PeriodEnd.Equal(expectedPeriodEnd) {
			t.Errorf("expected PeriodEnd %v, got %v", expectedPeriodEnd, inv.PeriodEnd)
		}
		if len(inv.Items) != 1 {
			t.Fatalf("expected 1 invoice item, got %d", len(inv.Items))
		}
		expectedDesc := "Aula Canto 30 - 01/09/2027 a 01/10/2027"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected item description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})

	t.Run("Creating agenda plan does not automatically create invoice", func(t *testing.T) {
		repo := newMockPlanRepoForService()
		createSvc := NewPlanCreate(repo, logger)

		initPrice := 100.0
		createReq := &dto.PlanCreateRequest{
			Nickname:  "cliente_1",
			PlanStart: "2026-09-01",
			PlanType:  int(domain.PlanTypeAgenda),
			Price:     &initPrice,
			Items:     []dto.PlanItemRequest{{ServiceID: 1, OrderIndex: 1}},
			Agenda: &dto.PlanAgendaRequest{
				Recurrence: 1,
				WeekDay:    2,
				TermType:   1,
				PaymentDay: 5,
			},
		}

		resp := createSvc.Run(createReq)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("PlanCreate failed: %s", resp.GetMessage())
		}
		if len(repo.invoices) != 0 {
			t.Errorf("expected 0 invoices created for agenda plan, got %d", len(repo.invoices))
		}
	})

	t.Run("Verifies exact parameters passed to invoicing service for package plan", func(t *testing.T) {
		repo := newMockPlanRepoForService()
		spyInvoicing := &spyInvoicingService{}
		createSvc := NewPlanCreateWithInvoicing(repo, logger, spyInvoicing)

		itemPrice := 50.0
		planEnd := "2026-11-05"
		createReq := &dto.PlanCreateRequest{
			Nickname:  "cliente_1",
			PlanStart: "2026-10-05",
			PlanEnd:   &planEnd,
			PlanType:  int(domain.PlanTypePackage),
			Items: []dto.PlanItemRequest{
				{ServiceID: 1, OrderIndex: 1, Price: &itemPrice},
			},
			Package: &dto.PlanPackageRequest{
				Quantity:    2,
				PaymentDate: "2026-10-20",
			},
		}

		resp := createSvc.Run(createReq)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("PlanCreate failed: %s", resp.GetMessage())
		}
		createdPlanID := resp.(*dto.PlanCreateResponse).PlanID

		if !spyInvoicing.called {
			t.Fatal("expected invoicing service to be called")
		}
		if spyInvoicing.req == nil {
			t.Fatal("expected invoicing request to be non-nil")
		}
		if spyInvoicing.req.PlanID == nil || *spyInvoicing.req.PlanID != createdPlanID {
			t.Errorf("expected PlanID %d, got %v", createdPlanID, spyInvoicing.req.PlanID)
		}
		if spyInvoicing.req.PlanType == nil || *spyInvoicing.req.PlanType != int(domain.PlanTypePackage) {
			t.Errorf("expected PlanType %d, got %v", int(domain.PlanTypePackage), spyInvoicing.req.PlanType)
		}
		expectedDate := "2026-10-05"
		if spyInvoicing.req.StartDate == nil || *spyInvoicing.req.StartDate != expectedDate {
			t.Errorf("expected StartDate %q, got %v", expectedDate, spyInvoicing.req.StartDate)
		}
		if spyInvoicing.req.EndDate == nil || *spyInvoicing.req.EndDate != expectedDate {
			t.Errorf("expected EndDate %q, got %v", expectedDate, spyInvoicing.req.EndDate)
		}
	})
}


