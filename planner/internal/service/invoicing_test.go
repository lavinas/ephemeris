package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoicingRepoForService struct {
	plans         []domain.Plan
	invoices      []domain.Invoice
	nextInvoiceID int64
	sessions      []domain.Session
	deletedIDs    []int64
}

func newMockInvoicingRepoForService() *mockInvoicingRepoForService {
	return &mockInvoicingRepoForService{
		nextInvoiceID: 1,
	}
}

func (m *mockInvoicingRepoForService) BeginTransaction() error    { return nil }
func (m *mockInvoicingRepoForService) CommitTransaction() error   { return nil }
func (m *mockInvoicingRepoForService) RollbackTransaction() error { return nil }
func (m *mockInvoicingRepoForService) Save(model interface{}) error {
	return nil
}
func (m *mockInvoicingRepoForService) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) FindCount(conditions map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *mockInvoicingRepoForService) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) Close() error { return nil }
func (m *mockInvoicingRepoForService) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, Nickname: "test_customer"}, nil
}
func (m *mockInvoicingRepoForService) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) GetVendor(nickname string) (*domain.Vendor, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) FindServices(vendorID int64) ([]domain.Service, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id, Name: "Test Service"}, nil
}
func (m *mockInvoicingRepoForService) SavePlan(plan *domain.Plan) error { return nil }
func (m *mockInvoicingRepoForService) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return nil, 0, nil
}
func (m *mockInvoicingRepoForService) GetPlanByID(id int64) (*domain.Plan, error) {
	for i := range m.plans {
		if m.plans[i].ID == id {
			return &m.plans[i], nil
		}
	}
	return nil, nil
}
func (m *mockInvoicingRepoForService) DeletePlan(id int64) error                    { return nil }
func (m *mockInvoicingRepoForService) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}

func (m *mockInvoicingRepoForService) SaveInvoice(invoice *domain.Invoice) error {
	if invoice.ID == 0 {
		invoice.ID = m.nextInvoiceID
		m.nextInvoiceID++
	}
	for i := range invoice.Items {
		if invoice.Items[i].ID == 0 {
			invoice.Items[i].ID = int64(i + 1)
			invoice.Items[i].InvoiceID = invoice.ID
		}
	}
	m.invoices = append(m.invoices, *invoice)
	return nil
}

func (m *mockInvoicingRepoForService) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	for _, inv := range m.invoices {
		if inv.ID == id && inv.DeletedAt == nil {
			return &inv, nil
		}
	}
	return nil, nil
}

func (m *mockInvoicingRepoForService) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	var result []domain.Invoice
	planID, hasPlanID := conditions["plan_id = ?"]
	for _, inv := range m.invoices {
		if inv.DeletedAt != nil {
			continue
		}
		if hasPlanID && inv.PlanID != planID.(int64) {
			continue
		}
		if start, ok := conditions["invoicing_date >= ?"]; ok {
			if inv.InvoicingDate.Before(start.(time.Time)) {
				continue
			}
		}
		if end, ok := conditions["invoicing_date <= ?"]; ok {
			if inv.InvoicingDate.After(end.(time.Time)) {
				continue
			}
		}
		if start, ok := conditions["due_date >= ?"]; ok {
			if inv.DueDate.Before(start.(time.Time)) {
				continue
			}
		}
		if end, ok := conditions["due_date <= ?"]; ok {
			if inv.DueDate.After(end.(time.Time)) {
				continue
			}
		}
		result = append(result, inv)
	}
	return result, nil
}

func (m *mockInvoicingRepoForService) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	invoices, err := m.FindInvoices(conditions, orderBy...)
	if err != nil {
		return nil, 0, 0, err
	}
	var totalAmount float64
	for _, inv := range invoices {
		totalAmount += inv.Amount
	}
	return invoices, int64(len(invoices)), totalAmount, nil
}

func (m *mockInvoicingRepoForService) DeleteInvoice(id int64) error {
	m.deletedIDs = append(m.deletedIDs, id)
	now := time.Now()
	for i := range m.invoices {
		if m.invoices[i].ID == id {
			m.invoices[i].DeletedAt = &now
		}
	}
	return nil
}

