package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

type mockLogger struct{}

func (m *mockLogger) IPrintf(level int, format string, v ...interface{}) {}
func (m *mockLogger) Close()                                             {}

type mockInvoicingRepo struct {
	invoices      []domain.Invoice
	nextInvoiceID int64
	sessions      []domain.Session
	deletedIDs    []int64
}

func newMockInvoicingRepo() *mockInvoicingRepo {
	return &mockInvoicingRepo{
		nextInvoiceID: 1,
	}
}

func (m *mockInvoicingRepo) BeginTransaction() error    { return nil }
func (m *mockInvoicingRepo) CommitTransaction() error   { return nil }
func (m *mockInvoicingRepo) RollbackTransaction() error { return nil }
func (m *mockInvoicingRepo) Save(model interface{}) error {
	return nil
}
func (m *mockInvoicingRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) FindCount(conditions map[string]interface{}) (int64, error) {
	return 0, nil
}
func (m *mockInvoicingRepo) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) Close() error { return nil }
func (m *mockInvoicingRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) GetCustomerByID(id int64) (*domain.Customer, error) { return nil, nil }
func (m *mockInvoicingRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) GetVendor(nickname string) (*domain.Vendor, error) { return nil, nil }
func (m *mockInvoicingRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	return nil, nil
}
func (m *mockInvoicingRepo) GetServiceByID(id int64) (*domain.Service, error) { return nil, nil }
func (m *mockInvoicingRepo) SavePlan(plan *domain.Plan) error                 { return nil }
func (m *mockInvoicingRepo) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return nil, 0, nil
}
func (m *mockInvoicingRepo) GetPlanByID(id int64) (*domain.Plan, error) { return nil, nil }
func (m *mockInvoicingRepo) DeletePlan(id int64) error                    { return nil }
func (m *mockInvoicingRepo) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}

func (m *mockInvoicingRepo) SaveInvoice(invoice *domain.Invoice) error {
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

func (m *mockInvoicingRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	for _, inv := range m.invoices {
		if inv.ID == id && inv.DeletedAt == nil {
			return &inv, nil
		}
	}
	return nil, nil
}

func (m *mockInvoicingRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
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

func (m *mockInvoicingRepo) DeleteInvoice(id int64) error {
	m.deletedIDs = append(m.deletedIDs, id)
	now := time.Now()
	for i := range m.invoices {
		if m.invoices[i].ID == id {
			m.invoices[i].DeletedAt = &now
		}
	}
	return nil
}

func (m *mockInvoicingRepo) FindActivePlansForInvoicing(vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	return nil, nil
}

func (m *mockInvoicingRepo) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
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

func (m *mockInvoicingRepo) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) {
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
func stringPtr(v string) *string  { return &v }
func intPtr(v int) *int          { return &v }

// Test Package Strategy
func TestPackageStrategy(t *testing.T) {
	repo := newMockInvoicingRepo()
	logger := &mockLogger{}
	strat := NewPackageStrategy(repo, logger)

	planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	paymentDate := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

	desc1 := "Aula Canto 30"
	desc2 := "Aula Piano 45"
	desc3 := "Aula Teoria"

	plan := &domain.Plan{
		ID:         10,
		CustomerID: 100,
		PlanStart:  planStart,
		PlanType:   2,
		Package: &domain.PlanPackage{
			ID:          10,
			Quantity:    10,
			PaymentDate: paymentDate,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0), Service: &domain.Service{Description: &desc1}},
			{ServiceID: 2, OrderIndex: 2, Price: floatPtr(60.0), Service: &domain.Service{Description: &desc2}},
			{ServiceID: 3, OrderIndex: 3, Price: floatPtr(40.0), Service: &domain.Service{Description: &desc3}},
		},
	}

	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	// 1. First process: generates invoice
	if err := strat.Process(plan, startDate, endDate, false); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(repo.invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(repo.invoices))
	}

	inv := repo.invoices[0]
	if inv.PlanID != 10 {
		t.Errorf("expected plan_id 10, got %d", inv.PlanID)
	}
	if !inv.InvoicingDate.Equal(planStart) {
		t.Errorf("expected InvoicingDate %v, got %v", planStart, inv.InvoicingDate)
	}
	if len(inv.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(inv.Items))
	}

	// 10 distributed among 3 items: 4, 3, 3
	if inv.Items[0].InvoicedQuantity != 4 {
		t.Errorf("expected item 1 quantity 4, got %d", inv.Items[0].InvoicedQuantity)
	}
	if inv.Items[1].InvoicedQuantity != 3 {
		t.Errorf("expected item 2 quantity 3, got %d", inv.Items[1].InvoicedQuantity)
	}
	if inv.Items[2].InvoicedQuantity != 3 {
		t.Errorf("expected item 3 quantity 3, got %d", inv.Items[2].InvoicedQuantity)
	}

	// Amount: 4*50 + 3*60 + 3*40 = 200 + 180 + 120 = 500
	if inv.Amount != 500.0 {
		t.Errorf("expected Amount 500.0, got %.2f", inv.Amount)
	}

	// 2. Second process without reprocess: skips
	if err := strat.Process(plan, startDate, endDate, false); err != nil {
		t.Fatalf("Process failed on repeat: %v", err)
	}
	if len(repo.invoices) != 1 {
		t.Fatalf("expected still 1 invoice, got %d", len(repo.invoices))
	}

	// 3. Third process with reprocess = true: soft deletes previous invoice and creates new
	if err := strat.Process(plan, startDate, endDate, true); err != nil {
		t.Fatalf("Process with reprocess failed: %v", err)
	}
	if len(repo.deletedIDs) != 1 {
		t.Errorf("expected 1 deleted invoice, got %d", len(repo.deletedIDs))
	}
	if len(repo.invoices) != 2 {
		t.Errorf("expected 2 total invoices in repo (1 deleted, 1 active), got %d", len(repo.invoices))
	}
}

