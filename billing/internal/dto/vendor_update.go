package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/port"

	"github.com/klassmann/cpfcnpj"
)

// VendorUpdateRequest represents the request payload for updating an existing vendor.
type VendorUpdateRequest struct {
	Nickname      string         `json:"nickname" validate:"required"`
	LegalName     *string        `json:"legal_name"`
	TradingName   *string        `json:"trading_name"`
	Document      *string        `json:"document"`
	TaxDocument   *string        `json:"tax_document"`
	AccountBank   *string        `json:"account_bank"`
	AccountAgency *string        `json:"account_agency"`
	AccountNumber *string        `json:"account_number"`
	PixToken      *string        `json:"pix_token"`
	PixName       *string        `json:"pix_name"`
	PixCity       *string        `json:"pix_city"`
	LogoName      *string        `json:"logo_name"`
	Email         *string        `json:"email"`
	Whatsapp      *string        `json:"whatsapp"`
	LastRps       *int64         `json:"last_rps"`
	SmtpHost      *string        `json:"smtp_host"`
	SmtpPort      *int           `json:"smtp_port"`
	SmtpUser      *string        `json:"smtp_user"`
	SmtpPassword  *string        `json:"smtp_password"`
	vendor        *domain.Vendor `json:"-" validate:"-"`
}

// VendorUpdateResponse represents the response payload after updating an existing vendor.
type VendorUpdateResponse struct {
	ResponseBase
}

// NewVendorUpdateResponse creates a new instance of VendorUpdateResponse.
func NewVendorUpdateResponse(httpCode int, status, message string) VendorUpdateResponse {
	return VendorUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the VendorUpdateRequest has all required fields and valid data.
func (r *VendorUpdateRequest) Validate(repo port.Repository) error {
	errs := make([]error, 0)
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateLegalName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDocument(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAtLeastOneField(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *VendorUpdateRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	vendor, err := repo.GetVendor(r.Nickname)
	if err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	}
	if vendor == nil {
		return fmt.Errorf("vendor with nickname '%s' does not exist", r.Nickname)
	}
	r.vendor = vendor
	return nil
}

func (r *VendorUpdateRequest) validateLegalName() error {
	if r.LegalName == nil {
		return nil
	}
	if *r.LegalName == "" {
		return fmt.Errorf("legal_name cannot be empty")
	}
	return nil
}

func (r *VendorUpdateRequest) validateDocument(repo port.Repository) error {
	if r.Document == nil {
		return nil
	}
	if *r.Document == "" {
		return fmt.Errorf("document cannot be empty")
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	exist, err := repo.FindVendors(0, 0, nil, nil, r.Document, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if len(exist) > 0 && r.vendor != nil && exist[0].ID != r.vendor.ID {
		return fmt.Errorf("document '%s' is already in use by another vendor", *r.Document)
	}
	return nil
}

func (r *VendorUpdateRequest) validateCpfCnpj() error {
	if r.Document == nil {
		return nil
	}
	cpf := cpfcnpj.NewCPF(*r.Document)
	if cpf.IsValid() {
		*r.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(*r.Document)
	if cnpj.IsValid() {
		*r.Document = cnpj.String()
		return nil
	}
	return fmt.Errorf("document '%s' is not a valid CPF or CNPJ", *r.Document)
}

func (r *VendorUpdateRequest) validateWhatsapp() error {
	if r.Whatsapp == nil {
		return nil
	}
	if *r.Whatsapp == "" {
		return fmt.Errorf("whatsapp cannot be empty")
	}
	num, err := ValidateCellNumber(*r.Whatsapp)
	if err == nil {
		r.Whatsapp = &num
		return nil
	}
	num, err = ValidateCellNumber(fmt.Sprintf("+%s", *r.Whatsapp))
	if err == nil {
		r.Whatsapp = &num
		return nil
	}
	return fmt.Errorf("invalid WhatsApp number format")
}

func (r *VendorUpdateRequest) validateEmail() error {
	if r.Email == nil {
		return nil
	}
	if *r.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	_, err := mail.ParseAddress(*r.Email)
	if err != nil {
		return fmt.Errorf("email '%s' is not in a valid format", *r.Email)
	}
	return nil
}

func (r *VendorUpdateRequest) validateAtLeastOneField() error {
	if r.LegalName == nil && r.TradingName == nil && r.Document == nil && r.TaxDocument == nil &&
		r.AccountBank == nil && r.AccountAgency == nil && r.AccountNumber == nil &&
		r.PixToken == nil && r.PixName == nil && r.PixCity == nil && r.LogoName == nil &&
		r.Email == nil && r.Whatsapp == nil && r.LastRps == nil && r.SmtpHost == nil &&
		r.SmtpPort == nil && r.SmtpUser == nil && r.SmtpPassword == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// GetDomain returns the updated domain.Vendor entity.
func (r *VendorUpdateRequest) GetDomain() interface{} {
	if r.vendor == nil {
		return nil
	}
	if r.LegalName != nil {
		r.vendor.LegalName = *r.LegalName
	}
	if r.TradingName != nil {
		r.vendor.TradingName = *r.TradingName
	}
	if r.Document != nil {
		r.vendor.Document = *r.Document
	}
	if r.TaxDocument != nil {
		r.vendor.TaxDocument = *r.TaxDocument
	}
	if r.AccountBank != nil {
		r.vendor.AccountBank = *r.AccountBank
	}
	if r.AccountAgency != nil {
		r.vendor.AccountAgency = *r.AccountAgency
	}
	if r.AccountNumber != nil {
		r.vendor.AccountNumber = *r.AccountNumber
	}
	if r.PixToken != nil {
		r.vendor.PixToken = *r.PixToken
	}
	if r.PixName != nil {
		r.vendor.PixName = *r.PixName
	}
	if r.PixCity != nil {
		r.vendor.PixCity = *r.PixCity
	}
	if r.LogoName != nil {
		r.vendor.LogoName = *r.LogoName
	}
	if r.Email != nil {
		r.vendor.Email = *r.Email
	}
	if r.Whatsapp != nil {
		r.vendor.Whatsapp = *r.Whatsapp
	}
	if r.LastRps != nil {
		r.vendor.LastRps = *r.LastRps
	}
	if r.SmtpHost != nil {
		r.vendor.SmtpHost = *r.SmtpHost
	}
	if r.SmtpPort != nil {
		r.vendor.SmtpPort = *r.SmtpPort
	}
	if r.SmtpUser != nil {
		r.vendor.SmtpUser = *r.SmtpUser
	}
	if r.SmtpPassword != nil {
		r.vendor.SmtpPassword = *r.SmtpPassword
	}
	r.vendor.UpdatedAt = time.Now()
	return r.vendor
}

// Reset resets the fields of VendorUpdateRequest.
func (r *VendorUpdateRequest) Reset() {
	r.Nickname = ""
	r.LegalName = nil
	r.TradingName = nil
	r.Document = nil
	r.TaxDocument = nil
	r.AccountBank = nil
	r.AccountAgency = nil
	r.AccountNumber = nil
	r.PixToken = nil
	r.PixName = nil
	r.PixCity = nil
	r.LogoName = nil
	r.Email = nil
	r.Whatsapp = nil
	r.LastRps = nil
	r.SmtpHost = nil
	r.SmtpPort = nil
	r.SmtpUser = nil
	r.SmtpPassword = nil
	r.vendor = nil
}
