package dto

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"signup/internal/domain"
	"signup/internal/port"

)

// VendorUpdateRequest represents the request payload for updating an existing vendor.
type VendorUpdateRequest struct {
	RequestBase   `json:"-" validate:"-"`
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

// NewVendorUpdateRequest creates a new instance of VendorUpdateRequest.
func NewVendorUpdateRequest(repo port.Repository) *VendorUpdateRequest {
	return &VendorUpdateRequest{
		RequestBase: NewRequestBase(repo),
	}
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
func (r *VendorUpdateRequest) Validate() error {
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

// GetDomain returns the domain model of the vendor.
func (r *VendorUpdateRequest) GetDomain() (port.Domain, error) {
	if r.vendor == nil {
		return nil, errors.New("vendor not found")
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
	return r.vendor, nil
}

// GetOutDTO converts the VendorUpdateRequest to a VendorUpdateResponse DTO.
func (r *VendorUpdateRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	return &VendorUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// GetPageParams returns the pagination parameters.
func (r *VendorUpdateRequest) GetPageParams() (int, int) {
	return 1, 1
}

// validateNickname checks if the provided nickname exists.
func (r *VendorUpdateRequest) validateNickname() error {
	if r.Nickname == "" {
		return fmt.Errorf("nickname is required")
	}
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Nickname); err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	} else if !ok {
		return fmt.Errorf("vendor with nickname '%s' does not exist", r.Nickname)
	}
	r.vendor = vendor
	return nil
}

// validateLegalName checks if the provided legal name is not empty.
func (r *VendorUpdateRequest) validateLegalName() error {
	if r.LegalName == nil {
		return nil
	}
	if *r.LegalName == "" {
		return fmt.Errorf("legal_name cannot be empty")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not used by another vendor.
func (r *VendorUpdateRequest) validateDocument() error {
	if r.Document == nil {
		return nil
	}
	if *r.Document == "" {
		return fmt.Errorf("document cannot be empty")
	}
	doc, err := ValidateDocument(*r.Document, "cnpj")
	if err != nil {
		err = fmt.Errorf("%v (only cnpj is accepted)", err)
		return err
	}
	*r.Document = doc
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByDocument(*r.Document); err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	} else if !ok {
		return nil
	}
	if r.vendor != nil && vendor.ID == r.vendor.ID {
		return nil
	}
	return fmt.Errorf("document '%s' is already in use by another vendor", *r.Document)
}

// validateWhatsapp checks if the provided WhatsApp number is in a valid format.
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

// validateEmail checks if the provided email is in a valid format.
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

// validateAtLeastOneField checks if at least one of the fields is provided for update.
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

// EmitEvent publishes the vendor updated event using the provided publisher.
func (r *VendorUpdateRequest) EmitEvent(ctx context.Context, publisher port.CustomerEventPublisher) error {
	if publisher == nil || r.vendor == nil {
		return nil
	}
	data := port.VendorEventData{
		Nickname:      r.vendor.Nickname,
		LegalName:     r.vendor.LegalName,
		TradingName:   r.vendor.TradingName,
		Document:      r.vendor.Document,
		TaxDocument:   r.vendor.TaxDocument,
		AccountBank:   r.vendor.AccountBank,
		AccountAgency: r.vendor.AccountAgency,
		AccountNumber: r.vendor.AccountNumber,
		PixToken:      r.vendor.PixToken,
		PixName:       r.vendor.PixName,
		PixCity:       r.vendor.PixCity,
		LogoName:      r.vendor.LogoName,
		Email:         r.vendor.Email,
		Whatsapp:      r.vendor.Whatsapp,
		LastRps:       r.vendor.LastRps,
		SmtpHost:      r.vendor.SmtpHost,
		SmtpPort:      r.vendor.SmtpPort,
		SmtpUser:      r.vendor.SmtpUser,
		SmtpPassword:  r.vendor.SmtpPassword,
	}
	return publisher.PublishVendorUpdated(ctx, data)
}
