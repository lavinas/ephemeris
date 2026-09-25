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

// VendorCreateRequest represents the request payload for creating a new vendor.
type VendorCreateRequest struct {
	Nickname      string `json:"nickname" validate:"required"`
	LegalName     string `json:"legal_name" validate:"required"`
	TradingName   string `json:"trading_name"`
	Document      string `json:"document" validate:"required"`
	TaxDocument   string `json:"tax_document"`
	AccountBank   string `json:"account_bank"`
	AccountAgency string `json:"account_agency"`
	AccountNumber string `json:"account_number"`
	PixToken      string `json:"pix_token"`
	PixName       string `json:"pix_name"`
	PixCity       string `json:"pix_city"`
	LogoName      string `json:"logo_name"`
	Email         string `json:"email" validate:"required,email"`
	Whatsapp      string `json:"whatsapp" validate:"required"`
	LastRps       int64  `json:"last_rps"`
	SmtpHost      string `json:"smtp_host"`
	SmtpPort      int    `json:"smtp_port"`
	SmtpUser      string `json:"smtp_user"`
	SmtpPassword  string `json:"smtp_password"`
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
	if err := r.validateEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateWhatsapp(); err != nil {
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
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	existing, err := repo.FindVendors(0, 0, nil, nil, &r.Document, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("document '%s' is already in use by another vendor", r.Document)
	}
	return nil
}

func (r *VendorCreateRequest) validateCpfCnpj() error {
	cpf := cpfcnpj.NewCPF(r.Document)
	if cpf.IsValid() {
		r.Document = cpf.String()
		return nil
	}
	cnpj := cpfcnpj.NewCNPJ(r.Document)
	if cnpj.IsValid() {
		r.Document = cnpj.String()
		return nil
	}
	return fmt.Errorf("invalid document format")
}

func (r *VendorCreateRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (r *VendorCreateRequest) validateWhatsapp() error {
	if r.Whatsapp == "" {
		return nil
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
