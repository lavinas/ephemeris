package service

import (
	"context"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

type mockStandaloneInvoiceRepo struct {
	mockPlanRepoForService
	*mockComptimeRepo
	invoices map[int64]*domain.Invoice
	nextID   int64
}

func newMockStandaloneInvoiceRepo() *mockStandaloneInvoiceRepo {
	return &mockStandaloneInvoiceRepo{
		mockComptimeRepo: newMockComptimeRepo(),
		invoices:         make(map[int64]*domain.Invoice),
		nextID:           100,
	}
}

func (m *mockStandaloneInvoiceRepo) GetServiceByID(id int64) (*domain.Service, error) {
	return m.mockComptimeRepo.GetServiceByID(id)
}

func (m *mockStandaloneInvoiceRepo) GetSessionByID(id int64) (*domain.Session, error) {
	return m.mockComptimeRepo.GetSessionByID(id)
}

func (m *mockStandaloneInvoiceRepo) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return m.invoices[id], nil
}

func (m *mockStandaloneInvoiceRepo) FindServices(vendorID int64) ([]domain.Service, error) {
	var list []domain.Service
	for _, s := range m.mockComptimeRepo.services {
		list = append(list, *s)
	}
	return list, nil
}

func (m *mockStandaloneInvoiceRepo) GetPlanByID(id int64) (*domain.Plan, error) {
	for _, p := range m.mockComptimeRepo.plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockStandaloneInvoiceRepo) GetCustomerByID(id int64) (*domain.Customer, error) {
	return &domain.Customer{ID: id, VendorID: 1, Nickname: "cli_teste", Name: "Cliente Teste"}, nil
}

func (m *mockStandaloneInvoiceRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	return &domain.Vendor{ID: id, Nickname: "vend_teste"}, nil
}

func (m *mockStandaloneInvoiceRepo) UpdateInvoice(invoice *domain.Invoice) error {
	now := time.Now()
	for i := range invoice.Items {
		it := &invoice.Items[i]
		if it.ID == 0 && it.DeletedAt == nil {
			m.nextID++
			it.ID = m.nextID
			it.InvoiceID = invoice.ID
			it.CreatedAt = now
			it.UpdatedAt = now
			itemCopy := *it
			itemCopy.Invoice = invoice
			if itemCopy.Service == nil {
				itemCopy.Service = m.mockComptimeRepo.services[it.ServiceID]
			}
			m.mockComptimeRepo.invoiceItems[it.ID] = &itemCopy
		} else if it.ID > 0 && it.DeletedAt != nil {
			if existing, ok := m.mockComptimeRepo.invoiceItems[it.ID]; ok {
				existing.DeletedAt = it.DeletedAt
			}
		} else if it.ID > 0 {
			if existing, ok := m.mockComptimeRepo.invoiceItems[it.ID]; ok {
				existing.Quantity = it.Quantity
				existing.InvoicedQuantity = it.InvoicedQuantity
			}
		}
	}
	m.invoices[invoice.ID] = invoice
	return nil
}

type mockStandaloneBillingClient struct {
	cancelled []port.BillingCancelInput
	created   []port.BillingInvoiceInput
	nextRefID int64
}

func (b *mockStandaloneBillingClient) CreateInvoice(_ context.Context, input port.BillingInvoiceInput) (int64, error) {
	b.created = append(b.created, input)
	b.nextRefID++
	return b.nextRefID, nil
}

func (b *mockStandaloneBillingClient) CancelInvoice(_ context.Context, input port.BillingCancelInput) error {
	b.cancelled = append(b.cancelled, input)
	return nil
}

func (b *mockStandaloneBillingClient) GetInvoiceByExternalCode(_ context.Context, _ string, _ string) (*int64, error) {
	return nil, nil
}

