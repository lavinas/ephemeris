package dto

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// VendorCreateRequest represents the request payload for creating a new vendor.
type VendorCreateRequest struct {
	Nickname      string `json:"nickname" validate:"required"`
	LegalName     string `json:"legal_name" validate:"required"`
	TradingName   string `json:"trading_name"`
	Document      string `json:"document" validate:"required"`
	TaxDocument   string `json:"tax_document" validate:"required"`
	AccountBank   string `json:"account_bank" validate:"required"`
	AccountAgency string `json:"account_agency" validate:"required"`
	AccountNumber string `json:"account_number" validate:"required"`
	PixToken      string `json:"pix_token" validate:"required"`
	PixName       string `json:"pix_name" validate:"required"`
	PixCity       string `json:"pix_city" validate:"required"`
	LogoName      string `json:"logo_name" validate:"required"`
	Email         string `json:"email" validate:"required,email"`
	Whatsapp      string `json:"whatsapp" validate:"required"`
	LastRps       int64  `json:"last_rps" validate:"gte=0"`
	SmtpHost      string `json:"smtp_host" validate:"required"`
	SmtpPort      int    `json:"smtp_port" validate:"required,gt=0"`
	SmtpUser      string `json:"smtp_user" validate:"required"`
	SmtpPassword  string `json:"smtp_password" validate:"required"`
}

// VendorCreateResponse represents the response payload after creating a new vendor.
type VendorCreateResponse struct {
	ResponseBase
}

// NewVendorCreateResponse creates a new instance of VendorCreateResponse.
func NewVendorCreateResponse(httpCode int, status, message string) VendorCreateResponse {
	return VendorCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the VendorCreateRequest has all required fields and valid data.
func (r *VendorCreateRequest) Validate(repo port.Repository) error {
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
	if err := r.validateTaxDocument(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAccountBank(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAccountAgency(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateAccountNumber(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePixToken(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePixName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePixCity(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateLogoName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateLastRps(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSmtpHost(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSmtpPort(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSmtpUser(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSmtpPassword(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *VendorCreateRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		return errors.New("nickname is required")
	}
	if r.Nickname != strings.ToLower(r.Nickname) || strings.Contains(r.Nickname, " ") {
		return errors.New("nickname must be lowercase and must not contain spaces")
	}
	existing, err := repo.GetVendor(r.Nickname)
	if err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	}
	if existing != nil {
		return fmt.Errorf("vendor '%s' already exists", r.Nickname)
	}
	return nil
}

func (r *VendorCreateRequest) validateLegalName() error {
	if r.LegalName == "" {
		return errors.New("legal_name is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateDocument(repo port.Repository) error {
	if r.Document == "" {
		return errors.New("document is required")
	}
	doc, err := ValidateDocument(r.Document, "cnpj")
	if err != nil {
		err = fmt.Errorf("%v (only cnpj is accepted)", err)
		return err
	}
	r.Document = doc
	existing, err := repo.FindVendors(0, 0, nil, nil, &r.Document, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("document '%s' is already in use by another vendor", r.Document)
	}
	return nil
}

func (r *VendorCreateRequest) validateTaxDocument() error {
	if r.TaxDocument == "" {
		return errors.New("tax_document is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateAccountBank() error {
	if r.AccountBank == "" {
		return errors.New("account_bank is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateAccountAgency() error {
	if r.AccountAgency == "" {
		return errors.New("account_agency is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateAccountNumber() error {
	if r.AccountNumber == "" {
		return errors.New("account_number is required")
	}
	return nil
}

func (r *VendorCreateRequest) validatePixToken() error {
	if r.PixToken == "" {
		return errors.New("pix_token is required")
	}
	return nil
}

func (r *VendorCreateRequest) validatePixName() error {
	if r.PixName == "" {
		return errors.New("pix_name is required")
	}
	return nil
}

func (r *VendorCreateRequest) validatePixCity() error {
	if r.PixCity == "" {
		return errors.New("pix_city is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateLogoName() error {
	if r.LogoName == "" {
		return errors.New("logo_name is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateEmail() error {
	if r.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (r *VendorCreateRequest) validateWhatsapp() error {
	if r.Whatsapp == "" {
		return errors.New("whatsapp is required")
	}
	num, err := ValidateCellNumber(r.Whatsapp)
	if err == nil {
		r.Whatsapp = num
		return nil
	}
	num, err = ValidateCellNumber(fmt.Sprintf("+%s", r.Whatsapp))
	if err == nil {
		r.Whatsapp = num
		return nil
	}
	return fmt.Errorf("invalid WhatsApp number format")
}

func (r *VendorCreateRequest) validateLastRps() error {
	if r.LastRps < 0 {
		return errors.New("last_rps must be greater than or equal to zero")
	}
	return nil
}

func (r *VendorCreateRequest) validateSmtpHost() error {
	if r.SmtpHost == "" {
		return errors.New("smtp_host is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateSmtpPort() error {
	if r.SmtpPort <= 0 {
		return errors.New("smtp_port is required and must be greater than zero")
	}
	return nil
}

func (r *VendorCreateRequest) validateSmtpUser() error {
	if r.SmtpUser == "" {
		return errors.New("smtp_user is required")
	}
	return nil
}

func (r *VendorCreateRequest) validateSmtpPassword() error {
	if r.SmtpPassword == "" {
		return errors.New("smtp_password is required")
	}
	return nil
}

// GetDomain converts the VendorCreateRequest to a domain.Vendor entity.
func (r *VendorCreateRequest) GetDomain() interface{} {
	return &domain.Vendor{
		Nickname:      r.Nickname,
		LegalName:     r.LegalName,
		TradingName:   r.TradingName,
		Document:      r.Document,
		TaxDocument:   r.TaxDocument,
		AccountBank:   r.AccountBank,
		AccountAgency: r.AccountAgency,
		AccountNumber: r.AccountNumber,
		PixToken:      r.PixToken,
		PixName:       r.PixName,
		PixCity:       r.PixCity,
		LogoName:      r.LogoName,
		Email:         r.Email,
		Whatsapp:      r.Whatsapp,
		LastRps:       r.LastRps,
		SmtpHost:      r.SmtpHost,
		SmtpPort:      r.SmtpPort,
		SmtpUser:      r.SmtpUser,
		SmtpPassword:  r.SmtpPassword,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Reset resets the fields of VendorCreateRequest.
func (r *VendorCreateRequest) Reset() {
	r.Nickname = ""
	r.LegalName = ""
	r.TradingName = ""
	r.Document = ""
	r.TaxDocument = ""
	r.AccountBank = ""
	r.AccountAgency = ""
	r.AccountNumber = ""
	r.PixToken = ""
	r.PixName = ""
	r.PixCity = ""
	r.LogoName = ""
	r.Email = ""
	r.Whatsapp = ""
	r.LastRps = 0
	r.SmtpHost = ""
	r.SmtpPort = 0
	r.SmtpUser = ""
	r.SmtpPassword = ""
}
