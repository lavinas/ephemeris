package dto

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/klassmann/cpfcnpj"
	"signup/internal/domain"
	"signup/internal/port"
)

// VendorCreateRequest represents the request payload for creating a new vendor.
type VendorCreateRequest struct {
	RequestBase   `json:"-" validate:"-"`
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

// NewVendorCreateRequest creates a new instance of VendorCreateRequest.
func NewVendorCreateRequest(repo port.Repository) *VendorCreateRequest {
	return &VendorCreateRequest{
		RequestBase: NewRequestBase(repo),
	}
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
func (r *VendorCreateRequest) Validate() error {
	errs := make([]error, 0)
	if err := r.validateNickname(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateLegalName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDocument(); err != nil {
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

// GetDomain converts the VendorCreateRequest to a domain.Vendor entity.
func (r *VendorCreateRequest) GetDomain() (port.Domain, error) {
	vendor := domain.NewVendor(
		r.Repo, r.Nickname, r.LegalName, r.TradingName,
		r.Document, r.TaxDocument, r.AccountBank, r.AccountAgency, r.AccountNumber,
		r.PixToken, r.PixName, r.PixCity, r.LogoName, r.Email, r.Whatsapp,
		r.SmtpHost, r.SmtpPort, r.SmtpUser, r.SmtpPassword,
	)
	vendor.LastRps = r.LastRps
	return vendor, nil
}

// GetOutDTO converts the VendorCreateRequest to a VendorCreateResponse DTO.
func (r *VendorCreateRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	return &VendorCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// GetPageParams returns the pagination parameters.
func (r *VendorCreateRequest) GetPageParams() (int, int) {
	return 1, 1
}

// validateNickname checks if the provided nickname is valid and not already in use.
func (r *VendorCreateRequest) validateNickname() error {
	if r.Nickname == "" {
		return errors.New("nickname is required")
	}
	if r.Nickname != strings.ToLower(r.Nickname) || strings.Contains(r.Nickname, " ") {
		return errors.New("nickname must be lowercase and must not contain spaces")
	}
	v := domain.StartVendor(r.Repo)
	if ok, err := v.GetByNickname(r.Nickname); err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	} else if ok {
		return fmt.Errorf("nickname is already in use")
	}
	return nil
}

// validateLegalName checks if the provided legal name is valid.
func (r *VendorCreateRequest) validateLegalName() error {
	if r.LegalName == "" {
		return errors.New("legal_name is required")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not already in use.
func (r *VendorCreateRequest) validateDocument() error {
	if r.Document == "" {
		return errors.New("document is required")
	}
	if err := r.validateCpfCnpj(); err != nil {
		return err
	}
	v := domain.StartVendor(r.Repo)
	if ok, err := v.GetByDocument(r.Document); err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	} else if ok {
		return fmt.Errorf("document is already in use")
	}
	return nil
}

// validateCpfCnpj checks if the provided document is a valid CPF or CNPJ.
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

// validateEmail checks if the provided email is valid.
func (r *VendorCreateRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is valid.
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

// EmitEvent publishes the vendor created event using the provided publisher.
func (r *VendorCreateRequest) EmitEvent(ctx context.Context, publisher port.CustomerEventPublisher) error {
	if publisher == nil {
		return nil
	}
	data := port.VendorEventData{
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
	}
	return publisher.PublishVendorCreated(ctx, data)
}
