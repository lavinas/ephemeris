package dto

import (
	"fmt"
	"strings"

	"planner/internal/port"
)

// ServiceDTO represents a service item returned by the session services use case.
type ServiceDTO struct {
	ID             int64   `json:"id"`
	VendorID       int64   `json:"vendor_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	SessionMinutes *int    `json:"session_minutes,omitempty"`
}

// SessionServicesRequest represents the input DTO for retrieving suggested services for a session/customer.
type SessionServicesRequest struct {
	VendorID int64  `json:"vendor_id,omitempty"`
	Nickname string `json:"nickname,omitempty"`
}

// SessionServicesResponse represents the output DTO containing suggested services.
type SessionServicesResponse struct {
	ResponseBase
	Services []ServiceDTO `json:"services"`
}

// NewSessionServicesResponse creates a new SessionServicesResponse instance.
func NewSessionServicesResponse(statusCode int, statusMessage string, errorMessage string, services []ServiceDTO) *SessionServicesResponse {
	return &SessionServicesResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Services: services,
	}
}

// Validate validates and normalizes the SessionServicesRequest fields.
func (r *SessionServicesRequest) Validate(repo port.Repository) error {
	if r.VendorID < 0 {
		return fmt.Errorf("invalid vendor_id: %d", r.VendorID)
	}
	if r.VendorID == 0 {
		r.VendorID = 1
	}
	r.Nickname = strings.TrimSpace(r.Nickname)
	return nil
}

