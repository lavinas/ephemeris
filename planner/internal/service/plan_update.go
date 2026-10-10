package service

import (
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// PlanUpdate is a service that handles updating existing plans.
type PlanUpdate struct {
	Base
}

// NewPlanUpdate creates a new instance of PlanUpdate.
func NewPlanUpdate(repo port.Repository, logger port.Logger) *PlanUpdate {
	return &PlanUpdate{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the plan update process using the provided input DTO.
func (s *PlanUpdate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing update plan request: %v", input)
	req := input.(*dto.PlanUpdateRequest)
	if err := req.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewPlanUpdateResponse(400, "Bad Request", err.Error())
	}

	planModel := req.GetPlan()
	s.applyUpdates(planModel, req)

	if err := s.repo.SavePlan(planModel); err != nil {
		s.logger.IPrintf(2, "Failed to save updated plan: %v", err)
		return dto.NewPlanUpdateResponse(500, "error", err.Error())
	}

	s.logger.IPrintf(2, "Successfully updated plan with ID: %d", planModel.ID)
	return dto.NewPlanUpdateResponse(200, "success", "plano atualizado com sucesso")
}

func (s *PlanUpdate) applyUpdates(plan *domain.Plan, req *dto.PlanUpdateRequest) {
	if req.CustomerID > 0 {
		plan.CustomerID = req.CustomerID
	}
	if req.PlanStart != "" {
		if st, err := time.Parse("2006-01-02", req.PlanStart); err == nil {
			plan.PlanStart = st
		}
	}
	if req.ClearPlanEnd {
		plan.PlanEnd = nil
	} else if req.PlanEnd != nil && *req.PlanEnd != "" {
		if et, err := time.Parse("2006-01-02", *req.PlanEnd); err == nil {
			plan.PlanEnd = &et
		}
	}
	if req.PlanType > 0 {
		plan.PlanType = req.PlanType
	}
	if req.ClearPrice {
		plan.Price = nil
	} else if req.Price != nil {
		plan.Price = req.Price
	}

	if len(req.Items) > 0 {
		items := make([]domain.PlanItem, len(req.Items))
		for i, it := range req.Items {
			orderIdx := it.OrderIndex
			if orderIdx <= 0 {
				orderIdx = i + 1
			}
			items[i] = domain.PlanItem{
				PlanID:     plan.ID,
				ServiceID:  it.ServiceID,
				OrderIndex: orderIdx,
				Price:      it.Price,
			}
		}
		plan.Items = items
	}

	switch plan.PlanType {
	case int(domain.PlanTypeAgenda):
		if req.Agenda != nil {
			plan.Agenda = &domain.PlanAgenda{
				ID:                  plan.ID,
				Recurrence:          req.Agenda.Recurrence,
				WeekDay:             req.Agenda.WeekDay,
				TermType:            req.Agenda.TermType,
				PaymentDay:          req.Agenda.PaymentDay,
				MonthlyServiceLimit: req.Agenda.MonthlyServiceLimit,
			}
		}
	case int(domain.PlanTypePackage):
		if req.Package != nil {
			pDate, _ := time.Parse("2006-01-02", req.Package.PaymentDate)
			plan.Package = &domain.PlanPackage{
				ID:          plan.ID,
				Quantity:    req.Package.Quantity,
				PaymentDate: pDate,
			}
		}
	case int(domain.PlanTypeNotebook):
		if req.Notebook != nil {
			plan.Notebook = &domain.PlanNotebook{
				ID:          plan.ID,
				PaymentDay:  req.Notebook.PaymentDay,
				PaymentTerm: req.Notebook.PaymentTerm,
			}
		}
	}
}