func (m *mockInvoicingRepoForService) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	var result []domain.Plan
	for _, p := range m.plans {
		if planID != nil && *planID != 0 && p.ID != *planID {
			continue
		}
		if vendorID != nil && *vendorID != 0 && p.Customer != nil && p.Customer.VendorID != *vendorID {
			continue
		}
		if customerID != nil && *customerID != 0 && p.CustomerID != *customerID {
			continue
		}
		if planType != nil && *planType != 0 && p.PlanType != *planType {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

func (m *mockInvoicingRepoForService) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	var result []domain.Session
	for _, s := range m.sessions {
		if s.CustomerID == customerID && s.ServiceID == serviceID &&
			!s.SessionDate.Before(startDate) && !s.SessionDate.After(endDate) &&
			s.SessionStatus != string(domain.StatusCancelNotCharge) && s.DeletedAt == nil {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockInvoicingRepoForService) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) {
	for _, inv := range m.invoices {
		if inv.DeletedAt != nil {
			continue
		}
		for _, item := range inv.Items {
			for _, sessLink := range item.ItemSessions {
				if sessLink.SessionID == sessionID {
					return &inv, nil
				}
			}
		}
	}
	return nil, nil
}

func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int          { return &v }
func stringPtr(v string) *string  { return &v }

func TestInvoicingService_Run(t *testing.T) {
	logger := &mockLogger{}

	t.Run("Validation failure returns 400", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		svc := NewInvoicing(repo, logger)
		badPlanType := 99
		req := &dto.InvoicingRequest{PlanType: &badPlanType}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 400 {
			t.Errorf("expected 400, got %d", resp.GetStatusCode())
		}
	})

	t.Run("No active plans returns 200 with notice message", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Errorf("expected 200, got %d", resp.GetStatusCode())
		}
	})

	t.Run("Success invoicing package plan", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		price := 200.0
		repo.plans = []domain.Plan{
			{
				ID:         1,
				CustomerID: 10,
				PlanStart:  planStart,
				PlanType:   2,
				Price:      &price,
				Package: &domain.PlanPackage{
					ID:          1,
					Quantity:    4,
					PaymentDate: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
				},
				Items: []domain.PlanItem{
					{ServiceID: 1, OrderIndex: 1},
				},
			},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Errorf("expected 200, got %d (msg: %s)", resp.GetStatusCode(), resp.GetMessage())
		}
		if len(repo.invoices) != 1 {
			t.Errorf("expected 1 invoice created, got %d", len(repo.invoices))
		}
	})

	t.Run("Success invoicing agenda plan", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		repo.plans = []domain.Plan{
			{
				ID:         2,
				CustomerID: 20,
				PlanStart:  planStart,
				PlanType:   1,
				Agenda: &domain.PlanAgenda{
					ID:         2,
					Recurrence: int(domain.RecurrenceWeekly),
					WeekDay:    int(domain.WeekDayMonday),
					TermType:   int(domain.TermTypePrePaid),
					PaymentDay: 5,
				},
				Items: []domain.PlanItem{
					{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0)},
				},
			},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Errorf("expected 200, got %d (msg: %s)", resp.GetStatusCode(), resp.GetMessage())
		}
		if len(repo.invoices) != 1 {
			t.Errorf("expected 1 invoice created, got %d", len(repo.invoices))
		}
	})

	t.Run("Success invoicing notebook plan", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		repo.plans = []domain.Plan{
			{
				ID:         3,
				CustomerID: 30,
				PlanStart:  planStart,
				PlanType:   3,
				Notebook: &domain.PlanNotebook{
					ID:         3,
					PaymentDay: 10,
				},
				Items: []domain.PlanItem{
					{ServiceID: 1, OrderIndex: 1, Price: floatPtr(70.0)},
				},
			},
		}
		repo.sessions = []domain.Session{
			{ID: 301, CustomerID: 30, ServiceID: 1, SessionDate: time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), SessionStatus: "realizada"},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Errorf("expected 200, got %d (msg: %s)", resp.GetStatusCode(), resp.GetMessage())
		}
		if len(repo.invoices) != 1 {
			t.Errorf("expected 1 invoice created, got %d", len(repo.invoices))
		}
	})

	t.Run("Success invoicing pay-per-view plan", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		repo.plans = []domain.Plan{
			{
				ID:         4,
				CustomerID: 40,
				PlanStart:  planStart,
				PlanType:   4,
				PayPerView: &domain.PlanPayPerView{
					ID:          4,
					PaymentTerm: 5,
				},
				Items: []domain.PlanItem{
					{ServiceID: 1, OrderIndex: 1, Price: floatPtr(120.0)},
				},
			},
		}
		repo.sessions = []domain.Session{
			{ID: 401, CustomerID: 40, ServiceID: 1, SessionDate: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC), SessionStatus: "realizada"},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Errorf("expected 200, got %d (msg: %s)", resp.GetStatusCode(), resp.GetMessage())
		}
		if len(repo.invoices) != 1 {
			t.Errorf("expected 1 invoice created, got %d", len(repo.invoices))
		}
	})

	t.Run("Idempotency and reprocess", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		price := 200.0
		repo.plans = []domain.Plan{
			{
				ID:         5,
				CustomerID: 50,
				PlanStart:  planStart,
				PlanType:   2,
				Price:      &price,
				Package: &domain.PlanPackage{
					ID:          5,
					Quantity:    4,
					PaymentDate: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
				},
				Items: []domain.PlanItem{
					{ServiceID: 1, OrderIndex: 1},
				},
			},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"

		// 1. First run generates 1 invoice
		req := &dto.InvoicingRequest{StartDate: &start, EndDate: &end}
		svc.Run(req)
		if len(repo.invoices) != 1 {
			t.Fatalf("expected 1 invoice, got %d", len(repo.invoices))
		}

		// 2. Second run without reprocess skips
		svc.Run(req)
		if len(repo.invoices) != 1 {
			t.Fatalf("expected still 1 invoice, got %d", len(repo.invoices))
		}

		// 3. Third run with reprocess = true deletes previous and creates new
		reprocessVal := true
		reqReprocess := &dto.InvoicingRequest{StartDate: &start, EndDate: &end, Reprocess: &reprocessVal}
		svc.Run(reqReprocess)
		if len(repo.deletedIDs) != 1 {
			t.Errorf("expected 1 deleted invoice, got %d", len(repo.deletedIDs))
		}
		if len(repo.invoices) != 2 {
			t.Errorf("expected 2 total invoices (1 deleted, 1 active), got %d", len(repo.invoices))
		}
	})

	t.Run("Process specific plan using plan_id", func(t *testing.T) {
		repo := newMockInvoicingRepoForService()
		logger := &mockLogger{}
		price := 200.0
		planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

		repo.plans = []domain.Plan{
			{
				ID:         10,
				CustomerID: 1,
				PlanStart:  planStart,
				PlanType:   2,
				Price:      &price,
				Package: &domain.PlanPackage{
					ID:          1,
					Quantity:    4,
					PaymentDate: time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
				},
				Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1}},
			},
			{
				ID:         20,
				CustomerID: 2,
				PlanStart:  planStart,
				PlanType:   2,
				Price:      &price,
				Package: &domain.PlanPackage{
					ID:          2,
					Quantity:    8,
					PaymentDate: time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
				},
				Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1}},
			},
		}

		svc := NewInvoicing(repo, logger)
		start := "2026-09-01"
		end := "2026-09-30"
		planID := int64(20)

		req := &dto.InvoicingRequest{
			PlanID:    &planID,
			StartDate: &start,
			EndDate:   &end,
		}
		resp := svc.Run(req)
		if resp.GetStatusCode() != 200 {
			t.Fatalf("invoicing failed: %s", resp.GetMessage())
		}
		if len(repo.invoices) != 1 {
			t.Fatalf("expected 1 invoice for plan 20, got %d", len(repo.invoices))
		}
		if repo.invoices[0].PlanID != 20 {
			t.Errorf("expected invoice plan_id 20, got %d", repo.invoices[0].PlanID)
		}
	})
}
