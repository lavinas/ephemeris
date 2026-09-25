package messaging

import (
	"context"
	"testing"
	"time"

	"billing/internal/domain"

	"github.com/nats-io/nats.go"
)

type mockLogger struct {
	logs []string
}

func (m *mockLogger) IPrintf(level int, format string, v ...interface{}) {
	m.logs = append(m.logs, format)
}

func (m *mockLogger) Close() {}

type mockRepo struct {
	customers map[string]*domain.Customer
	vendors   map[string]*domain.Vendor
	saved     []interface{}
}

func newMockRepo() *mockRepo {
	r := &mockRepo{
		customers: make(map[string]*domain.Customer),
		vendors:   make(map[string]*domain.Vendor),
	}
	r.vendors["acme"] = &domain.Vendor{
		ID:       1,
		Nickname: "acme",
	}
	return r
}

func (m *mockRepo) Save(model interface{}) error {
	m.saved = append(m.saved, model)
	switch v := model.(type) {
	case *[]domain.Customer:
		for _, c := range *v {
			cCopy := c
			m.customers[c.Nickname] = &cCopy
		}
	case *domain.Customer:
		m.customers[v.Nickname] = v
	case *domain.Vendor:
		m.vendors[v.Nickname] = v
	case *[]domain.Vendor:
		for _, vd := range *v {
			vCopy := vd
			m.vendors[vd.Nickname] = &vCopy
		}
	}
	return nil
}

func (m *mockRepo) BeginTransaction() error    { return nil }
func (m *mockRepo) CommitTransaction() error   { return nil }
func (m *mockRepo) RollbackTransaction() error { return nil }
func (m *mockRepo) Close() error               { return nil }

func (m *mockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	var list []domain.Customer
	for _, c := range m.customers {
		if nickname != nil && c.Nickname == *nickname {
			list = append(list, *c)
		} else if document != nil && c.Document != nil && *c.Document == *document {
			list = append(list, *c)
		}
	}
	return list, nil
}

func (m *mockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	if c, ok := m.customers[nickname]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	var list []domain.Vendor
	for _, v := range m.vendors {
		if nickname != nil && v.Nickname == *nickname {
			list = append(list, *v)
		} else if document != nil && v.Document == *document {
			list = append(list, *v)
		}
	}
	return list, nil
}

func (m *mockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	if v, ok := m.vendors[nickname]; ok {
		return v, nil
	}
	return nil, nil
}

func (m *mockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate,
	taxDate, cancellationDate *string) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *mockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}

func (m *mockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	return nil, nil
}

func (m *mockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}

func (m *mockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}

