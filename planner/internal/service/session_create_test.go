package service

import (
	"context"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

type mockComptimeRecorder struct {
	sessionCreatedCalls     []int64
	invoiceItemCreatedCalls []int64
}

func (m *mockComptimeRecorder) ProcessSessionCreated(ctx context.Context, sessionID int64) error {
	m.sessionCreatedCalls = append(m.sessionCreatedCalls, sessionID)
	return nil
}
func (m *mockComptimeRecorder) ProcessSessionUpdated(ctx context.Context, sessionID int64) error { return nil }
func (m *mockComptimeRecorder) ProcessSessionDeleted(ctx context.Context, sessionID int64) error { return nil }
func (m *mockComptimeRecorder) ProcessInvoiceItemCreated(ctx context.Context, invoiceItemID int64) error {
	m.invoiceItemCreatedCalls = append(m.invoiceItemCreatedCalls, invoiceItemID)
	return nil
}
func (m *mockComptimeRecorder) ProcessInvoiceItemUpdated(ctx context.Context, invoiceItemID int64) error { return nil }
func (m *mockComptimeRecorder) ProcessInvoiceItemDeleted(ctx context.Context, invoiceItemID int64) error { return nil }
func (m *mockComptimeRecorder) ProcessInvoiceDeleted(ctx context.Context, invoiceID int64) error     { return nil }
func (m *mockComptimeRecorder) ProcessComptimeExtraCreated(ctx context.Context, extra *domain.ComptimeExtra) error { return nil }
func (m *mockComptimeRecorder) ProcessComptimeExtraUpdated(ctx context.Context, extra *domain.ComptimeExtra) error { return nil }
func (m *mockComptimeRecorder) ProcessComptimeExtraDeleted(ctx context.Context, extraID int64) error { return nil }
func (m *mockComptimeRecorder) GetComptime(ctx context.Context, id int64) (*domain.Comptime, error) { return nil, nil }
func (m *mockComptimeRecorder) GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) { return nil, nil }
func (m *mockComptimeRecorder) ListComptime(ctx context.Context, filter port.ComptimeFilter) ([]port.ComptimeView, int64, error) { return nil, 0, nil }
func (m *mockComptimeRecorder) DeleteComptime(ctx context.Context, id int64) error { return nil }
func (m *mockComptimeRecorder) Resync(ctx context.Context, vendorID int64, startDate, endDate time.Time, dataType string, customerID, serviceID *int64) (*port.ResyncResult, error) { return nil, nil }

type mockSessionCreateRepo struct {
	mockPlanRepoForService
	customers    map[string]*domain.Customer
	savedSession *domain.Session
	savedInvoice *domain.Invoice
	plansForInv  []domain.Plan
	nextSessID   int64
	nextInvID    int64
}

func newMockSessionCreateRepo() *mockSessionCreateRepo {
	base := newMockPlanRepoForService()
	return &mockSessionCreateRepo{
		mockPlanRepoForService: *base,
		customers:             make(map[string]*domain.Customer),
		nextSessID:            101,
		nextInvID:             201,
	}
}

func (m *mockSessionCreateRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if c, ok := m.customers[nickname]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockSessionCreateRepo) GetServiceByID(id int64) (*domain.Service, error) {
	if s, ok := m.services[id]; ok {
		return s, nil
	}
	return &domain.Service{ID: id, Name: "service_default"}, nil
}

func (m *mockSessionCreateRepo) Save(model interface{}) error {
	if sess, ok := model.(*domain.Session); ok {
		if sess.ID == 0 {
			sess.ID = m.nextSessID
			m.nextSessID++
		}
		m.savedSession = sess
	}
	return nil
}

func (m *mockSessionCreateRepo) FindActivePlansForInvoicing(planID *int64, vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	return m.plansForInv, nil
}

func (m *mockSessionCreateRepo) SaveInvoice(invoice *domain.Invoice) error {
	if invoice.ID == 0 {
		invoice.ID = m.nextInvID
		m.nextInvID++
	}
	for i := range invoice.Items {
		if invoice.Items[i].ID == 0 {
			invoice.Items[i].ID = int64(300 + i)
			invoice.Items[i].InvoiceID = invoice.ID
		}
	}
	m.savedInvoice = invoice
	return nil
}

func floatPtrHelper(f float64) *float64 {
	return &f
}

