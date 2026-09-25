package dto

import (
	"strings"
	"testing"
	"time"

	"billing/internal/domain"
)

type mockRepo struct{}

func (m *mockRepo) Save(model interface{}) error    { return nil }
func (m *mockRepo) BeginTransaction() error         { return nil }
func (m *mockRepo) CommitTransaction() error        { return nil }
func (m *mockRepo) RollbackTransaction() error      { return nil }
func (m *mockRepo) Close() error                    { return nil }
func (m *mockRepo) FindCustomers(page, pageSize int, vendorID int64, name, nickname, document *string, status *int, email, whatsapp *string) ([]domain.Customer, error) {
	return nil, nil
}
func (m *mockRepo) GetCustomer(vendorID int64, nickname string) (*domain.Customer, error) {
	return nil, nil
}
func (m *mockRepo) FindVendors(page, pageSize int, legalName, nickname, document *string, accountBank, accountAgency, accountNumber *string) ([]domain.Vendor, error) {
	return nil, nil
}
func (m *mockRepo) GetVendor(nickname string) (*domain.Vendor, error) {
	return nil, nil
}
func (m *mockRepo) FindInvoices(page, pageSize int, customer int64, invoiceDate, dueDate, paymentDate, emailSentDate, whatsappSentDate, emailReceiptDate, whatsappReceiptDate, taxDate, cancellationDate *string) ([]domain.Invoice, error) {
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

func validBillingVendorCreateRequest() *VendorCreateRequest {
	return &VendorCreateRequest{
		Nickname:      "valid_vendor",
		LegalName:     "Valid Vendor LTDA",
		TradingName:   "", // Optional
		Document:      "27.928.875/0001-04",
		TaxDocument:   "5.727.888-1",
		AccountBank:   "033 - Santander",
		AccountAgency: "0985",
		AccountNumber: "13001001-4",
		PixToken:      "27.928.875/0001-04",
		PixName:       "Valid Vendor",
		PixCity:       "São Paulo",
		LogoName:      "logo.png",
		Email:         "vendor@example.com",
		Whatsapp:      "(11) 98088-8399",
		LastRps:       0,
		SmtpHost:      "smtp.example.com",
		SmtpPort:      587,
		SmtpUser:      "vendor@example.com",
		SmtpPassword:  "password",
	}
}

func TestBillingVendorCreateRequest_Validate_Success(t *testing.T) {
	repo := &mockRepo{}
	req := validBillingVendorCreateRequest()
	if err := req.Validate(repo); err != nil {
		t.Fatalf("expected valid request to pass, got: %v", err)
	}

	// Also should pass with trading_name populated
	req.TradingName = "Valid Vendor Fantasia"
	if err := req.Validate(repo); err != nil {
		t.Fatalf("expected valid request with trading name to pass, got: %v", err)
	}
}

func TestBillingVendorCreateRequest_Validate_RequiredFields(t *testing.T) {
	repo := &mockRepo{}
	tests := []struct {
		name        string
		modify      func(r *VendorCreateRequest)
		expectedErr string
	}{
		{
			name:        "missing nickname",
			modify:      func(r *VendorCreateRequest) { r.Nickname = "" },
			expectedErr: "nickname is required",
		},
		{
			name:        "missing legal_name",
			modify:      func(r *VendorCreateRequest) { r.LegalName = "" },
			expectedErr: "legal_name is required",
		},
		{
			name:        "missing document",
			modify:      func(r *VendorCreateRequest) { r.Document = "" },
			expectedErr: "document is required",
		},
		{
			name:        "missing tax_document",
			modify:      func(r *VendorCreateRequest) { r.TaxDocument = "" },
			expectedErr: "tax_document is required",
		},
		{
			name:        "missing account_bank",
			modify:      func(r *VendorCreateRequest) { r.AccountBank = "" },
			expectedErr: "account_bank is required",
		},
		{
			name:        "missing account_agency",
			modify:      func(r *VendorCreateRequest) { r.AccountAgency = "" },
			expectedErr: "account_agency is required",
		},
		{
			name:        "missing account_number",
			modify:      func(r *VendorCreateRequest) { r.AccountNumber = "" },
			expectedErr: "account_number is required",
		},
		{
			name:        "missing pix_token",
			modify:      func(r *VendorCreateRequest) { r.PixToken = "" },
			expectedErr: "pix_token is required",
		},
		{
			name:        "missing pix_name",
			modify:      func(r *VendorCreateRequest) { r.PixName = "" },
			expectedErr: "pix_name is required",
		},
		{
			name:        "missing pix_city",
			modify:      func(r *VendorCreateRequest) { r.PixCity = "" },
			expectedErr: "pix_city is required",
		},
		{
			name:        "missing logo_name",
			modify:      func(r *VendorCreateRequest) { r.LogoName = "" },
			expectedErr: "logo_name is required",
		},
		{
			name:        "missing email",
			modify:      func(r *VendorCreateRequest) { r.Email = "" },
			expectedErr: "email is required",
		},
		{
			name:        "missing whatsapp",
			modify:      func(r *VendorCreateRequest) { r.Whatsapp = "" },
			expectedErr: "whatsapp is required",
		},
		{
			name:        "negative last_rps",
			modify:      func(r *VendorCreateRequest) { r.LastRps = -1 },
			expectedErr: "last_rps must be greater than or equal to zero",
		},
		{
			name:        "missing smtp_host",
			modify:      func(r *VendorCreateRequest) { r.SmtpHost = "" },
			expectedErr: "smtp_host is required",
		},
		{
			name:        "zero smtp_port",
			modify:      func(r *VendorCreateRequest) { r.SmtpPort = 0 },
			expectedErr: "smtp_port is required and must be greater than zero",
		},
		{
			name:        "negative smtp_port",
			modify:      func(r *VendorCreateRequest) { r.SmtpPort = -5 },
			expectedErr: "smtp_port is required and must be greater than zero",
		},
		{
			name:        "missing smtp_user",
			modify:      func(r *VendorCreateRequest) { r.SmtpUser = "" },
			expectedErr: "smtp_user is required",
		},
		{
			name:        "missing smtp_password",
			modify:      func(r *VendorCreateRequest) { r.SmtpPassword = "" },
			expectedErr: "smtp_password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validBillingVendorCreateRequest()
			tt.modify(req)
			err := req.Validate(repo)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}
