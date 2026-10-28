package service

import (
	"fmt"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// SessionServices is an application service that retrieves available services for a session,
// filtering by the customer's active plans when a customer nickname is provided and has active plans.
type SessionServices struct {
	Base
	nowFunc func() time.Time
}

// NewSessionServices creates a new instance of SessionServices with the provided repository and logger.
func NewSessionServices(repo port.Repository, logger port.Logger) *SessionServices {
	return &SessionServices{
		Base:    *NewBase(repo, logger),
		nowFunc: time.Now,
	}
}

// SetNowFunc sets a custom time provider function (useful for deterministic testing).
func (s *SessionServices) SetNowFunc(fn func() time.Time) {
	if fn != nil {
		s.nowFunc = fn
	}
}

func (s *SessionServices) now() time.Time {
	if s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

// Run executes the session services retrieval use case using the provided input DTO and returns an output DTO.
func (s *SessionServices) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing session services request: %v", input)

	dtoIn, ok := input.(*dto.SessionServicesRequest)
	if !ok || dtoIn == nil {
		s.logger.IPrintf(2, "Invalid input DTO type: %T", input)
		return dto.NewSessionServicesResponse(400, "error", "Invalid input DTO type", nil)
	}
	if err := dtoIn.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Invalid session services request: %v", err)
		return dto.NewSessionServicesResponse(400, "error", err.Error(), nil)
	}

	services, err := s.resolveServices(*dtoIn)
	if err != nil {
		s.logger.IPrintf(2, "Error retrieving session services: %v", err)
		return dto.NewSessionServicesResponse(500, "error", err.Error(), nil)
	}

	dtoOut := make([]dto.ServiceDTO, len(services))
	for i, svc := range services {
		dtoOut[i] = dto.ServiceDTO{
			ID:             svc.ID,
			VendorID:       svc.VendorID,
			Name:           svc.Name,
			Description:    svc.Description,
			SessionMinutes: svc.SessionMinutes,
		}
	}

	msg := fmt.Sprintf("Successfully retrieved %d session services", len(dtoOut))
	s.logger.IPrintf(2, "%s", msg)
	return dto.NewSessionServicesResponse(200, "success", msg, dtoOut)
}

func (s *SessionServices) resolveServices(req dto.SessionServicesRequest) ([]domain.Service, error) {
	if s.repo == nil {
		return nil, nil
	}

	allServices, err := s.repo.FindServices(req.VendorID)
	if err != nil {
		return nil, err
	}

	today := s.now()
	if req.Nickname == "" {
		return domain.SortServicesByName(allServices), nil
	}

	customer, err := s.repo.GetCustomer(req.VendorID, req.Nickname)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return domain.SortServicesByName(allServices), nil
	}

	todayStr := today.Format("2006-01-02")
	conditions := map[string]interface{}{
		"deleted_at IS NULL":                  nil,
		"customer_id = ?":                     customer.ID,
		"plan_start <= ?":                     todayStr,
		"(plan_end IS NULL OR plan_end >= ?)": todayStr,
	}

	plans, _, err := s.repo.FindPlans(0, 0, conditions)
	if err != nil {
		return nil, err
	}

	var customerPlans []domain.Plan
	for i := range plans {
		p := plans[i]
		if p.CustomerID != customer.ID || p.DeletedAt != nil {
			continue
		}
		for j := range p.Items {
			if p.Items[j].DeletedAt == nil && p.Items[j].Service == nil && p.Items[j].ServiceID > 0 {
				if svc, err := s.repo.GetServiceByID(p.Items[j].ServiceID); err == nil && svc != nil {
					p.Items[j].Service = svc
				}
			}
		}
		customerPlans = append(customerPlans, p)
	}

	return domain.ResolveServicesForCustomerPlans(customerPlans, allServices, today), nil
}

