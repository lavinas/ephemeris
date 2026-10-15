package service

import (
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
)

// InvoicingService defines the interface for running invoicing logic.
type InvoicingService interface {
	Run(input port.InDTO) port.OutDTO
}

// PlanCreate is a service that handles the creation of plans.
type PlanCreate struct {
	Base
	invoicing InvoicingService
}

// NewPlanCreate creates a new instance of PlanCreate.
func NewPlanCreate(repo port.Repository, logger port.Logger) *PlanCreate {
	return &PlanCreate{
		Base:      *NewBase(repo, logger),
		invoicing: NewInvoicing(repo, logger),
	}
}

// NewPlanCreateWithInvoicing creates a new instance of PlanCreate with a custom invoicing service.
func NewPlanCreateWithInvoicing(repo port.Repository, logger port.Logger, invoicing InvoicingService) *PlanCreate {
	return &PlanCreate{
		Base:      *NewBase(repo, logger),
		invoicing: invoicing,
	}
}

// Run executes the plan creation process using the provided input DTO.
func (s *PlanCreate) Run(input port.InDTO) port.OutDTO {
	s.logger.IPrintf(2, "Processing create plan request: %v", input)
	if err := input.Validate(s.repo); err != nil {
		s.logger.IPrintf(2, "Validation failed: %v", err)
		return dto.NewPlanCreateResponse(400, "error", err.Error(), 0)
	}

	planModel := s.createPlanModel(input.(*dto.PlanCreateRequest))

	if err := s.repo.SavePlan(planModel); err != nil {
		s.logger.IPrintf(2, "Failed to save plan: %v", err)
		return dto.NewPlanCreateResponse(500, "error", err.Error(), 0)
	}

	s.logger.IPrintf(2, "Successfully created plan with ID: %d", planModel.ID)

	if planModel.PlanType == int(domain.PlanTypePackage) {
		s.generatePackageInvoice(planModel)
	}

	return dto.NewPlanCreateResponse(200, "success", "plano criado com sucesso", planModel.ID)
}

func (s *PlanCreate) generatePackageInvoice(plan *domain.Plan) {
	planStartStr := plan.PlanStart.Format("2006-01-02")
	planType := int(domain.PlanTypePackage)
	invReq := &dto.InvoicingRequest{
		PlanID:    &plan.ID,
		PlanType:  &planType,
		StartDate: &planStartStr,
		EndDate:   &planStartStr,
	}

	invSvc := s.invoicing
	if invSvc == nil {
		invSvc = NewInvoicing(s.repo, s.logger)
	}

	resp := invSvc.Run(invReq)
	if resp != nil && resp.GetStatusCode() != 200 {
		s.logger.IPrintf(2, "Aviso: falha ao gerar pré-fatura automática para plano pacote %d: %s",
			plan.ID, resp.GetMessage())
	} else {
		s.logger.IPrintf(2, "Pré-fatura gerada automaticamente com sucesso para plano pacote %d", plan.ID)
	}
}

func (s *PlanCreate) createPlanModel(req *dto.PlanCreateRequest) *domain.Plan {
	startTime, _ := time.Parse("2006-01-02", req.PlanStart)
	var endTime *time.Time
	if req.PlanEnd != nil && *req.PlanEnd != "" {
		t, err := time.Parse("2006-01-02", *req.PlanEnd)
		if err == nil {
			endTime = &t
		}
	}

	plan := domain.NewPlan(req.CustomerID, startTime, endTime, req.PlanType, req.Price)

	// Items
	items := make([]domain.PlanItem, len(req.Items))
	for i, it := range req.Items {
		orderIdx := it.OrderIndex
		if orderIdx <= 0 {
			orderIdx = i + 1
		}
		items[i] = domain.PlanItem{
			ServiceID:  it.ServiceID,
			OrderIndex: orderIdx,
			Price:      it.Price,
		}
	}
	plan.Items = items

	// Specialized tables
	switch req.PlanType {
	case int(domain.PlanTypeAgenda):
		if req.Agenda != nil {
			plan.Agenda = &domain.PlanAgenda{
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
				Quantity:    req.Package.Quantity,
				PaymentDate: pDate,
			}
		}
	case int(domain.PlanTypeNotebook):
		if req.Notebook != nil {
			plan.Notebook = &domain.PlanNotebook{
				PaymentDay: req.Notebook.PaymentDay,
			}
		}
	case int(domain.PlanTypePayPerView):
		if req.PayPerView != nil {
			plan.PayPerView = &domain.PlanPayPerView{
				PaymentTerm: req.PayPerView.PaymentTerm,
			}
		}
	}

	return plan
}