func TestSessionCreate_PayPerView_AutomaticInvoicingAndComptime(t *testing.T) {
	repo := newMockSessionCreateRepo()
	logger := &mockLogger{}
	comptimeRec := &mockComptimeRecorder{}

	repo.customers["cliente_ppv"] = &domain.Customer{ID: 10, Nickname: "cliente_ppv", VendorID: 1}
	desc := "Aula de Canto Avulsa"
	repo.services[2] = &domain.Service{ID: 2, Name: "online/canto/45", Description: &desc}

	planEnd := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	repo.plansForInv = []domain.Plan{
		{
			ID:         50,
			CustomerID: 10,
			PlanType:   int(domain.PlanTypePayPerView),
			PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PlanEnd:    &planEnd,
			PayPerView: &domain.PlanPayPerView{
				ID:          50,
				PaymentTerm: 15,
			},
			Items: []domain.PlanItem{
				{
					ID:         1,
					PlanID:     50,
					ServiceID:  2,
					OrderIndex: 1,
					Price:      floatPtrHelper(150.0),
					Service:    repo.services[2],
				},
			},
		},
	}

	svc := NewSessionCreate(repo, logger)
	svc.SetComptimeService(comptimeRec)

	req := &dto.SessionCreateRequest{
		Nickname:  "cliente_ppv",
		Date:      "2026-10-15",
		Minutes:   45,
		ServiceID: 2,
		Status:    "realizada",
		Comments:  "Primeira sessão avulsa",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.savedInvoice == nil {
		t.Fatal("expected an automatic pre-invoice to be created, but none was saved")
	}

	inv := repo.savedInvoice
	if inv.PlanID != 50 {
		t.Errorf("expected PlanID 50, got %d", inv.PlanID)
	}
	expectedDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if !inv.InvoicingDate.Equal(expectedDate) {
		t.Errorf("expected InvoicingDate %v, got %v", expectedDate, inv.InvoicingDate)
	}
	expectedDueDate := time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC) // 15 + 15
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}
	if !inv.PeriodStart.Equal(expectedDate) {
		t.Errorf("expected PeriodStart %v, got %v", expectedDate, inv.PeriodStart)
	}
	if !inv.PeriodEnd.Equal(expectedDate) {
		t.Errorf("expected PeriodEnd %v, got %v", expectedDate, inv.PeriodEnd)
	}
	if inv.Amount != 150.0 {
		t.Errorf("expected Amount 150.0, got %.2f", inv.Amount)
	}
	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 invoice item, got %d", len(inv.Items))
	}
	item := inv.Items[0]
	if item.ServiceID != 2 {
		t.Errorf("expected item ServiceID 2, got %d", item.ServiceID)
	}
	if item.Price != 150.0 {
		t.Errorf("expected item Price 150.0, got %.2f", item.Price)
	}
	if item.Quantity != 1 {
		t.Errorf("expected item Quantity 1, got %d", item.Quantity)
	}
	expectedDesc := "Aula de Canto Avulsa - 15/10/2026"
	if item.Description == nil || *item.Description != expectedDesc {
		t.Errorf("expected item Description %q, got %v", expectedDesc, item.Description)
	}

	// Comptime assertions
	if len(comptimeRec.invoiceItemCreatedCalls) != 1 || comptimeRec.invoiceItemCreatedCalls[0] != 300 {
		t.Errorf("expected ProcessInvoiceItemCreated(300), got %v", comptimeRec.invoiceItemCreatedCalls)
	}
	if len(comptimeRec.sessionCreatedCalls) != 1 || comptimeRec.sessionCreatedCalls[0] != 101 {
		t.Errorf("expected ProcessSessionCreated(101), got %v", comptimeRec.sessionCreatedCalls)
	}
}

