package dto

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"signup/internal/domain"
	"signup/internal/port"
)

const (
	// requestLimit defines the maximum number of customer creation
	// 	requests that can be processed in a single batch.
	requestLimit = 100
)

// CustomerCreateRequest represents the request payload for creating a new customer.
type CustomerCreateRequest struct {
	RequestBase `json:"-" validate:"-"`
	Vendor      string `json:"vendor" validate:"required"`
	VendorID    int64  `json:"-" validate:"-"`
	Name        string `json:"name" validate:"required"`
	Nickname    string `json:"nickname" validate:"required"`
	Document    string `json:"document" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	Whatsapp    string `json:"whatsapp" validate:"required"`
}

// NewCustomerCreateRequest creates a new instance of CustomerCreateRequest
func NewCustomerCreateRequest(repo port.Repository) *CustomerCreateRequest {
	return &CustomerCreateRequest{
		RequestBase: NewRequestBase(repo),
	}
}

// CustomerCreateResponse represents the response payload after creating a new customer.
type CustomerCreateResponse struct {
	ResponseBase
}

// NewCustomerCreateResponse creates a new instance of CustomerCreateResponse
func NewCustomerCreateResponse(httpCode int, status, message string) CustomerCreateResponse {
	return CustomerCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the CustomerCreateRequest has all required fields and valid data.
func (r *CustomerCreateRequest) Validate() error {
	errs := make([]error, 0)
	if err := r.validateVendor(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateName(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(); err != nil {
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

// GetDomain converts the CustomerCreateRequestItem to a domain.Customer entity.
func (r *CustomerCreateRequest) GetDomain() (port.Domain, error) {
	var document, email, whatsapp *string
	if r.Document != "" {
		document = &r.Document
	}
	if r.Email != "" {
		email = &r.Email
	}
	if r.Whatsapp != "" {
		whatsapp = &r.Whatsapp
	}
	return domain.NewCustomer(r.Repo, r.VendorID, &r.Name, &r.Nickname, document, email, whatsapp), nil
}

// GetOutDTO converts the CustomerCreateRequest to a CustomerCreateResponse DTO.
func (r *CustomerCreateRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	return &CustomerCreateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// GetPageParams returns the pagination parameters.
func (r *CustomerCreateRequest) GetPageParams() (int, int) {
	return 1, 1
}

// validateVendor checks if the provided vendor is valid and exists in the system.
func (r *CustomerCreateRequest) validateVendor() error {
	if r.Vendor == "" {
		return errors.New("vendor is required")
	}
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil {
		return fmt.Errorf("failed to validate vendor: %v", err)
	} else if !ok {
		return fmt.Errorf("vendor '%s' does not exist", r.Vendor)
	}
	r.VendorID = vendor.ID
	return nil
}

// validateName checks if the provided name is valid.
func (r *CustomerCreateRequest) validateName() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// validateNickname checks if the provided nickname is valid and not already in use.
func (r *CustomerCreateRequest) validateNickname() error {
	if r.Nickname == "" {
		return errors.New("nickname is required")
	}
	if r.Nickname != strings.ToLower(r.Nickname) || strings.Contains(r.Nickname, " ") {
		return errors.New("nickname must be lowercase and must not contain spaces")
	}
	c := domain.StartCustomer(r.Repo)
	if ok, err := c.GetByNickname(r.VendorID, r.Nickname); err != nil {
		return fmt.Errorf("failed to validate nickname: %v", err)
	} else if ok {
		return fmt.Errorf("nickname is already in use")
	}
	return nil
}

// validateDocument checks if the provided document is valid and not already in use.
func (r *CustomerCreateRequest) validateDocument() error {
	if r.Document == "" {
		return nil
	}
	doc, err := ValidateDocument(r.Document, "all")
	if err != nil {
		return err
	}
	r.Document = doc
	c := domain.StartCustomer(r.Repo)
	if ok, err := c.GetByDocument(r.VendorID, r.Document); err != nil {
		return fmt.Errorf("failed to validate document: %v", err)
	} else if ok {
		return fmt.Errorf("document is already in use")
	}
	return nil
}

// validateEmail checks if the provided email is valid and not already in use.
func (r *CustomerCreateRequest) validateEmail() error {
	if r.Email == "" {
		return nil
	}
	if err := ValidateEmail(r.Email); err != nil {
		return err
	}
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is valid and not already in use.
func (r *CustomerCreateRequest) validateWhatsapp() error {
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

// EmitEvent publishes the customer created event using the provided publisher.
func (r *CustomerCreateRequest) EmitEvent(ctx context.Context, publisher port.CustomerEventPublisher) error {
	if publisher == nil {
		return nil
	}
	var doc, email, whatsapp *string
	if r.Document != "" {
		doc = &r.Document
	}
	if r.Email != "" {
		email = &r.Email
	}
	if r.Whatsapp != "" {
		whatsapp = &r.Whatsapp
	}
	status := 1
	data := port.CustomerEventData{
		Name:     r.Name,
		Nickname: r.Nickname,
		Document: doc,
		Email:    email,
		Whatsapp: whatsapp,
		Status:   &status,
	}
	return publisher.PublishCustomerCreated(ctx, r.Vendor, data)
}
