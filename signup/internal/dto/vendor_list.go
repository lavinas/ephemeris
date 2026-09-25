package dto

import (
	"errors"
	"fmt"
	"strings"

	"signup/internal/domain"
	"signup/internal/port"
)

// VendorListRequest represents the data transfer object for listing vendors with pagination.
type VendorListRequest struct {
	RequestBase `json:"-" validate:"-"`
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	Nickname    *string `json:"nickname,omitempty"`
	LegalName   *string `json:"legal_name,omitempty"`
	TradingName *string `json:"trading_name,omitempty"`
	Document    *string `json:"document,omitempty"`
	Email       *string `json:"email,omitempty"`
	Whatsapp    *string `json:"whatsapp,omitempty"`
}

// NewVendorListRequest creates a new instance of VendorListRequest with the provided repository.
func NewVendorListRequest(repo port.Repository) *VendorListRequest {
	return &VendorListRequest{
		RequestBase: NewRequestBase(repo),
	}
}

// VendorListResponse represents the data transfer object for the response after listing vendors.
type VendorListResponse struct {
	ResponseBase
	Vendors []VendorDTO `json:"vendors"`
}

// VendorDTO represents the data transfer object for a vendor in the list response.
type VendorDTO struct {
	ID            int64  `json:"id"`
	Nickname      string `json:"nickname"`
	LegalName     string `json:"legal_name"`
	TradingName   string `json:"trading_name"`
	Document      string `json:"document"`
	TaxDocument   string `json:"tax_document"`
	AccountBank   string `json:"account_bank"`
	AccountAgency string `json:"account_agency"`
	AccountNumber string `json:"account_number"`
	PixToken      string `json:"pix_token"`
	PixName       string `json:"pix_name"`
	PixCity       string `json:"pix_city"`
	LogoName      string `json:"logo_name"`
	Email         string `json:"email"`
	Whatsapp      string `json:"whatsapp"`
	LastRps       int64  `json:"last_rps"`
	SmtpHost      string `json:"smtp_host"`
	SmtpPort      int    `json:"smtp_port"`
	SmtpUser      string `json:"smtp_user"`
}

// NewVendorListResponse creates a new instance of VendorListResponse with the provided parameters.
func NewVendorListResponse(httpCode int, status, message string, vendors []VendorDTO) VendorListResponse {
	return VendorListResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
		Vendors:      vendors,
	}
}

// NewVendorDTO creates a new instance of VendorDTO from a domain.Vendor.
func NewVendorDTO(v *domain.Vendor) VendorDTO {
	return VendorDTO{
		ID:            v.ID,
		Nickname:      v.Nickname,
		LegalName:     v.LegalName,
		TradingName:   v.TradingName,
		Document:      v.Document,
		TaxDocument:   v.TaxDocument,
		AccountBank:   v.AccountBank,
		AccountAgency: v.AccountAgency,
		AccountNumber: v.AccountNumber,
		PixToken:      v.PixToken,
		PixName:       v.PixName,
		PixCity:       v.PixCity,
		LogoName:      v.LogoName,
		Email:         v.Email,
		Whatsapp:      v.Whatsapp,
		LastRps:       v.LastRps,
		SmtpHost:      v.SmtpHost,
		SmtpPort:      v.SmtpPort,
		SmtpUser:      v.SmtpUser,
	}
}

// Validate validates the VendorListRequest fields.
func (r *VendorListRequest) Validate() error {
	errs := make([]error, 0)
	if r.Page <= 0 {
		errs = append(errs, fmt.Errorf("page must be greater than 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, fmt.Errorf("page_size must be greater than 0"))
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

// GetDomain converts the VendorListRequest to a domain.Vendor entity.
func (r *VendorListRequest) GetDomain() (port.Domain, error) {
	vendor := domain.StartVendor(r.Repo)
	if r.Nickname != nil {
		vendor.Nickname = *r.Nickname
	}
	if r.LegalName != nil {
		vendor.LegalName = *r.LegalName
	}
	if r.TradingName != nil {
		vendor.TradingName = *r.TradingName
	}
	if r.Document != nil {
		vendor.Document = *r.Document
	}
	if r.Email != nil {
		vendor.Email = *r.Email
	}
	if r.Whatsapp != nil {
		vendor.Whatsapp = *r.Whatsapp
	}
	return vendor, nil
}

// GetOutDTO constructs an output DTO for the vendor list request.
func (r *VendorListRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	vendors := make([]VendorDTO, 0, 10)
	if list, ok := data.([]port.Domain); ok {
		for _, d := range list {
			if v, ok := d.(*domain.Vendor); ok {
				vendors = append(vendors, NewVendorDTO(v))
			}
		}
		return NewVendorListResponse(httpCode, status, message, vendors)
	}
	return NewVendorListResponse(httpCode, status, message, []VendorDTO{})
}

// GetPageParams returns the pagination parameters.
func (r *VendorListRequest) GetPageParams() (int, int) {
	return r.Page, r.PageSize
}