func TestSessionCreate_PayPerView_PlanPriceOverridesItemPrice(t *testing.T) {
	repo := newMockSessionCreateRepo()
	logger := &mockLogger{}
	comptimeRec := &mockComptimeRecorder{}

	repo.customers["cliente_ppv"] = &domain.Customer{ID: 10, Nickname: "cliente_ppv", VendorID: 1}
	desc := "Mentoria Musical"
	repo.services[3] = &domain.Service{ID: 3, Name: "mentoria", Description: &desc}

	repo.plansForInv = []domain.Plan{
		{
			ID:         51,
			CustomerID: 10,
			PlanType:   int(domain.PlanTypePayPerView),
			Price:      floatPtrHelper(250.0), // valor do pacote/plano
			PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PayPerView: &domain.PlanPayPerView{ID: 51, PaymentTerm: 10},
			Items: []domain.PlanItem{
				{
					ID:         2,
					PlanID:     51,
					ServiceID:  3,
					OrderIndex: 1,
					Price:      floatPtrHelper(180.0), // valor do item, sobreposto por Price do plano
					Service:    repo.services[3],
				},
			},
		},
	}

	svc := NewSessionCreate(repo, logger)
	svc.SetComptimeService(comptimeRec)

	req := &dto.SessionCreateRequest{
		Nickname:  "cliente_ppv",
		Date:      "2026-10-20",
		Minutes:   60,
		ServiceID: 3,
		Status:    "realizada",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.savedInvoice == nil {
		t.Fatal("expected pre-invoice to be saved")
	}
	if repo.savedInvoice.Amount != 250.0 {
		t.Errorf("expected Amount 250.0, got %.2f", repo.savedInvoice.Amount)
	}
	if repo.savedInvoice.Items[0].Price != 250.0 {
		t.Errorf("expected Item Price 250.0, got %.2f", repo.savedInvoice.Items[0].Price)
	}
}

func TestSessionCreate_PayPerView_OutsidePlanPeriod_NoInvoice(t *testing.T) {
	repo := newMockSessionCreateRepo()
	logger := &mockLogger{}
	comptimeRec := &mockComptimeRecorder{}

	repo.customers["cliente_ppv"] = &domain.Customer{ID: 10, Nickname: "cliente_ppv", VendorID: 1}
	repo.services[2] = &domain.Service{ID: 2, Name: "canto"}

	planEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	repo.plansForInv = []domain.Plan{
		{
			ID:         52,
			CustomerID: 10,
			PlanType:   int(domain.PlanTypePayPerView),
			PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PlanEnd:    &planEnd,
			PayPerView: &domain.PlanPayPerView{ID: 52, PaymentTerm: 5},
			Items: []domain.PlanItem{
				{ID: 3, PlanID: 52, ServiceID: 2, Price: floatPtrHelper(100.0)},
			},
		},
	}

	svc := NewSessionCreate(repo, logger)
	svc.SetComptimeService(comptimeRec)

	// Sessão em outubro, após o fim do plano (setembro)
	req := &dto.SessionCreateRequest{
		Nickname:  "cliente_ppv",
		Date:      "2026-10-05",
		Minutes:   30,
		ServiceID: 2,
		Status:    "realizada",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.savedInvoice != nil {
		t.Errorf("expected no invoice to be created for session outside plan period, got %v", repo.savedInvoice)
	}
	if len(comptimeRec.invoiceItemCreatedCalls) != 0 {
		t.Errorf("expected 0 invoice item comptime calls, got %v", comptimeRec.invoiceItemCreatedCalls)
	}
	if len(comptimeRec.sessionCreatedCalls) != 1 {
		t.Errorf("expected 1 session comptime call, got %v", comptimeRec.sessionCreatedCalls)
	}
}

func TestSessionCreate_PayPerView_CancelledNotCharge_NoInvoice(t *testing.T) {
	repo := newMockSessionCreateRepo()
	logger := &mockLogger{}
	comptimeRec := &mockComptimeRecorder{}

	repo.customers["cliente_ppv"] = &domain.Customer{ID: 10, Nickname: "cliente_ppv", VendorID: 1}
	repo.services[2] = &domain.Service{ID: 2, Name: "canto"}

	repo.plansForInv = []domain.Plan{
		{
			ID:         53,
			CustomerID: 10,
			PlanType:   int(domain.PlanTypePayPerView),
			PlanStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			PayPerView: &domain.PlanPayPerView{ID: 53, PaymentTerm: 5},
			Items: []domain.PlanItem{
				{ID: 4, PlanID: 53, ServiceID: 2, Price: floatPtrHelper(100.0)},
			},
		},
	}

	svc := NewSessionCreate(repo, logger)
	svc.SetComptimeService(comptimeRec)

	req := &dto.SessionCreateRequest{
		Nickname:  "cliente_ppv",
		Date:      "2026-10-05",
		Minutes:   30,
		ServiceID: 2,
		Status:    "cancelada_nao_cobrar",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.savedInvoice != nil {
		t.Errorf("expected no invoice for cancelada_nao_cobrar, got %v", repo.savedInvoice)
	}
}

func TestSessionCreate_NoPayPerViewPlan_NoInvoice(t *testing.T) {
	repo := newMockSessionCreateRepo()
	logger := &mockLogger{}
	comptimeRec := &mockComptimeRecorder{}

	repo.customers["cliente_comum"] = &domain.Customer{ID: 20, Nickname: "cliente_comum", VendorID: 1}
	repo.services[1] = &domain.Service{ID: 1, Name: "piano"}
	repo.plansForInv = nil // sem planos

	svc := NewSessionCreate(repo, logger)
	svc.SetComptimeService(comptimeRec)

	req := &dto.SessionCreateRequest{
		Nickname:  "cliente_comum",
		Date:      "2026-10-10",
		Minutes:   30,
		ServiceID: 1,
		Status:    "realizada",
	}

	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", resp.GetStatusCode(), resp.GetMessage())
	}

	if repo.savedInvoice != nil {
		t.Errorf("expected no invoice when customer has no plan, got %v", repo.savedInvoice)
	}
	if len(comptimeRec.sessionCreatedCalls) != 1 {
		t.Errorf("expected 1 session comptime call, got %v", comptimeRec.sessionCreatedCalls)
	}
}
