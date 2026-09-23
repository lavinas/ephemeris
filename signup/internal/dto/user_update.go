package dto

import (
	"errors"
	"fmt"

	"signup/internal/domain"
	"signup/internal/port"
)

// UserUpdateRequest represents the request payload for updating a user's information.
type UserUpdateRequest struct {
	RequestBase `json:"-" validate:"-"`
	Vendor      string       `json:"vendor,omitempty"`
	VendorID    int64        `json:"-" validate:"-"`
	Username    string       `json:"username,omitempty"`
	Password    *string      `json:"password,omitempty"`
	Name        *string      `json:"name,omitempty"`
	Email       *string      `json:"email,omitempty"`
	Whatsapp    *string      `json:"whatsapp,omitempty"`
	Status      *int         `json:"status,omitempty"`
	user        *domain.User `json:"-"`
}

// NewUserUpdateRequest creates a new UserUpdateRequest.
func NewUserUpdateRequest(repo port.Repository) *UserUpdateRequest {
	return &UserUpdateRequest{
		RequestBase: NewRequestBase(repo),
	}
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
func (r *UserUpdateRequest) Validate() error {
	if err := r.validateVendor(); err != nil {
		return err
	}
	if err := r.validateUsername(); err != nil {
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

// GetDomain returns the domain entity.
func (r *UserUpdateRequest) GetDomain() (port.Domain, error) {
	if r.user == nil {
		return nil, errors.New("user not found")
	}
	if r.Name != nil {
		r.user.Name = *r.Name
	}
	if r.Password != nil {
		r.user.PassHash = *r.Password
	}
	if r.Email != nil {
		r.user.Email = r.Email
	}
	if r.Whatsapp != nil {
		r.user.Whatsapp = r.Whatsapp
	}
	if r.Status != nil {
		r.user.Status = r.Status
	}
	return r.user, nil
}

// GetOutDTO converts the CustomerCreateRequest to a CustomerCreateResponse DTO.
func (r *UserUpdateRequest) GetOutDTO(httpCode int, status, message string, data interface{}) port.OutDTO {
	return &UserUpdateResponse{
		ResponseBase: NewResponseBase(httpCode, status, message),
	}
}

// GetPageParams returns the pagination parameters.
func (r *UserUpdateRequest) GetPageParams() (int, int) {
	return 1, 1
}

// validateVendor checks if the vendor field is valid.
func (r *UserUpdateRequest) validateVendor() error {
	vendor := domain.StartVendor(r.Repo)
	if ok, err := vendor.GetByNickname(r.Vendor); err != nil {
		return fmt.Errorf("invalid vendor: %w", err)
	} else if !ok {
		return errors.New("vendor not found")
	}
	r.VendorID = vendor.ID
	return nil
}

// validateUsername checks if the username field is valid.
func (r *UserUpdateRequest) validateUsername() error {
	user := domain.StartUser(r.Repo)
	if ok, err := user.GetByUsername(r.VendorID, r.Username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	} else if !ok {
		return errors.New("username not found")
	}
	r.user = user
	return nil
}

// validatePassword checks if the provided password is not empty and has a valid format.
func (r *UserUpdateRequest) validatePassword() error {
	if r.Password == nil {
		return nil
	}
	if *r.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if len(*r.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	var err error
	hashedPass, err := HashPassword(*r.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	r.Password = &hashedPass
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
