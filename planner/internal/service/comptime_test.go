package service

import (
	"context"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

type mockComptimeRepo struct {
	comptimes        map[int64]*domain.Comptime
	sessions         map[int64]*domain.Session
	invoiceItems     map[int64]*domain.InvoiceItem
	comptimeSessions map[int64]*domain.ComptimeSession
	comptimeItems    map[int64]*domain.ComptimeInvoiceItem
	comptimeExtras   map[int64]*domain.ComptimeExtra
	plans            []*domain.Plan
	services         map[int64]*domain.Service
	idCounter        int64
}

func newMockComptimeRepo() *mockComptimeRepo {
	return &mockComptimeRepo{
		comptimes:        make(map[int64]*domain.Comptime),
		sessions:         make(map[int64]*domain.Session),
		invoiceItems:     make(map[int64]*domain.InvoiceItem),
		comptimeSessions: make(map[int64]*domain.ComptimeSession),
		comptimeItems:    make(map[int64]*domain.ComptimeInvoiceItem),
		comptimeExtras:   make(map[int64]*domain.ComptimeExtra),
		services:         make(map[int64]*domain.Service),
		idCounter:        1,
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

func (m *mockComptimeRepo) DeleteComptime(ctx context.Context, id int64) error {
	if c, ok := m.comptimes[id]; ok {
		now := time.Now()
		c.DeletedAt = &now
		for _, cs := range m.comptimeSessions {
			if cs.ComptimeID == id {
				cs.DeletedAt = &now
			}
		}
		for _, cii := range m.comptimeItems {
			if cii.ComptimeID == id {
				cii.DeletedAt = &now
			}
		}
		for _, ce := range m.comptimeExtras {
			if ce.ComptimeID == id {
				ce.DeletedAt = &now
			}
		}
	}
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
			MonthBalance:    c.MonthBalance(),
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

func (m *mockComptimeRepo) FindActiveSessionsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.Session, error) {
	var list []domain.Session
	for _, s := range m.sessions {
		if s.DeletedAt == nil {
			if customerID != nil && *customerID > 0 && s.CustomerID != *customerID {
				continue
			}
			if serviceID != nil && *serviceID > 0 && s.ServiceID != *serviceID {
				continue
			}
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) FindActiveInvoiceItemsForComptime(ctx context.Context, vendorID int64, startDate, endDate time.Time, customerID, serviceID *int64) ([]domain.InvoiceItem, error) {
	var list []domain.InvoiceItem
	for _, it := range m.invoiceItems {
		if it.DeletedAt == nil {
			if customerID != nil && *customerID > 0 {
				if it.Invoice == nil || it.Invoice.Plan == nil || it.Invoice.Plan.CustomerID != *customerID {
					continue
				}
			}
			if serviceID != nil && *serviceID > 0 && it.ServiceID != *serviceID {
				continue
			}
			list = append(list, *it)
		}
	}
	return list, nil
}

func (m *mockComptimeRepo) GetSessionByID(id int64) (*domain.Session, error) {
	return m.sessions[id], nil
}

func (m *mockComptimeRepo) GetServiceByID(id int64) (*domain.Service, error) {
	if svc, ok := m.services[id]; ok {
		return svc, nil
	}
	min := 50
	return &domain.Service{ID: id, SessionMinutes: &min}, nil
}

func (m *mockComptimeRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return nil, nil
}

func (m *mockComptimeRepo) GetInvoiceItemByID(ctx context.Context, id int64) (*domain.InvoiceItem, error) {
	return m.invoiceItems[id], nil
}

func (m *mockComptimeRepo) FindActivePlanForCustomerService(ctx context.Context, customerID, serviceID int64, refDate time.Time) (*domain.Plan, error) {
	for _, p := range m.plans {
		if p != nil && p.CustomerID == customerID && p.IsActiveAt(refDate) && p.ContainsService(serviceID) {
			return p, nil
		}
	}
	return nil, nil
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

	// 1. Resync without filters (assumes all)
	res, err := svc.Resync(ctx, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "all", nil, nil)
	if err != nil {
		t.Fatalf("Resync failed: %v", err)
	}
	if res.SessionsSynced != 1 || res.InvoicesSynced != 1 {
		t.Errorf("expected 1 session and 1 invoice synced, got sessions=%d invoices=%d",
			res.SessionsSynced, res.InvoicesSynced)
	}
	if res.ComptimeUpdated != 1 {
		t.Errorf("expected 1 comptime updated, got %d", res.ComptimeUpdated)
	}

	comp, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 100, 200, periodStart)
	if comp == nil {
		t.Fatalf("expected comptime created after resync")
	}
	if comp.InvoicedMinutes != 120 || comp.ConsumedMinutes != 60 || comp.Balance != 60 {
		t.Errorf("expected invoiced=120, consumed=60, balance=60; got invoiced=%d, consumed=%d, balance=%d",
			comp.InvoicedMinutes, comp.ConsumedMinutes, comp.Balance)
	}

	// 2. Resync filtering for a different customer (should sync 0)
	otherCustID := int64(999)
	resOther, err := svc.Resync(ctx, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "all", &otherCustID, nil)
	if err != nil {
		t.Fatalf("Resync with other customer failed: %v", err)
	}
	if resOther.SessionsSynced != 0 || resOther.InvoicesSynced != 0 {
		t.Errorf("expected 0 synced for other customer, got sessions=%d invoices=%d",
			resOther.SessionsSynced, resOther.InvoicesSynced)
	}

	// 3. Resync filtering for matching customer and service
	targetCustID := int64(100)
	targetServiceID := int64(200)
	resFilter, err := svc.Resync(ctx, 1, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), "all", &targetCustID, &targetServiceID)
	if err != nil {
		t.Fatalf("Resync with filter failed: %v", err)
	}
	if resFilter.CustomerID == nil || *resFilter.CustomerID != 100 || resFilter.ServiceID == nil || *resFilter.ServiceID != 200 {
		t.Errorf("expected result to reflect customer and service filters")
	}
}

func TestComptimeService_DeleteComptime(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	m1 := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	comp := &domain.Comptime{
		ID:              50,
		VendorID:        1,
		CustomerID:      100,
		ServiceID:       200,
		Month:           m1,
		InvoicedMinutes: 100,
		ConsumedMinutes: 40,
		Balance:         60,
	}
	repo.comptimes[50] = comp
	repo.comptimeSessions[1] = &domain.ComptimeSession{ID: 1, ComptimeID: 50, SessionID: 10}
	repo.comptimeItems[1] = &domain.ComptimeInvoiceItem{ID: 1, ComptimeID: 50, InvoiceItemID: 20}

	// 1. Delete existing comptime
	if err := svc.DeleteComptime(ctx, 50); err != nil {
		t.Fatalf("DeleteComptime failed: %v", err)
	}

	if comp.DeletedAt == nil {
		t.Errorf("expected comptime to be soft-deleted")
	}
	if repo.comptimeSessions[1].DeletedAt == nil {
		t.Errorf("expected comptimeSession to be soft-deleted")
	}
	if repo.comptimeItems[1].DeletedAt == nil {
		t.Errorf("expected comptimeItem to be soft-deleted")
	}

	// 2. Delete non-existent comptime returns error
	if err := svc.DeleteComptime(ctx, 9999); err == nil {
		t.Errorf("expected error deleting non-existent comptime")
	}
}

func TestComptimeService_MonthBalanceAndQueries(t *testing.T) {
	repo := newMockComptimeRepo()
	svc := NewComptime(repo, nil)
	ctx := context.Background()

	m := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	comp := &domain.Comptime{
		ID:              1,
		VendorID:        1,
		CustomerID:      10,
		ServiceID:       20,
		Month:           m,
		InvoicedMinutes: 200,
		ConsumedMinutes: 150,
		Balance:         50,
	}
	repo.comptimes[1] = comp

	// Check domain MonthBalance method
	if comp.MonthBalance() != 50 {
		t.Errorf("expected MonthBalance 50, got %d", comp.MonthBalance())
	}

	// Service GetComptime
	got, err := svc.GetComptime(ctx, 1)
	if err != nil || got == nil || got.ID != 1 {
		t.Fatalf("expected GetComptime(1) to return record, got %v, err %v", got, err)
	}
	if got.MonthBalance() != 50 {
		t.Errorf("expected got.MonthBalance 50, got %d", got.MonthBalance())
	}

	// Service ListComptime
	views, total, err := svc.ListComptime(ctx, port.ComptimeFilter{})
	if err != nil || total != 1 || len(views) != 1 {
		t.Fatalf("expected ListComptime to return 1 view, got %d / total %d", len(views), total)
	}
	if views[0].MonthBalance != 50 {
		t.Errorf("expected views[0].MonthBalance = 50, got %d", views[0].MonthBalance)
	}
	if views[0].Balance != 50 {
		t.Errorf("expected views[0].Balance = 50, got %d", views[0].Balance)
	}
	if views[0].PredictMinutes != 200 || views[0].DoneMinutes != 150 {
		t.Errorf("expected PredictMinutes=200 DoneMinutes=150 for past month, got %d / %d", views[0].PredictMinutes, views[0].DoneMinutes)
	}
	if views[0].PredictSessions != 4.0 || views[0].DoneSessions != 3.0 || views[0].MonthBalanceSessions != 1.0 || views[0].BalanceSessions != 1.0 {
		t.Errorf("unexpected session counts for past month (50 min/session): %+v", views[0])
	}
}

func TestComptimeService_ListComptime_CurrentMonthPredictions(t *testing.T) {
	ctx := context.Background()
	// Reference "today": 2026-10-16 (halfway through a 30-day interval from Oct 1 to Oct 31)
	now := time.Date(2026, time.October, 16, 12, 0, 0, 0, time.UTC)
	octMonth := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)

	t.Run("no active plan sets PredictMinutes to 0 and adjusts Balance", func(t *testing.T) {
		repo := newMockComptimeRepo()
		min60 := 60
		repo.services[20] = &domain.Service{ID: 20, SessionMinutes: &min60}
		// Previous month balance was +60. In current month, Invoiced=240, Consumed=120 -> DB Balance = 60 + 240 - 120 = 180.
		repo.comptimes[1] = &domain.Comptime{
			ID:              1,
			VendorID:        1,
			CustomerID:      10,
			ServiceID:       20,
			Month:           octMonth,
			InvoicedMinutes: 240,
			ConsumedMinutes: 120,
			Balance:         180,
		}

		svc := NewComptime(repo, nil)
		svc.SetNowFunc(func() time.Time { return now })

		views, _, err := svc.ListComptime(ctx, port.ComptimeFilter{})
		if err != nil || len(views) != 1 {
			t.Fatalf("unexpected error or count: %v, %d", err, len(views))
		}
		v := views[0]
		if v.PredictMinutes != 0 {
			t.Errorf("expected PredictMinutes = 0 when no active plan, got %d", v.PredictMinutes)
		}
		if v.DoneMinutes != 120 {
			t.Errorf("expected DoneMinutes = 120, got %d", v.DoneMinutes)
		}
		if v.MonthBalance != -120 {
			t.Errorf("expected MonthBalance = -120, got %d", v.MonthBalance)
		}
		// previousBalance (60) + MonthBalance (-120) = -60
		if v.Balance != -60 {
			t.Errorf("expected Balance = -60, got %d", v.Balance)
		}
		if v.PredictSessions != 0 || v.DoneSessions != 2.0 || v.MonthBalanceSessions != -2.0 || v.BalanceSessions != -1.0 {
			t.Errorf("unexpected session counts: %+v", v)
		}
	})

	t.Run("notebook and payperview plans set PredictMinutes equal to DoneMinutes", func(t *testing.T) {
		for _, planType := range []domain.PlanType{domain.PlanTypeNotebook, domain.PlanTypePayPerView} {
			repo := newMockComptimeRepo()
			min60 := 60
			repo.services[20] = &domain.Service{ID: 20, SessionMinutes: &min60}
			repo.comptimes[1] = &domain.Comptime{
				ID:              1,
				VendorID:        1,
				CustomerID:      10,
				ServiceID:       20,
				Month:           octMonth,
				InvoicedMinutes: 0,
				ConsumedMinutes: 180,
				Balance:         -180,
			}
			repo.plans = []*domain.Plan{
				{
					ID:         100,
					CustomerID: 10,
					PlanType:   int(planType),
					PlanStart:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
					Items:      []domain.PlanItem{{ServiceID: 20, OrderIndex: 1}},
				},
			}

			svc := NewComptime(repo, nil)
			svc.SetNowFunc(func() time.Time { return now })

			views, _, err := svc.ListComptime(ctx, port.ComptimeFilter{})
			if err != nil || len(views) != 1 {
				t.Fatalf("unexpected error: %v", err)
			}
			v := views[0]
			if v.PredictMinutes != 180 || v.DoneMinutes != 180 || v.MonthBalance != 0 || v.Balance != 0 {
				t.Errorf("planType %d: expected Predict=180 Done=180 MonthBal=0 Bal=0, got %+v", planType, v)
			}
			if v.PredictSessions != 3.0 || v.DoneSessions != 3.0 {
				t.Errorf("planType %d: expected 3.0 sessions, got %+v", planType, v)
			}
		}
	})

	t.Run("package plan calculates proportional minutes between plan_start and plan_end", func(t *testing.T) {
		repo := newMockComptimeRepo()
		min60 := 60
		repo.services[20] = &domain.Service{ID: 20, SessionMinutes: &min60}
		repo.comptimes[1] = &domain.Comptime{
			ID:              1,
			VendorID:        1,
			CustomerID:      10,
			ServiceID:       20,
			Month:           octMonth,
			InvoicedMinutes: 600,
			ConsumedMinutes: 180,
			Balance:         420,
		}
		// Plan from Oct 1 to Oct 31 (30 days span), today is Oct 16 (15 days elapsed -> ratio 0.5)
		// Package quantity = 10 sessions * 60 min = 600 min -> 0.5 * 600 = 300 min predicted
		planEnd := time.Date(2026, time.October, 31, 0, 0, 0, 0, time.UTC)
		repo.plans = []*domain.Plan{
			{
				ID:         101,
				CustomerID: 10,
				PlanType:   int(domain.PlanTypePackage),
				PlanStart:  time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC),
				PlanEnd:    &planEnd,
				Package:    &domain.PlanPackage{Quantity: 10},
				Items:      []domain.PlanItem{{ServiceID: 20, OrderIndex: 1}},
			},
		}

		svc := NewComptime(repo, nil)
		svc.SetNowFunc(func() time.Time { return now })

		views, _, err := svc.ListComptime(ctx, port.ComptimeFilter{})
		if err != nil || len(views) != 1 {
			t.Fatalf("unexpected error: %v", err)
		}
		v := views[0]
		if v.PredictMinutes != 300 {
			t.Errorf("expected PredictMinutes = 300, got %d", v.PredictMinutes)
		}
		if v.DoneMinutes != 180 {
			t.Errorf("expected DoneMinutes = 180, got %d", v.DoneMinutes)
		}
		if v.MonthBalance != 120 || v.Balance != 120 {
			t.Errorf("expected MonthBalance=120 Balance=120, got %d / %d", v.MonthBalance, v.Balance)
		}
		if v.PredictSessions != 5.0 || v.DoneSessions != 3.0 || v.MonthBalanceSessions != 2.0 {
			t.Errorf("unexpected session counts: %+v", v)
		}
	})

	t.Run("agenda plan calculates monthly sessions and proportional minutes", func(t *testing.T) {
		repo := newMockComptimeRepo()
		min60 := 60
		repo.services[20] = &domain.Service{ID: 20, SessionMinutes: &min60}
		repo.comptimes[1] = &domain.Comptime{
			ID:              1,
			VendorID:        1,
			CustomerID:      10,
			ServiceID:       20,
			Month:           octMonth,
			InvoicedMinutes: 240,
			ConsumedMinutes: 90,
			Balance:         150,
		}
		// October 2026: Oct 1 is Thursday (WeekDay=5). Thursdays in Oct 2026: 1, 8, 15, 22, 29 (5 Thursdays).
		// With MonthlyServiceLimit = 4 -> 4 sessions * 60 min = 240 min.
		// effStart = Oct 1, effEnd = Oct 31 (30 days), today = Oct 16 (15 days -> 0.5) -> 120 min predicted.
		limit := 4
		repo.plans = []*domain.Plan{
			{
				ID:         102,
				CustomerID: 10,
				PlanType:   int(domain.PlanTypeAgenda),
				PlanStart:  time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
				Agenda: &domain.PlanAgenda{
					Recurrence:          int(domain.RecurrenceWeekly),
					WeekDay:             5, // Thursday
					MonthlyServiceLimit: &limit,
				},
				Items: []domain.PlanItem{{ServiceID: 20, OrderIndex: 1}},
			},
		}

		svc := NewComptime(repo, nil)
		svc.SetNowFunc(func() time.Time { return now })

		views, _, err := svc.ListComptime(ctx, port.ComptimeFilter{})
		if err != nil || len(views) != 1 {
			t.Fatalf("unexpected error: %v", err)
		}
		v := views[0]
		if v.PredictMinutes != 120 {
			t.Errorf("expected PredictMinutes = 120, got %d", v.PredictMinutes)
		}
		if v.DoneMinutes != 90 {
			t.Errorf("expected DoneMinutes = 90, got %d", v.DoneMinutes)
		}
		if v.MonthBalance != 30 || v.Balance != 30 {
			t.Errorf("expected MonthBalance=30 Balance=30, got %d / %d", v.MonthBalance, v.Balance)
		}
		if v.PredictSessions != 2.0 || v.DoneSessions != 1.5 || v.MonthBalanceSessions != 0.5 || v.BalanceSessions != 0.5 {
			t.Errorf("unexpected session counts: %+v", v)
		}
	})
}
