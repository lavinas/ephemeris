package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

type mockInvoicingRepoForService struct {
	plans    []domain.Plan
	invoices []domain.Invoice
}

func newMockInvoicingRepoForService() *mockInvoicingRepoForService {
	return &mockInvoicingRepoForService{}
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
func (m *mockInvoicingRepoForService) GetPlanByID(id int64) (*domain.Plan, error) { return nil, nil }
func (m *mockInvoicingRepoForService) DeletePlan(id int64) error                    { return nil }
func (m *mockInvoicingRepoForService) FindCustomerActivePlansWithServices(customerID int64, serviceIDs []int64, excludePlanID int64) ([]domain.Plan, error) {
	return nil, nil
}

func (m *mockInvoicingRepoForService) SaveInvoice(invoice *domain.Invoice) error {
	m.invoices = append(m.invoices, *invoice)
	return nil
}
func (m *mockInvoicingRepoForService) GetInvoiceByID(id int64) (*domain.Invoice, error) {
	return nil, nil
}
func (m *mockInvoicingRepoForService) FindInvoices(conditions map[string]interface{}, orderBy ...string) ([]domain.Invoice, error) {
	return m.invoices, nil
}
func (m *mockInvoicingRepoForService) DeleteInvoice(id int64) error { return nil }

func (m *mockInvoicingRepoForService) FindActivePlansForInvoicing(vendorID *int64, customerID *int64, planType *int, serviceID *int64, startDate, endDate time.Time) ([]domain.Plan, error) {
	var result []domain.Plan
	for _, p := range m.plans {
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
	return nil, nil
}
func (m *mockInvoicingRepoForService) GetInvoiceForSession(sessionID int64) (*domain.Invoice, error) {
	return nil, nil
}

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
}