// Test Notebook Strategy
func TestNotebookStrategy(t *testing.T) {
	repo := newMockInvoicingRepo()
	logger := &mockLogger{}
	strat := NewNotebookStrategy(repo, logger)

	planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:         20,
		CustomerID: 200,
		PlanStart:  planStart,
		PlanType:   3,
		Notebook: &domain.PlanNotebook{
			ID:         20,
			PaymentDay: 10,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(80.0)},
		},
	}

	// Add sessions in August 2026 (previous month to September 2026)
	repo.sessions = []domain.Session{
		{ID: 101, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC), SessionStatus: "realizada"},
		{ID: 102, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), SessionStatus: "realizada"},
		{ID: 103, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC), SessionStatus: "cancelada_cobrar"},
		{ID: 104, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC), SessionStatus: "cancelada_nao_cobrar"}, // should NOT be billed
	}

	// Process for September 2026
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	if err := strat.Process(plan, startDate, endDate, false); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(repo.invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(repo.invoices))
	}

	inv := repo.invoices[0]
	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}

	item := inv.Items[0]
	// 3 billable sessions in August
	if item.InvoicedQuantity != 3 {
		t.Errorf("expected InvoicedQuantity 3, got %d", item.InvoicedQuantity)
	}
	if item.RealizedQuantity != 3 {
		t.Errorf("expected RealizedQuantity 3, got %d", item.RealizedQuantity)
	}
	if len(item.ItemSessions) != 3 {
		t.Errorf("expected 3 session links, got %d", len(item.ItemSessions))
	}
	// Amount: 3 * 80 = 240
	if inv.Amount != 240.0 {
		t.Errorf("expected Amount 240.0, got %.2f", inv.Amount)
	}
}

// Test Pay-Per-View Strategy
func TestPayPerViewStrategy(t *testing.T) {
	repo := newMockInvoicingRepo()
	logger := &mockLogger{}
	strat := NewPayPerViewStrategy(repo, logger)

	planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:         30,
		CustomerID: 300,
		PlanStart:  planStart,
		PlanType:   4,
		PayPerView: &domain.PlanPayPerView{
			ID:          30,
			PaymentTerm: 5,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(120.0)},
		},
	}

	sessDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	repo.sessions = []domain.Session{
		{ID: 201, CustomerID: 300, ServiceID: 1, SessionDate: sessDate, SessionStatus: "realizada"},
	}

	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	if err := strat.Process(plan, startDate, endDate, false); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(repo.invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(repo.invoices))
	}

	inv := repo.invoices[0]
	if inv.Amount != 120.0 {
		t.Errorf("expected Amount 120.0, got %.2f", inv.Amount)
	}
	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}
	if inv.Items[0].InvoicedQuantity != 1 || inv.Items[0].RealizedQuantity != 1 {
		t.Errorf("expected quantities 1, got inv=%d, real=%d", inv.Items[0].InvoicedQuantity, inv.Items[0].RealizedQuantity)
	}
	if len(inv.Items[0].ItemSessions) != 1 || inv.Items[0].ItemSessions[0].SessionID != 201 {
		t.Errorf("expected session link to 201, got %v", inv.Items[0].ItemSessions)
	}
}

// Test Agenda Strategy
func TestAgendaStrategy(t *testing.T) {
	repo := newMockInvoicingRepo()
	logger := &mockLogger{}
	strat := NewAgendaStrategy(repo, logger)

	planStart := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	// Pre-paid, Weekly on Tuesdays (WeekDay 3 in 1..7 scheme: Sunday=1, Monday=2, Tuesday=3)
	// Payment day: 5th of month
	plan := &domain.Plan{
		ID:         40,
		CustomerID: 400,
		PlanStart:  planStart,
		PlanType:   1,
		Agenda: &domain.PlanAgenda{
			ID:                  40,
			Recurrence:          int(domain.RecurrenceWeekly),
			WeekDay:             int(domain.WeekDayTuesday),
			TermType:            int(domain.TermTypePrePaid),
			PaymentDay:          5,
			MonthlyServiceLimit: intPtr(4),
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(75.0)},
		},
	}

	// Month: September 2026.
	// Tuesdays in Sept 2026: Sept 1, 8, 15, 22, 29 (5 Tuesdays).
	// But MonthlyServiceLimit is 4, so it should be capped at 4!
	startDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	if err := strat.Process(plan, startDate, endDate, false); err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(repo.invoices) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(repo.invoices))
	}

	inv := repo.invoices[0]
	// invoice_date: min(2026-09-01, 2026-08-15) -> 2026-08-15
	if !inv.InvoicingDate.Equal(planStart) {
		t.Errorf("expected InvoicingDate %v, got %v", planStart, inv.InvoicingDate)
	}

	// due_date: 2026-09-05
	expectedDueDate := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}

	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}

	// Capped at 4
	if inv.Items[0].InvoicedQuantity != 4 {
		t.Errorf("expected InvoicedQuantity 4, got %d", inv.Items[0].InvoicedQuantity)
	}
	// Amount: 4 * 75 = 300
	if inv.Amount != 300.0 {
		t.Errorf("expected Amount 300.0, got %.2f", inv.Amount)
	}
}