func TestInvoiceItemAddAndDelete_WithComptimeAndBilling(t *testing.T) {
	nextID := int64(200)
	repo := newMockStandaloneInvoiceRepo()
	repo.nextID = nextID
	logger := &mockLogger{}
	ctx := context.Background()

	minutes := 60
	svcEntity := &domain.Service{ID: 10, Name: "Psicoterapia", SessionMinutes: &minutes}
	repo.mockComptimeRepo.services[10] = svcEntity

	plan := &domain.Plan{
		ID:         1,
		CustomerID: 50,
		PlanType:   int(domain.PlanTypeAgenda),
		Customer:   &domain.Customer{ID: 50, VendorID: 1, Nickname: "cli_teste"},
	}
	repo.mockComptimeRepo.plans = append(repo.mockComptimeRepo.plans, plan)

	octStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	octEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	baseDesc := "Psicoterapia - 01/10/2026 a 31/10/2026"
	refID := int64(900)

	inv := &domain.Invoice{
		ID:                 1,
		PlanID:             1,
		Plan:               plan,
		InvoicingDate:      octStart,
		DueDate:            time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
		Amount:             600.0,
		ReferenceInvoiceID: &refID,
		Items: []domain.InvoiceItem{
			{
				ID:               101,
				InvoiceID:        1,
				ServiceID:        10,
				Service:          svcEntity,
				Price:            150.0,
				Description:      &baseDesc,
				Quantity:         4,
				InvoicedQuantity: 4,
				RealizedQuantity: 4,
				PeriodStart:      octStart,
				PeriodEnd:        octEnd,
			},
		},
	}
	repo.invoices[1] = inv
	itemCopy := inv.Items[0]
	itemCopy.Invoice = inv
	repo.mockComptimeRepo.invoiceItems[101] = &itemCopy

	comptimeSvc := NewComptime(repo, logger)
	if err := comptimeSvc.ProcessInvoiceItemCreated(ctx, 101); err != nil {
		t.Fatalf("failed initial comptime credit: %v", err)
	}

	billingClient := &mockStandaloneBillingClient{nextRefID: 900}
	addSvc := NewInvoiceItemAdd(repo, logger, billingClient)
	addSvc.SetComptimeService(comptimeSvc)

	// 1. Add a standalone addition item (+2 sessions in October = +R$ 300,00 and +120 min in comptime)
	addReq := &dto.InvoiceItemAddRequest{
		InvoiceID:   1,
		ServiceID:   10,
		Operation:   "add",
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Price:       150.0,
		Quantity:    2,
	}
	addResp := addSvc.Run(addReq)
	if addResp.GetStatusCode() != 201 {
		t.Fatalf("expected 201 on add item, got %d: %s", addResp.GetStatusCode(), addResp.GetMessage())
	}

	if inv.Amount != 900.0 {
		t.Errorf("expected invoice Amount 900.0 after addition, got %.2f", inv.Amount)
	}
	compOct, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 50, 10, octStart)
	if compOct == nil || compOct.InvoicedMinutes != 360 {
		t.Errorf("expected 360 invoiced minutes (240 + 120), got %+v", compOct)
	}

	// 2. Add a standalone discount item (-1 session in September = -R$ 150,00 and -60 min in September comptime)
	sepStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	discReq := &dto.InvoiceItemAddRequest{
		InvoiceID:   1,
		ServiceID:   10,
		Operation:   "discount",
		PeriodStart: "2026-09-01",
		PeriodEnd:   "2026-09-30",
		Price:       150.0,
		Quantity:    1,
	}
	discResp := addSvc.Run(discReq)
	if discResp.GetStatusCode() != 201 {
		t.Fatalf("expected 201 on discount item, got %d: %s", discResp.GetStatusCode(), discResp.GetMessage())
	}
	discOut := discResp.(*dto.InvoiceItemAddResponse)
	if discOut.Item == nil || discOut.Item.Price != -150.0 {
		t.Fatalf("expected discount item price -150.0, got %+v", discOut.Item)
	}
	if inv.Amount != 750.0 {
		t.Errorf("expected invoice Amount 750.0 after discount (900 - 150), got %.2f", inv.Amount)
	}

	compSep, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 50, 10, sepStart)
	if compSep == nil || compSep.InvoicedMinutes != -60 {
		t.Errorf("expected -60 invoiced minutes in September for discount item, got %+v", compSep)
	}
	// October balance should also reflect the cascaded -60 from September (360 - 60 = 300)
	compOctAfterDisc, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 50, 10, octStart)
	if compOctAfterDisc.Balance != 300 {
		t.Errorf("expected October balance 300 (360 + (-60) from Sep), got %d", compOctAfterDisc.Balance)
	}

	// Verify billing client received updated invoice with 3 items (including negative discount item) and Amount=750
	if len(billingClient.created) != 2 {
		t.Fatalf("expected 2 billing re-creations, got %d", len(billingClient.created))
	}
	lastBilling := billingClient.created[1]
	if lastBilling.Amount != 750.0 || len(lastBilling.Items) != 3 {
		t.Errorf("expected billing Amount=750.0 and 3 items, got Amount=%.2f items=%d", lastBilling.Amount, len(lastBilling.Items))
	}

	// 3. Delete the discount item -> restores invoice Amount to 900.0 and reverses -60 min in September comptime
	delSvc := NewInvoiceItemDelete(repo, logger, billingClient)
	delSvc.SetComptimeService(comptimeSvc)
	delReq := &dto.InvoiceItemDeleteRequest{
		ItemID:    discOut.Item.ID,
		InvoiceID: 1,
	}
	delResp := delSvc.Run(delReq)
	if delResp.GetStatusCode() != 200 {
		t.Fatalf("expected 200 on delete item, got %d: %s", delResp.GetStatusCode(), delResp.GetMessage())
	}
	if inv.Amount != 900.0 {
		t.Errorf("expected invoice Amount restored to 900.0 after deleting discount item, got %.2f", inv.Amount)
	}
	compSepAfterDel, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 50, 10, sepStart)
	if compSepAfterDel.InvoicedMinutes != 0 {
		t.Errorf("expected 0 invoiced minutes in September after deleting discount item, got %d", compSepAfterDel.InvoicedMinutes)
	}
	compOctAfterDel, _ := repo.GetComptimeByCustomerServiceMonth(ctx, 50, 10, octStart)
	if compOctAfterDel.Balance != 360 {
		t.Errorf("expected October balance restored to 360, got %d", compOctAfterDel.Balance)
	}
}