func (m *mockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

func TestNoopConsumer(t *testing.T) {
	c := NewNoopConsumer()
	if err := c.Start(context.Background()); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if err := c.Close(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestNATSConsumer_HandleCustomerCreated(t *testing.T) {
	repo := newMockRepo()
	logger := &mockLogger{}
	consumer := NewNATSConsumer("nats://dummy:4222", "test-group", repo, logger)

	payload := `{
		"event_type": "customer.created",
		"timestamp": "2026-09-23T21:00:00Z",
		"vendor": "acme",
		"data": {
			"name": "Sync Client",
			"nickname": "syncclient",
			"document": "11144477735",
			"email": "sync@example.com",
			"whatsapp": "+5511999998888"
		}
	}`

	msg := &nats.Msg{
		Subject: SubjectCustomerCreated,
		Data:    []byte(payload),
	}

	consumer.handleCustomerCreated(msg)

	cust, ok := repo.customers["syncclient"]
	if !ok {
		t.Fatalf("expected customer 'syncclient' to be created in repo")
	}
	if cust.Name != "Sync Client" {
		t.Errorf("expected name 'Sync Client', got '%s'", cust.Name)
	}
}

func TestNATSConsumer_HandleCustomerUpdated(t *testing.T) {
	repo := newMockRepo()
	logger := &mockLogger{}
	consumer := NewNATSConsumer("nats://dummy:4222", "test-group", repo, logger)

	// Pre-seed customer
	doc := "11144477735"
	email := "sync@example.com"
	phone := "+5511999998888"
	existing := &domain.Customer{
		ID:        1,
		VendorID:  1,
		Name:      "Sync Client",
		Nickname:  "syncclient",
		Document:  &doc,
		Email:     &email,
		Whatsapp:  &phone,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.customers["syncclient"] = existing

	payload := `{
		"event_type": "customer.updated",
		"timestamp": "2026-09-23T21:05:00Z",
		"vendor": "acme",
		"data": {
			"name": "Sync Client Updated",
			"nickname": "syncclient",
			"whatsapp": "+5511988887777"
		}
	}`

	msg := &nats.Msg{
		Subject: SubjectCustomerUpdated,
		Data:    []byte(payload),
	}

	consumer.handleCustomerUpdated(msg)

	cust, ok := repo.customers["syncclient"]
	if !ok {
		t.Fatalf("expected customer 'syncclient' to exist in repo")
	}
	if cust.Name != "Sync Client Updated" {
		t.Errorf("expected updated name 'Sync Client Updated', got '%s'", cust.Name)
	}
	if cust.Whatsapp == nil || *cust.Whatsapp != "+5511988887777" {
		t.Errorf("expected updated whatsapp, got %v", cust.Whatsapp)
	}
}

func TestNATSConsumer_HandleVendorCreated(t *testing.T) {
	repo := newMockRepo()
	logger := &mockLogger{}
	consumer := NewNATSConsumer("nats://dummy:4222", "test-group", repo, logger)

	payload := `{
		"event_type": "vendor.created",
		"timestamp": "2026-09-24T21:00:00Z",
		"vendor": "new_vendor",
		"data": {
			"nickname": "new_vendor",
			"legal_name": "New Vendor LTDA",
			"trading_name": "New Vendor",
			"document": "27.928.875/0001-04",
			"email": "vendor@test.com",
			"whatsapp": "+5511980888399"
		}
	}`

	msg := &nats.Msg{
		Subject: SubjectVendorCreated,
		Data:    []byte(payload),
	}

	consumer.handleVendorCreated(msg)

	vnd, ok := repo.vendors["new_vendor"]
	if !ok {
		t.Fatalf("expected vendor 'new_vendor' to be created in repo")
	}
	if vnd.LegalName != "New Vendor LTDA" {
		t.Errorf("expected legal name 'New Vendor LTDA', got '%s'", vnd.LegalName)
	}
	if vnd.Email != "vendor@test.com" {
		t.Errorf("expected email 'vendor@test.com', got '%s'", vnd.Email)
	}
}

func TestNATSConsumer_HandleVendorUpdated(t *testing.T) {
	repo := newMockRepo()
	logger := &mockLogger{}
	consumer := NewNATSConsumer("nats://dummy:4222", "test-group", repo, logger)

	// Pre-seed vendor
	existing := &domain.Vendor{
		ID:        1,
		Nickname:  "acme",
		LegalName: "Acme Old",
		Document:  "27.928.875/0001-04",
		Email:     "old@acme.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.vendors["acme"] = existing

	payload := `{
		"event_type": "vendor.updated",
		"timestamp": "2026-09-24T21:05:00Z",
		"vendor": "acme",
		"data": {
			"nickname": "acme",
			"legal_name": "Acme Updated LTDA",
			"email": "updated@acme.com"
		}
	}`

	msg := &nats.Msg{
		Subject: SubjectVendorUpdated,
		Data:    []byte(payload),
	}

	consumer.handleVendorUpdated(msg)

	vnd, ok := repo.vendors["acme"]
	if !ok {
		t.Fatalf("expected vendor 'acme' to exist in repo")
	}
	if vnd.LegalName != "Acme Updated LTDA" {
		t.Errorf("expected updated legal name 'Acme Updated LTDA', got '%s'", vnd.LegalName)
	}
	if vnd.Email != "updated@acme.com" {
		t.Errorf("expected updated email 'updated@acme.com', got '%s'", vnd.Email)
	}
}

