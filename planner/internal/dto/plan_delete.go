package dto

import (
	"errors"
	"planner/internal/port"
)

// PlanDeleteRequest represents a request to delete a plan
type PlanDeleteRequest struct {
	PlanID int64 `json:"id" validate:"required"`
}

// PlanDeleteResponse represents the response after deleting a plan
type PlanDeleteResponse struct {
	ResponseBase
}

// NewPlanDeleteResponse creates a new instance of PlanDeleteResponse
func NewPlanDeleteResponse(statusCode int, statusMessage string, errorMessage string) *PlanDeleteResponse {
	return &PlanDeleteResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// Validate checks if the PlanDeleteRequest has valid data
func (r *PlanDeleteRequest) Validate(repo port.Repository) error {
	if r.PlanID <= 0 {
		return errors.New("id do plano inválido: deve ser maior que zero")
	}
	return nil
}

// Reset resets the PlanDeleteRequest fields
func (r *PlanDeleteRequest) Reset() {
	r.PlanID = 0
}
