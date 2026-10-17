package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockReconciliationRepo struct {
	sessions         map[int64]*domain.Session
	invoices         map[int64]*domain.Invoice
	invoiceItems     map[int64]*domain.InvoiceItem
	itemSessions     map[int64]int64 // sessionID -> invoiceItemID
	ppvPlans         map[int64]*domain.Plan
	savedInvoices    []*domain.Invoice
	nextInvoiceID    int64
	nextItemAssocID  int64
}

func newMockReconciliationRepo() *mockReconciliationRepo {
	return &mockReconciliationRepo{
		sessions:        make(map[int64]*domain.Session),
		invoices:        make(map[int64]*domain.Invoice),
		invoiceItems:    make(map[int64]*domain.InvoiceItem),
		itemSessions:    make(map[int64]int64),
		ppvPlans:        make(map[int64]*domain.Plan),
		nextInvoiceID:   100,
		nextItemAssocID: 1,
	}
}

func (m *mockReconciliationRepo) GetSessionByID(id int64) (*domain.Session, error) {
	s, ok := m.sessions[id]
	if !ok || s.DeletedAt != nil {
		return nil, nil
	}
	return s, nil
}

func (m *mockReconciliationRepo) FindMatchingInvoiceItem(customerID, serviceID int64, sessionDate time.Time) (*domain.InvoiceItem, *domain.Invoice, error) {
	for _, item := range m.invoiceItems {
		if item.DeletedAt != nil || item.ServiceID != serviceID {
			continue
		}
		inv, ok := m.invoices[item.InvoiceID]
		if !ok || inv.DeletedAt != nil || inv.Plan == nil || inv.Plan.CustomerID != customerID {
			continue
		}
		d := sessionDate.Format("2006-01-02")
		pStart := inv.PeriodStart.Format("2006-01-02")
		pEnd := inv.PeriodEnd.Format("2006-01-02")
		if d >= pStart && d <= pEnd {
			return item, inv, nil
		}
	}
	return nil, nil, nil
}

func (m *mockReconciliationRepo) IsSessionReconciled(sessionID int64) (bool, error) {
	_, ok := m.itemSessions[sessionID]
	return ok, nil
}

func (m *mockReconciliationRepo) FindUnreconciledSessions(customerID, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	var result []domain.Session
	pStart := startDate.Format("2006-01-02")
	pEnd := endDate.Format("2006-01-02")

	for _, sess := range m.sessions {
		if sess.DeletedAt != nil || sess.CustomerID != customerID || sess.ServiceID != serviceID {
			continue
		}
		if _, reconciled := m.itemSessions[sess.ID]; reconciled {
			continue
		}
		d := sess.SessionDate.Format("2006-01-02")
		if d >= pStart && d <= pEnd {
			result = append(result, *sess)
		}
	}
	return result, nil
}

func (m *mockReconciliationRepo) AssociateSessionToInvoiceItem(invoiceItemID, sessionID int64) error {
	m.itemSessions[sessionID] = invoiceItemID
	if item, ok := m.invoiceItems[invoiceItemID]; ok {
		item.RealizedQuantity++
	}
	return nil
}

func (m *mockReconciliationRepo) DissociateSessionFromInvoiceItem(sessionID int64) (*domain.InvoiceItem, error) {
	itemID, ok := m.itemSessions[sessionID]
	if !ok {
		return nil, nil
	}
	delete(m.itemSessions, sessionID)
	if item, ok := m.invoiceItems[itemID]; ok {
		if item.RealizedQuantity > 0 {
			item.RealizedQuantity--
		}
		return item, nil
	}
	return nil, nil
}

func (m *mockReconciliationRepo) FindActivePayPerViewPlan(customerID, serviceID int64, sessionDate time.Time) (*domain.Plan, *domain.PlanItem, error) {
	for _, plan := range m.ppvPlans {
		if plan.DeletedAt != nil || plan.CustomerID != customerID || plan.PlanType != int(domain.PlanTypePayPerView) {
			continue
		}
		d := sessionDate.Format("2006-01-02")
		pStart := plan.PlanStart.Format("2006-01-02")
		if d < pStart {
			continue
		}
		if plan.PlanEnd != nil && !plan.PlanEnd.IsZero() && d > plan.PlanEnd.Format("2006-01-02") {
			continue
		}
		for i := range plan.Items {
			if plan.Items[i].ServiceID == serviceID && plan.Items[i].DeletedAt == nil {
				return plan, &plan.Items[i], nil
			}
		}
	}
	return nil, nil, nil
}

