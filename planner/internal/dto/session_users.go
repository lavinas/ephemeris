package dto

import (
	"fmt"
	"time"

	"planner/internal/port"
)

// SessionUsersRequest represents the request DTO for retrieving session users within a specified date range.
type SessionUsersRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Status    string `json:"status"`
	Service   string `json:"service"`
	Minutes   int    `json:"minutes"`
}

// SessionUsersResponse represents the response DTO for session users retrieval, including status, message, and data.
type SessionUsersResponse struct {
	ResponseBase
	Nicknames []string `json:"nicknames"`
}

// NewSessionUsersResponse creates a new instance of SessionUsersResponse with the provided parameters.
func NewSessionUsersResponse(statusCode int, statusMessage string, errorMessage string, nicknames []string) *SessionUsersResponse {
	return &SessionUsersResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Nicknames: nicknames,
	}
}

// Validate checks if the provided status and service are valid based on predefined valid values.
func (r *SessionUsersRequest) Validate(repo port.Repository) error {
	if r.StartDate != "" {
		if _, err := time.Parse("2006-01-02", r.StartDate); err != nil {
			return fmt.Errorf("invalid start date: %s", r.StartDate)
		}
	}
	if r.EndDate != "" {
		if _, err := time.Parse("2006-01-02", r.EndDate); err != nil {
			return fmt.Errorf("invalid end date: %s", r.EndDate)
		}
	}
	if r.Status != "" && !validStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if r.Service != "" && !validServices[r.Service] {
		return fmt.Errorf("invalid service: %s", r.Service)
	}
	if r.Minutes < 0 {
		return fmt.Errorf("invalid minutes: %d", r.Minutes)
	}
	return nil
}

// Reset resets the SessionUsersRequest fields to their zero values.
func (r *SessionUsersRequest) Reset() {
	// No fields to reset in SessionUsersRequest
}
