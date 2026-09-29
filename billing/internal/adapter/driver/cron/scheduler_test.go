package cron

import (
	"os"
	"sync"
	"testing"
	"time"

	"billing/internal/domain"
	"billing/internal/port"
	"billing/internal/service"
)

type cronMockLogger struct {
	mu   sync.Mutex
	logs []string
}

func (l *cronMockLogger) IPrintf(level int, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, format)
}

func (l *cronMockLogger) Close() {}

func (l *cronMockLogger) HasLogContaining(substr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, log := range l.logs {
		if len(log) >= len(substr) {
			for i := 0; i <= len(log)-len(substr); i++ {
				if log[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

type cronMockRepo struct {
	invoices map[int64]*domain.Invoice
	vendors  map[int64]*domain.Vendor
}

func (m *cronMockRepo) Save(model interface{}) error   { return nil }
func (m *cronMockRepo) BeginTransaction() error        { return nil }
func (m *cronMockRepo) CommitTransaction() error       { return nil }
func (m *cronMockRepo) RollbackTransaction() error     { return nil }
func (m *cronMockRepo) Close() error                   { return nil }

func (m *cronMockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname,
	document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *cronMockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (m *cronMockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string,
	accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *cronMockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	for _, v := range m.vendors {
		if v.Nickname == nickname {
			return v, nil
		}
	}
	return nil, nil
}
func (m *cronMockRepo) GetVendorByID(id int64) (*domain.Vendor, error) {
	if v, ok := m.vendors[id]; ok {
		return v, nil
	}
	return nil, nil
}
func (m *cronMockRepo) FindInvoices(page, pageSize int, customer int64,
	invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate,
	taxDate, cancellationDate *string) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cronMockRepo) FindInvoicesPendingSend(dueBeforeOrEqual time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cronMockRepo) GetInvoicesByPeriod(vendorID int64, start, end time.Time) ([]domain.Invoice, error) {
	return nil, nil
}
func (m *cronMockRepo) GetInvoice(id int64) (*domain.Invoice, error) {
	return nil, nil
}
func (m *cronMockRepo) GetEmissions(vendorID int64, invoiceStartDate, invoiceEndDate time.Time) ([]domain.Emission, error) {
	return nil, nil
}
func (m *cronMockRepo) GetEmissionLastRPS(vendorID int64) (int64, error) {
	return 0, nil
}
func (m *cronMockRepo) GetEmission(id int64) (*domain.Emission, error) {
	return nil, nil
}

type cronMockIssuer struct{}

func (i *cronMockIssuer) GetBase64(data port.InDTO, html_pdf string) ([]byte, error) {
	return []byte("pdf"), nil
}
func (i *cronMockIssuer) SendMail(data port.InDTO, subject, filename, html_pdf, html_email string) error {
	return nil
}

type cronMockPixer struct{}

func (p *cronMockPixer) Get(request port.InDTO) (string, string, error) {
	return "pix", "qr", nil
}

func TestScheduler_StartStop(t *testing.T) {
	logger := &cronMockLogger{}
	repo := &cronMockRepo{}
	billSvc := service.NewBill(repo, logger, &cronMockIssuer{}, &cronMockPixer{})
	autoBillSvc := service.NewInvoiceAutoBill(repo, logger, billSvc, 3)

	schedules := []string{"0 8 * * *", "0 14 * * *"}
	scheduler, err := NewScheduler(autoBillSvc, logger, 3, schedules, "America/Sao_Paulo")
	if err != nil {
		t.Fatalf("unexpected error creating scheduler: %v", err)
	}

	if err := scheduler.Start(); err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}

	if err := scheduler.Stop(); err != nil {
		t.Fatalf("failed to stop scheduler: %v", err)
	}
}

func TestScheduler_OverlapPrevention(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir("../../../.."); err != nil {
		t.Fatalf("could not chdir to billing root: %v", err)
	}

	logger := &cronMockLogger{}
	repo := &cronMockRepo{}
	billSvc := service.NewBill(repo, logger, &cronMockIssuer{}, &cronMockPixer{})
	autoBillSvc := service.NewInvoiceAutoBill(repo, logger, billSvc, 3)

	scheduler, err := NewScheduler(autoBillSvc, logger, 3, []string{"0 8 * * *"}, "America/Sao_Paulo")
	if err != nil {
		t.Fatalf("unexpected error creating scheduler: %v", err)
	}

	// Manually set isRunning to true to simulate an ongoing job
	scheduler.isRunning.Store(true)

	// Attempt to execute another job
	scheduler.ExecuteJob()

	// Verify that overlap warning was logged
	if !logger.HasLogContaining("overlap detected") {
		t.Errorf("expected log to contain overlap detected message")
	}
}
