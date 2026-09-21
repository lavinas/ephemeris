package dto

import (
	"errors"
	"fmt"
	"strings"

	"signup/internal/port"
	"signup/internal/domain"
)

// UserUpdateRequest represents the request payload for updating a user's information.
type UserUpdateRequest struct {
	Vendor   string `json:"vendor,omitempty"`
	VendorID int64  `json:"-" validate:"-"`
	Username string `json:"username,omitempty"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Whatsapp *string `json:"whatsapp,omitempty"`
	Status   *int    `json:"status,omitempty"`
}

// UserUpdateResponse represents the response payload for updating a user's information.
type UserUpdateResponse struct {
	ResponseBase
}

// NewUserUpdateResponse creates a new UserUpdateResponse with the given response base.
func NewUserUpdateResponse(httpCode int, status, message string) UserUpdateResponse {
	return UserUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// Validate checks if the UserUpdateRequest has valid fields.
func (r *UserUpdateRequest) Validate(repo port.Repository) error {
	if err := r.validateVendor(repo); err != nil {
		return err
	}
	if err := r.validateUsername(repo); err != nil {
		return err
	}
	if err := r.validateName(); err != nil {
		return err
	}
	if err := r.validateEmail(); err != nil {
		return err
	}
	if err := r.validateWhatsapp(); err != nil {
		return err
	}
	if err := r.validateStatus(); err != nil {
		return err
	}
	return nil
}

// validateVendor checks if the vendor field is valid.
func (r *UserUpdateRequest) validateVendor(repo port.Repository) error {
	if strings.TrimSpace(r.Vendor) == "" {
		return errors.New("vendor cannot be empty")
	}
	var vendor domain.Vendor
	if ok, err := vendor.GetByNickname(repo, r.Vendor); err != nil {
		return fmt.Errorf("invalid vendor: %w", err)
	} else if !ok {
		return errors.New("vendor not found")
	}
	r.VendorID = vendor.ID
	return nil
}

// validateUsername checks if the username field is valid.
func (r *UserUpdateRequest) validateUsername(repo port.Repository) error {
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("username cannot be empty")
	}
	var user domain.User
	if ok, err := user.GetByUsername(repo, r.VendorID, r.Username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	} else if !ok {
		return errors.New("username not found")
	}
	return nil
}

// validateName checks if the provided name is not empty.
func (r *UserUpdateRequest) validateName() error {
	if r.Name == nil {
		return nil
	}
	if *r.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return nil
}

// validateEmail checks if the provided email is not empty and has a valid format.
func (r *UserUpdateRequest) validateEmail() error {
	if r.Email == nil {
		return nil
	}
	if *r.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if err := ValidateEmail(*r.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	// Add more email format validation if needed
	return nil
}

// validateWhatsapp checks if the provided WhatsApp number is not empty and has a valid format.
func (r *UserUpdateRequest) validateWhatsapp() error {
	if r.Whatsapp == nil {
		return nil
	}
	if *r.Whatsapp == "" {
		return fmt.Errorf("whatsapp cannot be empty")
	}
	newNumber, err := ValidateCellNumber(*r.Whatsapp)
	if err != nil {
		return fmt.Errorf("invalid whatsapp format: %w", err)
	}
	r.Whatsapp = &newNumber
	return nil
}

// validateStatus checks if the provided status is either 0 (inactive) or 1 (active).
func (r *UserUpdateRequest) validateStatus() error {
	if r.Status != nil && *r.Status != 0 && *r.Status != 1 {
		return fmt.Errorf("status must be either 0 (inactive) or 1 (active)")
	}
	return nil
}
