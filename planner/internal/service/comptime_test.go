package service

import (
	"context"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

type mockComptimeRepo struct {
	comptimes       map[int64]*domain.Comptime
	sessions        map[int64]*domain.Session
	invoiceItems    map[int64]*domain.InvoiceItem
	comptimeSessions map[int64]*domain.ComptimeSession
	comptimeItems   map[int64]*domain.ComptimeInvoiceItem
	comptimeExtras  map[int64]*domain.ComptimeExtra
	idCounter       int64
}

func newMockComptimeRepo() *mockComptimeRepo {
	return &mockComptimeRepo{
		comptimes:       make(map[int64]*domain.Comptime),
		sessions:        make(map[int64]*domain.Session),
		invoiceItems:    make(map[int64]*domain.InvoiceItem),
		comptimeSessions: make(map[int64]*domain.ComptimeSession),
		comptimeItems:   make(map[int64]*domain.ComptimeInvoiceItem),
		comptimeExtras:  make(map[int64]*domain.ComptimeExtra),
		idCounter:       1,
	}
}

func (m *mockComptimeRepo) LockCustomerService(ctx context.Context, customerID, serviceID int64) error {
	return nil
}

func (m *mockComptimeRepo) FindOrCreateComptime(ctx context.Context, vendorID, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	norm := domain.NormalizeMonth(month)
	for _, c := range m.comptimes {
		if c.CustomerID == customerID && c.ServiceID == serviceID && c.Month.Equal(norm) && c.DeletedAt == nil {
			return c, nil
		}
	}
	// Initial balance from previous month
	prev, _ := m.GetPreviousComptime(ctx, customerID, serviceID, norm)
	prevBal := 0
	if prev != nil {
		prevBal = prev.Balance
	}
	m.idCounter++
	c := domain.NewComptime(vendorID, customerID, serviceID, norm)
	c.ID = m.idCounter
	c.Balance = prevBal
	m.comptimes[c.ID] = c
	return c, nil
}

func (m *mockComptimeRepo) GetComptime(ctx context.Context, id int64) (*domain.Comptime, error) {
	return m.comptimes[id], nil
}

func (m *mockComptimeRepo) GetComptimeByCustomerServiceMonth(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	norm := domain.NormalizeMonth(month)
	for _, c := range m.comptimes {
		if c.CustomerID == customerID && c.ServiceID == serviceID && c.Month.Equal(norm) && c.DeletedAt == nil {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockComptimeRepo) GetPreviousComptime(ctx context.Context, customerID, serviceID int64, month time.Time) (*domain.Comptime, error) {
	norm := domain.NormalizeMonth(month)
	var latest *domain.Comptime
	for _, c := range m.comptimes {
		if c.CustomerID == customerID && c.ServiceID == serviceID && c.Month.Before(norm) && c.DeletedAt == nil {
			if latest == nil || c.Month.After(latest.Month) {
				latest = c
			}
		}
	}
	return latest, nil
}

func (m *mockComptimeRepo) GetSubsequentComptimes(ctx context.Context, customerID, serviceID int64, month time.Time) ([]domain.Comptime, error) {
	norm := domain.NormalizeMonth(month)
	var list []domain.Comptime
	for _, c := range m.comptimes {
		if c.CustomerID == customerID && c.ServiceID == serviceID && c.Month.After(norm) && c.DeletedAt == nil {
			list = append(list, *c)
		}
	}
	// Sort by month asc
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Month.Before(list[i].Month) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) SaveComptime(ctx context.Context, comptime *domain.Comptime) error {
	m.comptimes[comptime.ID] = comptime
	return nil
}

func (m *mockComptimeRepo) ListComptime(ctx context.Context, filter port.ComptimeFilter) ([]port.ComptimeView, int64, error) {
	var views []port.ComptimeView
	for _, c := range m.comptimes {
		views = append(views, port.ComptimeView{
			ID:              c.ID,
			VendorID:        c.VendorID,
			CustomerID:      c.CustomerID,
			ServiceID:       c.ServiceID,
			Month:           c.Month,
			InvoicedMinutes: c.InvoicedMinutes,
			ConsumedMinutes: c.ConsumedMinutes,
			Balance:         c.Balance,
		})
	}
	return views, int64(len(views)), nil
}

func (m *mockComptimeRepo) FindComptimeSessionBySessionID(ctx context.Context, sessionID int64) (*domain.ComptimeSession, error) {
	for _, cs := range m.comptimeSessions {
		if cs.SessionID == sessionID && cs.DeletedAt == nil {
			return cs, nil
		}
	}
	return nil, nil
}

func (m *mockComptimeRepo) SaveComptimeSession(ctx context.Context, cs *domain.ComptimeSession) error {
	if cs.ID == 0 {
		m.idCounter++
		cs.ID = m.idCounter
	}
	m.comptimeSessions[cs.ID] = cs
	return nil
}

func (m *mockComptimeRepo) DeleteComptimeSession(ctx context.Context, sessionID int64) error {
	for _, cs := range m.comptimeSessions {
		if cs.SessionID == sessionID && cs.DeletedAt == nil {
			now := time.Now()
			cs.DeletedAt = &now
		}
	}
	return nil
}

func (m *mockComptimeRepo) FindComptimeInvoiceItemByInvoiceItemID(ctx context.Context, invoiceItemID int64) (*domain.ComptimeInvoiceItem, error) {
	for _, cii := range m.comptimeItems {
		if cii.InvoiceItemID == invoiceItemID && cii.DeletedAt == nil {
			return cii, nil
		}
	}
	return nil, nil
}

func (m *mockComptimeRepo) FindComptimeInvoiceItemsByInvoiceID(ctx context.Context, invoiceID int64) ([]domain.ComptimeInvoiceItem, error) {
	var list []domain.ComptimeInvoiceItem
	for _, cii := range m.comptimeItems {
		if cii.DeletedAt == nil {
			list = append(list, *cii)
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) SaveComptimeInvoiceItem(ctx context.Context, cii *domain.ComptimeInvoiceItem) error {
	if cii.ID == 0 {
		m.idCounter++
		cii.ID = m.idCounter
	}
	m.comptimeItems[cii.ID] = cii
	return nil
}

func (m *mockComptimeRepo) DeleteComptimeInvoiceItem(ctx context.Context, invoiceItemID int64) error {
	for _, cii := range m.comptimeItems {
		if cii.InvoiceItemID == invoiceItemID && cii.DeletedAt == nil {
			now := time.Now()
			cii.DeletedAt = &now
		}
	}
	return nil
}

func (m *mockComptimeRepo) SaveComptimeExtra(ctx context.Context, ce *domain.ComptimeExtra) error {
	if ce.ID == 0 {
		m.idCounter++
		ce.ID = m.idCounter
	}
	m.comptimeExtras[ce.ID] = ce
	return nil
}

func (m *mockComptimeRepo) DeleteComptimeExtra(ctx context.Context, id int64) error {
	if ce, ok := m.comptimeExtras[id]; ok {
		now := time.Now()
		ce.DeletedAt = &now
	}
	return nil
}

func (m *mockComptimeRepo) FindActiveSessionsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time) ([]domain.Session, error) {
	var list []domain.Session
	for _, s := range m.sessions {
		if s.DeletedAt == nil {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) FindActiveInvoiceItemsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time) ([]domain.InvoiceItem, error) {
	var list []domain.InvoiceItem
	for _, it := range m.invoiceItems {
		if it.DeletedAt == nil {
			list = append(list, *it)
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) GetSessionByID(id int64) (*domain.Session, error) {
	return m.sessions[id], nil
}

func (m *mockComptimeRepo) GetServiceByID(id int64) (*domain.Service, error) {
	min := 50
	return &domain.Service{ID: id, SessionMinutes: &min}, nil
}

func (m *mockComptimeRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return nil, nil
}

func (m *mockComptimeRepo) GetInvoiceItemByID(ctx context.Context, id int64) (*domain.InvoiceItem, error) {
	return m.invoiceItems[id], nil
}

// Tests

func TestComptimeService_SessionLifecycleAndBalance(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	// Seed session
	sessDate := time.Date(2026, time.March, 10, 10, 0, 0, 0, time.UTC)
	repo.sessions[10] = &domain.Session{
		ID:             10,
		CustomerID:     100,
		ServiceID:      200,
		SessionDate:    sessDate,
		SessionMinutes: 60,
		SessionStatus:  string(domain.StatusDone),
	}

	// 1. ProcessSessionCreated
	if err := svc.ProcessSessionCreated(ctx, 10); err != nil {
		t.Fatalf("ProcessSessionCreated failed: %v", err)
	}

	normMonth := domain.NormalizeMonth(sessDate)
	comp, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 100, 200, normMonth)
	if comp == nil {
		t.Fatalf("expected comptime created")
	}
	if comp.ConsumedMinutes != 60 || comp.Balance != -60 {
		t.Errorf("expected consumed 60 and balance -60, got consumed=%d balance=%d",
			comp.ConsumedMinutes, comp.Balance)
	}

	// 2. Idempotency test
	if err := svc.ProcessSessionCreated(ctx, 10); err != nil {
		t.Fatalf("ProcessSessionCreated second call failed: %v", err)
	}
	if comp.ConsumedMinutes != 60 {
		t.Errorf("expected still 60 minutes after second call, got %d", comp.ConsumedMinutes)
	}

	// 3. ProcessSessionUpdated
	repo.sessions[10].SessionMinutes = 90
	if err := svc.ProcessSessionUpdated(ctx, 10); err != nil {
		t.Fatalf("ProcessSessionUpdated failed: %v", err)
	}
	if comp.ConsumedMinutes != 90 || comp.Balance != -90 {
		t.Errorf("expected consumed 90 and balance -90, got consumed=%d balance=%d",
			comp.ConsumedMinutes, comp.Balance)
	}

	// 4. Cancelled without charge should count as 0 minutes
	repo.sessions[10].SessionStatus = string(domain.StatusCancelNotCharge)
	if err := svc.ProcessSessionUpdated(ctx, 10); err != nil {
		t.Fatalf("ProcessSessionUpdated cancel failed: %v", err)
	}
	if comp.ConsumedMinutes != 0 || comp.Balance != 0 {
		t.Errorf("expected consumed 0 and balance 0, got consumed=%d balance=%d",
			comp.ConsumedMinutes, comp.Balance)
	}

	// 5. ProcessSessionDeleted
	if err := svc.ProcessSessionDeleted(ctx, 10); err != nil {
		t.Fatalf("ProcessSessionDeleted failed: %v", err)
	}
	cs, _ := repo.FindComptimeSessionBySessionID(ctx, 10)
	if cs != nil {
		t.Errorf("expected comptime_session to be deleted")
	}
}

func TestComptimeService_CascadeBalancePropagation(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	// Month 1: March 2026
	m1 := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	// Month 2: April 2026
	m2 := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)

	// Create March comptime with 120 invoiced and 40 consumed => balance 80
	cMarch, _ := repo.FindOrCreateComptime(ctx, 1, 100, 200, m1)
	cMarch.InvoicedMinutes = 120
	cMarch.ConsumedMinutes = 40
	cMarch.RecalculateBalance(0)
	_ = repo.SaveComptime(ctx, cMarch)

	// Create April comptime with 60 invoiced and 30 consumed => initial balance should inherit 80, balance = 80 + 60 - 30 = 110
	cApril, _ := repo.FindOrCreateComptime(ctx, 1, 100, 200, m2)
	cApril.InvoicedMinutes = 60
	cApril.ConsumedMinutes = 30
	cApril.RecalculateBalance(cMarch.Balance)
	_ = repo.SaveComptime(ctx, cApril)

	if cApril.Balance != 110 {
		t.Fatalf("expected April balance 110, got %d", cApril.Balance)
	}

	// Now add a session to March with 20 minutes
	// March balance should become 80 - 20 = 60
	// April balance should automatically propagate to 60 + 60 - 30 = 90!
	repo.sessions[55] = &domain.Session{
		ID:             55,
		CustomerID:     100,
		ServiceID:      200,
		SessionDate:    time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC),
		SessionMinutes: 20,
		SessionStatus:  string(domain.StatusDone),
	}

	if err := svc.ProcessSessionCreated(ctx, 55); err != nil {
		t.Fatalf("ProcessSessionCreated failed: %v", err)
	}

	cMarchUpdated, _ := repo.GetComptime(ctx, cMarch.ID)
	if cMarchUpdated.Balance != 60 {
		t.Errorf("expected March balance 60, got %d", cMarchUpdated.Balance)
	}

	cAprilUpdated, _ := repo.GetComptime(ctx, cApril.ID)
	if cAprilUpdated.Balance != 90 {
		t.Errorf("expected April balance propagated to 90, got %d", cAprilUpdated.Balance)
	}
}

func TestComptimeService_InvoiceItemCredit(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	minutes := 50
	service := &domain.Service{ID: 200, SessionMinutes: &minutes}
	periodStart := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	invoice := &domain.Invoice{
		ID:          1,
		PeriodStart: periodStart,
		Plan: &domain.Plan{
			CustomerID: 100,
			Customer:   &domain.Customer{VendorID: 1},
		},
	}
	repo.invoiceItems[501] = &domain.InvoiceItem{
		ID:        501,
		InvoiceID: 1,
		Invoice:   invoice,
		ServiceID: 200,
		Service:   service,
		Quantity:  4, // 4 * 50 = 200 invoiced minutes
	}

	if err := svc.ProcessInvoiceItemCreated(ctx, 501); err != nil {
		t.Fatalf("ProcessInvoiceItemCreated failed: %v", err)
	}

	comp, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 100, 200, periodStart)
	if comp == nil {
		t.Fatalf("expected comptime created for May")
	}
	if comp.InvoicedMinutes != 200 || comp.Balance != 200 {
		t.Errorf("expected 200 invoiced minutes and balance 200, got invoiced=%d balance=%d",
			comp.InvoicedMinutes, comp.Balance)
	}

	// Update invoice item quantity to 3 (150 minutes)
	repo.invoiceItems[501].Quantity = 3
	if err := svc.ProcessInvoiceItemUpdated(ctx, 501); err != nil {
		t.Fatalf("ProcessInvoiceItemUpdated failed: %v", err)
	}
	compUpdated, _ := repo.GetComptime(ctx, comp.ID)
	if compUpdated.InvoicedMinutes != 150 || compUpdated.Balance != 150 {
		t.Errorf("expected 150 invoiced minutes, got %d", compUpdated.InvoicedMinutes)
	}

	// Delete invoice item
	if err := svc.ProcessInvoiceItemDeleted(ctx, 501); err != nil {
		t.Fatalf("ProcessInvoiceItemDeleted failed: %v", err)
	}
	compAfterDelete, _ := repo.GetComptime(ctx, comp.ID)
	if compAfterDelete.InvoicedMinutes != 0 || compAfterDelete.Balance != 0 {
		t.Errorf("expected 0 invoiced minutes after deletion, got %d", compAfterDelete.InvoicedMinutes)
	}
}

func TestComptimeService_Resync(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	minutes := 60
	service := &domain.Service{ID: 200, SessionMinutes: &minutes}
	periodStart := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)
	invoice := &domain.Invoice{
		ID:          1,
		PeriodStart: periodStart,
		Plan: &domain.Plan{
			CustomerID: 100,
			Customer:   &domain.Customer{VendorID: 1},
		},
	}
	repo.invoiceItems[10] = &domain.InvoiceItem{
		ID:        10,
		InvoiceID: 1,
		Invoice:   invoice,
		ServiceID: 200,
		Service:   service,
		Quantity:  2, // 2 * 60 = 120 min
	}

	repo.sessions[20] = &domain.Session{
		ID:             20,
		CustomerID:     100,
		ServiceID:      200,
		SessionDate:    time.Date(2026, time.June, 5, 0, 0, 0, 0, time.UTC),
		SessionMinutes: 60,
		SessionStatus:  string(domain.StatusDone),
	}

	res, err := svc.Resync(ctx, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "all")
	if err != nil {
		t.Fatalf("Resync failed: %v", err)
	}
	if res.SessionsSynced != 1 || res.InvoicesSynced != 1 {
		t.Errorf("expected 1 session and 1 invoice synced, got sessions=%d invoices=%d",
			res.SessionsSynced, res.InvoicesSynced)
	}

	comp, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 100, 200, periodStart)
	if comp == nil {
		t.Fatalf("expected comptime created after resync")
	}
	if comp.InvoicedMinutes != 120 || comp.ConsumedMinutes != 60 || comp.Balance != 60 {
		t.Errorf("expected invoiced=120, consumed=60, balance=60; got invoiced=%d, consumed=%d, balance=%d",
			comp.InvoicedMinutes, comp.ConsumedMinutes, comp.Balance)
	}
}

