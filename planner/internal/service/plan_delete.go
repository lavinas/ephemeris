package service

import (
	"planner/internal/dto"
	"planner/internal/port"
)

// PlanDelete is a service that handles the deletion of plans.
type PlanDelete struct {
	Base
}

// NewPlanDelete creates a new instance of PlanDelete.
func NewPlanDelete(repo port.Repository, logger port.Logger) *PlanDelete {
	return &PlanDelete{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the plan deletion process.
func (s *PlanDelete) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing delete plan request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation error: %v", err)
		return dto.NewPlanDeleteResponse(400, "Bad Request", err.Error())
	}

	planID := input.(*dto.PlanDeleteRequest).PlanID
	existing, err := s.repo.GetPlanByID(planID)
	if err != nil {
		s.logger.IPrintf(2, "Error checking plan existence: %v", err)
		return dto.NewPlanDeleteResponse(500, "Internal Server Error", err.Error())
	}
	if existing == nil {
		return dto.NewPlanDeleteResponse(404, "Not Found", "plano não encontrado")
	}

	if err := s.repo.DeletePlan(planID); err != nil {
		s.logger.IPrintf(2, "Error deleting plan: %v", err)
		return dto.NewPlanDeleteResponse(500, "Internal Server Error", err.Error())
	}

	s.logger.IPrintf(2, "Plan with ID %d deleted successfully", planID)
	return dto.NewPlanDeleteResponse(200, "OK", "plano excluído com sucesso")
}
