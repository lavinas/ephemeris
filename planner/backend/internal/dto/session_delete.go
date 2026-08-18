package dto

import (
	"errors"
	"planner/internal/port"
)

// SessionDeleteRequest represents a request to delete a session.
type SessionDeleteRequest struct {
	SessionID int64 `json:"id" validate:"required"`
}

// SessionDeleteResponse represents the response after deleting a session.
type SessionDeleteResponse struct {
	ResponseBase
}

// NewSessionDeleteResponse creates a new instance of SessionDeleteResponse with the provided parameters.
func NewSessionDeleteResponse(statusCode int, statusMessage string, errorMessage string) *SessionDeleteResponse {
	return &SessionDeleteResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// Validate checks if the SessionDeleteRequest has valid data.
func (r *SessionDeleteRequest) Validate(repo port.Repository) error {
	if r.SessionID <= 0 {
		return errors.New("invalid session_id: must be a positive integer")
	}
	return nil
}

// Reset resets the SessionDeleteRequest fields to their zero values.
func (r *SessionDeleteRequest) Reset() {
	r.SessionID = 0
}