func (m *mockReconciliationRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	inv, ok := m.invoices[id]
	if !ok || inv.DeletedAt != nil {
		return nil, nil
	}
	return inv, nil
}

func (m *mockReconciliationRepo) SaveInvoice(invoice *domain.Invoice) error {
	if invoice.ID == 0 {
		m.nextInvoiceID++
		invoice.ID = m.nextInvoiceID
	}
	m.invoices[invoice.ID] = invoice
	m.savedInvoices = append(m.savedInvoices, invoice)

	for i := range invoice.Items {
		item := &invoice.Items[i]
		if item.ID == 0 {
			item.ID = int64(len(m.invoiceItems) + 1)
		}
		item.InvoiceID = invoice.ID
		m.invoiceItems[item.ID] = item

		for _, sessAssoc := range item.ItemSessions {
			m.itemSessions[sessAssoc.SessionID] = item.ID
		}
	}
	return nil
}

type mockReconcileLogger struct{}

func (l *mockReconcileLogger) IPrintf(level int, format string, v ...interface{}) {}
func (l *mockReconcileLogger) Close()                                             {}

func TestReconciliation_ReconcileSession_MatchesExistingInvoice(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	// Set up customer, plan, and invoice
	customerID := int64(1)
	serviceID := int64(10)
	plan := &domain.Plan{
		ID:         1,
		CustomerID: customerID,
		PlanType:   int(domain.PlanTypePackage),
	}
	invoice := &domain.Invoice{
		ID:          50,
		PlanID:      plan.ID,
		Plan:        plan,
		PeriodStart: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	item := &domain.InvoiceItem{
		ID:               500,
		InvoiceID:        invoice.ID,
		ServiceID:        serviceID,
		InvoicedQuantity: 10,
		RealizedQuantity: 0,
	}
	invoice.Items = []domain.InvoiceItem{*item}
	repo.invoices[invoice.ID] = invoice
	repo.invoiceItems[item.ID] = item

	// Session created on 2026-10-15
	sessionDate := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	sess := domain.NewSession(customerID, serviceID, sessionDate, 60, string(domain.StatusDone), nil)
	sess.ID = 1001
	repo.sessions[sess.ID] = sess

	err := svc.ReconcileSession(sess.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify association and realized quantity increment
	if repo.itemSessions[sess.ID] != item.ID {
		t.Errorf("expected session %d to be associated with item %d, got %d", sess.ID, item.ID, repo.itemSessions[sess.ID])
	}
	if item.RealizedQuantity != 1 {
		t.Errorf("expected realized quantity 1, got %d", item.RealizedQuantity)
	}
}

func TestReconciliation_ReconcileSession_AlreadyReconciled(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	sess := domain.NewSession(1, 10, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), 60, string(domain.StatusDone), nil)
	sess.ID = 1001
	repo.sessions[sess.ID] = sess

	item := &domain.InvoiceItem{ID: 500, RealizedQuantity: 1}
	repo.invoiceItems[item.ID] = item
	repo.itemSessions[sess.ID] = item.ID // already associated

	err := svc.ReconcileSession(sess.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not increment again
	if item.RealizedQuantity != 1 {
		t.Errorf("expected realized quantity to remain 1, got %d", item.RealizedQuantity)
	}
}

func TestReconciliation_ReconcileSession_NotEligible(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	// Cancel not charge
	sess := domain.NewSession(1, 10, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), 60, string(domain.StatusCancelNotCharge), nil)
	sess.ID = 1001
	repo.sessions[sess.ID] = sess

	err := svc.ReconcileSession(sess.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := repo.itemSessions[sess.ID]; ok {
		t.Errorf("expected ineligible session NOT to be associated")
	}
}

func TestReconciliation_ReconcileSession_PayPerView_GeneratesInvoice(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	customerID := int64(2)
	serviceID := int64(20)
	price := 150.0

	ppvPlan := &domain.Plan{
		ID:         77,
		CustomerID: customerID,
		PlanType:   int(domain.PlanTypePayPerView),
		PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		PayPerView: &domain.PlanPayPerView{PaymentTerm: 5},
		Items: []domain.PlanItem{
			{
				ID:        1,
				ServiceID: serviceID,
				Price:     &price,
				Service:   &domain.Service{ID: serviceID, Description: strPtr("Consulta Avulsa")},
			},
		},
	}
	repo.ppvPlans[ppvPlan.ID] = ppvPlan

	sess := domain.NewSession(customerID, serviceID, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), 60, string(domain.StatusDone), nil)
	sess.ID = 2002
	repo.sessions[sess.ID] = sess

	err := svc.ReconcileSession(sess.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repo.savedInvoices) != 1 {
		t.Fatalf("expected 1 invoice to be saved for PPV, got %d", len(repo.savedInvoices))
	}
	savedInv := repo.savedInvoices[0]
	if savedInv.PlanID != ppvPlan.ID {
		t.Errorf("expected planID %d, got %d", ppvPlan.ID, savedInv.PlanID)
	}
	if len(savedInv.Items) != 1 {
		t.Fatalf("expected 1 item in PPV invoice, got %d", len(savedInv.Items))
	}
	if savedInv.Items[0].RealizedQuantity != 1 {
		t.Errorf("expected realized quantity 1, got %d", savedInv.Items[0].RealizedQuantity)
	}
	if repo.itemSessions[sess.ID] != savedInv.Items[0].ID {
		t.Errorf("expected session %d to be associated with PPV invoice item %d", sess.ID, savedInv.Items[0].ID)
	}
}

func TestReconciliation_DeconcileSession(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	sessionID := int64(1001)
	itemID := int64(500)
	item := &domain.InvoiceItem{ID: itemID, RealizedQuantity: 3}
	repo.invoiceItems[itemID] = item
	repo.itemSessions[sessionID] = itemID

	err := svc.DeconcileSession(sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := repo.itemSessions[sessionID]; ok {
		t.Errorf("expected session association to be removed")
	}
	if item.RealizedQuantity != 2 {
		t.Errorf("expected realized quantity to be decremented to 2, got %d", item.RealizedQuantity)
	}
}

func TestReconciliation_ReconcileInvoice_AssociatesUnreconciledSessions(t *testing.T) {
	repo := newMockReconciliationRepo()
	logger := &mockReconcileLogger{}
	svc := NewReconciliation(repo, logger)

	customerID := int64(3)
	serviceID := int64(30)
	plan := &domain.Plan{
		ID:         3,
		CustomerID: customerID,
		PlanType:   int(domain.PlanTypeAgenda),
	}

	// 2 unreconciled sessions in October
	sess1 := domain.NewSession(customerID, serviceID, time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), 60, string(domain.StatusDone), nil)
	sess1.ID = 3001
	repo.sessions[sess1.ID] = sess1

	sess2 := domain.NewSession(customerID, serviceID, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), 60, string(domain.StatusDone), nil)
	sess2.ID = 3002
	repo.sessions[sess2.ID] = sess2

	// New invoice for October
	invoice := &domain.Invoice{
		ID:          88,
		PlanID:      plan.ID,
		Plan:        plan,
		PeriodStart: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	item := &domain.InvoiceItem{
		ID:               880,
		InvoiceID:        invoice.ID,
		ServiceID:        serviceID,
		InvoicedQuantity: 4,
		RealizedQuantity: 0,
	}
	invoice.Items = []domain.InvoiceItem{*item}
	repo.invoices[invoice.ID] = invoice
	repo.invoiceItems[item.ID] = item

	err := svc.ReconcileInvoice(invoice.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.itemSessions[sess1.ID] != item.ID {
		t.Errorf("expected session 1 to be associated with item %d", item.ID)
	}
	if repo.itemSessions[sess2.ID] != item.ID {
		t.Errorf("expected session 2 to be associated with item %d", item.ID)
	}
	if item.RealizedQuantity != 2 {
		t.Errorf("expected realized quantity 2, got %d", item.RealizedQuantity)
	}
}

type mockReconciliationService struct {
	reconcileSessionCalled bool
	reconcileSessionID     int64
	deconcileSessionCalled bool
	deconcileSessionID     int64
	reconcileInvoiceCalled bool
	reconcileInvoiceID     int64
}

func (m *mockReconciliationService) ReconcileSession(sessionID int64) error {
	m.reconcileSessionCalled = true
	m.reconcileSessionID = sessionID
	return nil
}

func (m *mockReconciliationService) DeconcileSession(sessionID int64) error {
	m.deconcileSessionCalled = true
	m.deconcileSessionID = sessionID
	return nil
}

func (m *mockReconciliationService) ReconcileInvoice(invoiceID int64) error {
	m.reconcileInvoiceCalled = true
	m.reconcileInvoiceID = invoiceID
	return nil
}

type mockSessionRepo struct {
	savedModel interface{}
	sessions   []interface{}
}

func (m *mockSessionRepo) BeginTransaction() error    { return nil }
func (m *mockSessionRepo) CommitTransaction() error   { return nil }
func (m *mockSessionRepo) RollbackTransaction() error { return nil }
func (m *mockSessionRepo) Save(model interface{}) error {
	m.savedModel = model
	if s, ok := model.(*domain.Session); ok {
		if s.ID == 0 {
			s.ID = 999
		}
	}
	return nil
}
func (m *mockSessionRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return m.sessions, nil
}
func (m *mockSessionRepo) FindCount(conditions map[string]interface{}) (int64, error) {
	return int64(len(m.sessions)), nil
}
func (m *mockSessionRepo) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockSessionRepo) Close() error { return nil }
func (m *mockSessionRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return &domain.Customer{ID: 1, Nickname: nickname}, nil
}
func (m *mockSessionRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id}, nil
}
func (m *mockSessionRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	return nil, nil
}
func (m *mockSessionRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return &domain.Service{ID: id}, nil
}
func (m *mockSessionRepo) SavePlan(plan *domain.Plan) error { return nil }
func (m *mockSessionRepo) FindPlans(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Plan, int64, error) {
	return nil, 0, nil
}
func (m *mockSessionRepo) GetPlanByID(id int64) (*domain.Plan, error) { return nil, nil }
func (m *mockSessionRepo) DeletePlan(id int64) error                  { return nil }
func (m *mockSessionRepo) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}
func (m *mockSessionRepo) SaveInvoice(invoice *domain.Invoice) error { return nil }
func (m *mockSessionRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return nil, nil
}
func (m *mockSessionRepo) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *mockSessionRepo) FindInvoicesPaged(page, pageSize int, conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, int64, float64, error) {
	return nil, 0, 0, nil
}
func (m *mockSessionRepo) DeleteInvoice(id int64) error { return nil }
func (m *mockSessionRepo) UpdateInvoice(invoice *domain.Invoice) error { return nil }
func (m *mockSessionRepo) DeleteInvoicesByIDs(ids []int64) error { return nil }
func (m *mockSessionRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	return nil, nil
}
func (m *mockSessionRepo) FindBillableSessions(customerID int64, serviceID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	return nil, nil
}
func (m *mockSessionRepo) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) {
	return nil, nil
}

