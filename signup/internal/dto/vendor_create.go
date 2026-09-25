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

// validateTaxDocument checks if the tax document is provided.
func (r *VendorCreateRequest) validateTaxDocument() error {
	if r.TaxDocument == "" {
		return errors.New("tax_document is required")
	}
	return nil
}

// validateAccountBank checks if the bank account name/code is provided.
func (r *VendorCreateRequest) validateAccountBank() error {
	if r.AccountBank == "" {
		return errors.New("account_bank is required")
	}
	return nil
}

// validateAccountAgency checks if the bank agency is provided.
func (r *VendorCreateRequest) validateAccountAgency() error {
	if r.AccountAgency == "" {
		return errors.New("account_agency is required")
	}
	return nil
}

// validateAccountNumber checks if the bank account number is provided.
func (r *VendorCreateRequest) validateAccountNumber() error {
	if r.AccountNumber == "" {
		return errors.New("account_number is required")
	}
	return nil
}

// validatePixToken checks if the PIX token is provided.
func (r *VendorCreateRequest) validatePixToken() error {
	if r.PixToken == "" {
		return errors.New("pix_token is required")
	}
	return nil
}

// validatePixName checks if the PIX beneficiary name is provided.
func (r *VendorCreateRequest) validatePixName() error {
	if r.PixName == "" {
		return errors.New("pix_name is required")
	}
	return nil
}

// validatePixCity checks if the PIX city is provided.
func (r *VendorCreateRequest) validatePixCity() error {
	if r.PixCity == "" {
		return errors.New("pix_city is required")
	}
	return nil
}

// validateLogoName checks if the logo name is provided.
func (r *VendorCreateRequest) validateLogoName() error {
	if r.LogoName == "" {
		return errors.New("logo_name is required")
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
		return errors.New("email is required")
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is valid.
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

// validateLastRps checks if the last RPS number is valid.
func (r *VendorCreateRequest) validateLastRps() error {
	if r.LastRps < 0 {
		return errors.New("last_rps must be greater than or equal to zero")
	}
	return nil
}

// validateSmtpHost checks if the SMTP host is provided.
func (r *VendorCreateRequest) validateSmtpHost() error {
	if r.SmtpHost == "" {
		return errors.New("smtp_host is required")
	}
	return nil
}

// validateSmtpPort checks if the SMTP port is valid.
func (r *VendorCreateRequest) validateSmtpPort() error {
	if r.SmtpPort <= 0 {
		return errors.New("smtp_port is required and must be greater than zero")
	}
	return nil
}

// validateSmtpUser checks if the SMTP username is provided.
func (r *VendorCreateRequest) validateSmtpUser() error {
	if r.SmtpUser == "" {
		return errors.New("smtp_user is required")
	}
	return nil
}

// validateSmtpPassword checks if the SMTP password is provided.
func (r *VendorCreateRequest) validateSmtpPassword() error {
	if r.SmtpPassword == "" {
		return errors.New("smtp_password is required")
	}
	return nil
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
