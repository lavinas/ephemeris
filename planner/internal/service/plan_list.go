package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// PlanList is a service that handles retrieving a list of plans.
type PlanList struct {
	Base
}

// NewPlanList creates a new instance of PlanList.
func NewPlanList(repo port.Repository, logger port.Logger) *PlanList {
	return &PlanList{
		Base: *NewBase(repo, logger),
	}
}

// Run executes the plan list retrieval process using the provided input DTO.
func (s *PlanList) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing list plans request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewPlanListResponse(400, "error", err.Error(), nil, 0, 0, 0, 0)
	}

	req := input.(*dto.PlanListRequest)
	conditions := map[string]interface{}{}
	conditions["deleted_at IS NULL"] = nil

	if req.PlanID != 0 {
		conditions["id = ?"] = req.PlanID
	}
	if len(req.CustomerIDs) == 1 {
		conditions["customer_id = ?"] = req.CustomerIDs[0]
	} else if len(req.CustomerIDs) > 1 {
		conditions["customer_id IN (?)"] = req.CustomerIDs
	}
	if req.PlanType != 0 {
		conditions["plan_type = ?"] = req.PlanType
	}
	if req.DateStart != "" {
		if dStart, err := time.Parse("2006-01-02", req.DateStart); err == nil {
			conditions["plan_start >= ?"] = dStart
		}
	}
	if req.DateEnd != "" {
		if dEnd, err := time.Parse("2006-01-02", req.DateEnd); err == nil {
			conditions["plan_start <= ?"] = dEnd
		}
	}
	if req.ServiceID != 0 {
		conditions["id IN (SELECT plan_id FROM planner.plan_item WHERE service_id = ? AND deleted_at IS NULL)"] = req.ServiceID
	}
	if req.Active != nil {
		todayStr := time.Now().Format("2006-01-02")
		if *req.Active {
			conditions["plan_start <= ?"] = todayStr
			conditions["(plan_end IS NULL OR plan_end >= ?)"] = todayStr
		} else {
			conditions[fmt.Sprintf("(plan_start > '%s' OR (plan_end IS NOT NULL AND plan_end < '%s'))", todayStr, todayStr)] = nil
		}
	}

	plans, total, err := s.repo.FindPlans(req.Page, req.PageSize, conditions)
	if err != nil {
		s.logger.IPrintf(2, "Failed to retrieve plans: %v", err)
		return dto.NewPlanListResponse(500, "error", err.Error(), nil, 0, 0, 0, 0)
	}

	page := req.Page
	pageSize := req.PageSize
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}

	dtoPlans := s.mapPlansToDTO(plans)
	msg := fmt.Sprintf("retrieved %d records", len(dtoPlans))
	return dto.NewPlanListResponse(200, "success", msg, dtoPlans, page, pageSize, totalPages, total)
}

func (s *PlanList) mapPlansToDTO(plans []domain.Plan) []dto.PlanDTO {
	result := make([]dto.PlanDTO, len(plans))
	todayStr := time.Now().Format("2006-01-02")
	for i, p := range plans {
		nickname := ""
		custName := ""
		if p.Customer != nil {
			nickname = p.Customer.Nickname
			custName = p.Customer.Name
		} else if p.CustomerID != 0 && s.repo != nil {
			c, err := s.repo.GetCustomerByID(p.CustomerID)
			if err == nil && c != nil {
				nickname = c.Nickname
				custName = c.Name
			}
		}

		var planEndStr *string
		if p.PlanEnd != nil {
			s := p.PlanEnd.Format("2006-01-02")
			planEndStr = &s
		}

		// Items
		items := make([]dto.PlanItemDTO, len(p.Items))
		for j, it := range p.Items {
			svcName := ""
			if it.Service != nil {
				svcName = it.Service.Name
			} else if it.ServiceID != 0 && s.repo != nil {
				svc, err := s.repo.GetServiceByID(it.ServiceID)
				if err == nil && svc != nil {
					svcName = svc.Name
				}
			}
			items[j] = dto.PlanItemDTO{
				ID:          it.ID,
				ServiceID:   it.ServiceID,
				ServiceName: svcName,
				OrderIndex:  it.OrderIndex,
				Price:       it.Price,
			}
		}

		var agendaDTO *dto.PlanAgendaDTO
		if p.Agenda != nil {
			recName := domain.Recurrence(p.Agenda.Recurrence).String()
			wDayName := domain.WeekDay(p.Agenda.WeekDay).String()
			tTypeName := domain.TermType(p.Agenda.TermType).String()
			agendaDTO = &dto.PlanAgendaDTO{
				Recurrence:          p.Agenda.Recurrence,
				RecurrenceName:      recName,
				WeekDay:             p.Agenda.WeekDay,
				WeekDayName:         wDayName,
				TermType:            p.Agenda.TermType,
				TermTypeName:        tTypeName,
				PaymentDay:          p.Agenda.PaymentDay,
				MonthlyServiceLimit: p.Agenda.MonthlyServiceLimit,
			}
		}

		var pkgDTO *dto.PlanPackageDTO
		if p.Package != nil {
			pkgDTO = &dto.PlanPackageDTO{
				Quantity:    p.Package.Quantity,
				PaymentDate: p.Package.PaymentDate.Format("2006-01-02"),
			}
		}

		var nbDTO *dto.PlanNotebookDTO
		if p.Notebook != nil {
			nbDTO = &dto.PlanNotebookDTO{
				PaymentDay:  p.Notebook.PaymentDay,
				PaymentTerm: p.Notebook.PaymentTerm,
			}
		}

		planStartStr := p.PlanStart.Format("2006-01-02")
		isActive := planStartStr <= todayStr
		if p.PlanEnd != nil {
			planEndStr := p.PlanEnd.Format("2006-01-02")
			isActive = isActive && (todayStr <= planEndStr)
		}

		result[i] = dto.PlanDTO{
			ID:           p.ID,
			CustomerID:   p.CustomerID,
			Nickname:     nickname,
			CustomerName: custName,
			PlanStart:    planStartStr,
			PlanEnd:      planEndStr,
			PlanType:     p.PlanType,
			PlanTypeName: domain.PlanType(p.PlanType).String(),
			Price:        p.Price,
			Items:        items,
			Agenda:       agendaDTO,
			Package:      pkgDTO,
			Notebook:     nbDTO,
			Active:       isActive,
		}
	}
	return result
}