func TestSessionCreate_TriggersReconciliation(t *testing.T) {
	repo := &mockSessionRepo{}
	logger := &mockReconcileLogger{}
	svc := NewSessionCreate(repo, logger)
	mockReconciler := &mockReconciliationService{}
	svc.SetReconciler(mockReconciler)

	req := &dto.SessionCreateRequest{
		CustomerID: 1,
		Nickname:   "test_customer",
		ServiceID:  10,
		Date:       "2026-10-15",
		Minutes:    60,
		Status:     "realizada",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if !mockReconciler.reconcileSessionCalled {
		t.Errorf("expected ReconcileSession to be called upon session creation")
	}
	if mockReconciler.reconcileSessionID != 999 {
		t.Errorf("expected ReconcileSession to be called with ID 999, got %d", mockReconciler.reconcileSessionID)
	}
}

func TestSessionDelete_TriggersDeconciliation(t *testing.T) {
	sess := domain.NewSession(1, 10, time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC), 60, "realizada", nil)
	sess.ID = 777
	repo := &mockSessionRepo{
		sessions: []interface{}{*sess},
	}
	logger := &mockReconcileLogger{}
	svc := NewSessionDelete(repo, logger)
	mockReconciler := &mockReconciliationService{}
	svc.SetReconciler(mockReconciler)

	req := &dto.SessionDeleteRequest{
		SessionID: 777,
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if !mockReconciler.deconcileSessionCalled {
		t.Errorf("expected DeconcileSession to be called upon session deletion")
	}
	if mockReconciler.deconcileSessionID != 777 {
		t.Errorf("expected DeconcileSession to be called with ID 777, got %d", mockReconciler.deconcileSessionID)
	}
}

func strPtr(s string) *string {
	return &s
}
