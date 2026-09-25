package dto

import (
	"strings"
	"testing"
)

type mockRepo struct{}

func (m *mockRepo) BeginTransaction() error    { return nil }
func (m *mockRepo) CommitTransaction() error   { return nil }
func (m *mockRepo) RollbackTransaction() error { return nil }
func (m *mockRepo) Save(model interface{}) error { return nil }
func (m *mockRepo) Find(model interface{}, conditions map[string]interface{}, page, pagesize int, orderBy ...string) ([]interface{}, error) {
	return []interface{}{}, nil
}
func (m *mockRepo) Close() error { return nil }

func validVendorCreateRequest() *VendorCreateRequest {
	req := NewVendorCreateRequest(&mockRepo{})
	req.Nickname = "valid_vendor"
	req.LegalName = "Valid Vendor LTDA"
	req.TradingName = "" // Optional
	req.Document = "27.928.875/0001-04"
	req.TaxDocument = "5.727.888-1"
	req.AccountBank = "033 - Santander"
	req.AccountAgency = "0985"
	req.AccountNumber = "13001001-4"
	req.PixToken = "27.928.875/0001-04"
	req.PixName = "Valid Vendor"
	req.PixCity = "São Paulo"
	req.LogoName = "logo.png"
	req.Email = "vendor@example.com"
	req.Whatsapp = "(11) 98088-8399"
	req.LastRps = 0
	req.SmtpHost = "smtp.example.com"
	req.SmtpPort = 587
	req.SmtpUser = "vendor@example.com"
	req.SmtpPassword = "password"
	return req
}

func TestVendorCreateRequest_Validate_Success(t *testing.T) {
	req := validVendorCreateRequest()
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request to pass, got: %v", err)
	}

	// Also should pass with trading_name populated
	req.TradingName = "Valid Vendor Fantasia"
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid request with trading name to pass, got: %v", err)
	}
}

func TestVendorCreateRequest_Validate_RequiredFields(t *testing.T) {
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
			req := validVendorCreateRequest()
			tt.modify(req)
			err := req.Validate()
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.expectedErr)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
			}
		})
	}
}
