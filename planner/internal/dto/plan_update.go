package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// PlanUpdateRequest represents a request to update an existing plan
type PlanUpdateRequest struct {
	ID           int64                `json:"id" validate:"required"`
	CustomerID   int64                `json:"customer_id,omitempty"`
	Nickname     string               `json:"nickname,omitempty"`
	PlanStart    string               `json:"plan_start,omitempty"`
	PlanEnd      *string              `json:"plan_end,omitempty"`
	ClearPlanEnd bool                 `json:"clear_plan_end,omitempty"`
	PlanType     int                  `json:"plan_type,omitempty"`
	Price        *float64             `json:"price,omitempty"`
	ClearPrice   bool                 `json:"clear_price,omitempty"`
	Items        []PlanItemRequest    `json:"items,omitempty"`
	Agenda       *PlanAgendaRequest   `json:"agenda,omitempty"`
	Package      *PlanPackageRequest  `json:"package,omitempty"`
	Notebook     *PlanNotebookRequest `json:"notebook,omitempty"`
	Plan         domain.Plan          `json:"-"`
}

// PlanUpdateResponse represents the response after updating a plan
type PlanUpdateResponse struct {
	ResponseBase
}

// NewPlanUpdateResponse creates a new instance of PlanUpdateResponse
func NewPlanUpdateResponse(statusCode int, statusMessage string, errorMessage string) *PlanUpdateResponse {
	return &PlanUpdateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
	}
}

// GetPlan returns the loaded domain.Plan
func (r *PlanUpdateRequest) GetPlan() *domain.Plan {
	return &r.Plan
}

// Validate checks if the PlanUpdateRequest has valid data
func (r *PlanUpdateRequest) Validate(repo port.Repository) error {
	var errs []error
	if err := r.validateID(repo); err != nil {
		return err
	}
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePlanType(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateItems(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateSpecialized(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePrices(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}

	if err := r.validateServicePeriodOverlap(repo); err != nil {
		return err
	}

	return nil
}

func (r *PlanUpdateRequest) validateID(repo port.Repository) error {
	if r.ID <= 0 {
		return errors.New("id do plano inválido: deve ser maior que zero")
	}
	if repo == nil {
		return nil
	}
	existing, err := repo.GetPlanByID(r.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar plano: %v", err)
	}
	if existing == nil {
		return fmt.Errorf("plano com ID %d não encontrado", r.ID)
	}
	r.Plan = *existing
	return nil
}

func (r *PlanUpdateRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		r.CustomerID = r.Plan.CustomerID
		return nil
	}
	if r.Nickname != strings.ToLower(r.Nickname) {
		return fmt.Errorf("nickname deve estar em letras minúsculas")
	}
	if strings.Contains(r.Nickname, " ") {
		return fmt.Errorf("nickname não deve conter espaços")
	}
	for _, char := range r.Nickname {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("nickname deve conter apenas letras minúsculas, números e sublinhados")
		}
	}
	if repo != nil {
		customer, err := repo.GetCustomer(1, r.Nickname)
		if err != nil {
			return fmt.Errorf("erro ao buscar cliente: %v", err)
		}
		if customer == nil {
			return fmt.Errorf("cliente com nickname '%s' não encontrado", r.Nickname)
		}
		r.CustomerID = customer.ID
	}
	return nil
}

func (r *PlanUpdateRequest) validateDates() error {
	startTime := r.Plan.PlanStart
	if r.PlanStart != "" {
		st, err := time.Parse("2006-01-02", r.PlanStart)
		if err != nil {
			return fmt.Errorf("formato inválido para data de início, esperado AAAA-MM-DD")
		}
		startTime = st
	}

	var endTime *time.Time
	if r.ClearPlanEnd {
		endTime = nil
	} else if r.PlanEnd != nil && *r.PlanEnd != "" {
		et, err := time.Parse("2006-01-02", *r.PlanEnd)
		if err != nil {
			return fmt.Errorf("formato inválido para data de fim, esperado AAAA-MM-DD")
		}
		endTime = &et
	} else {
		endTime = r.Plan.PlanEnd
	}

	if endTime != nil && endTime.Before(startTime) {
		return fmt.Errorf("data de término não pode ser anterior à data de início")
	}
	return nil
}

func (r *PlanUpdateRequest) validatePlanType() error {
	if r.PlanType != 0 && (r.PlanType < 1 || r.PlanType > 3) {
		return fmt.Errorf("tipo de plano inválido (%d), deve ser 1, 2 ou 3", r.PlanType)
	}
	return nil
}

func (r *PlanUpdateRequest) validateItems(repo port.Repository) error {
	if len(r.Items) == 0 {
		return nil
	}
	for i, item := range r.Items {
		if item.ServiceID <= 0 {
			return fmt.Errorf("item %d: service_id inválido", i+1)
		}
		if repo != nil {
			svc, err := repo.GetServiceByID(item.ServiceID)
			if err != nil {
				return fmt.Errorf("erro ao verificar serviço %d: %v", item.ServiceID, err)
			}
			if svc == nil {
				return fmt.Errorf("serviço com ID %d não encontrado", item.ServiceID)
			}
		}
		if item.Price != nil && *item.Price < 0 {
			return fmt.Errorf("item %d: preço não pode ser negativo", i+1)
		}
		if r.Items[i].OrderIndex <= 0 {
			r.Items[i].OrderIndex = i + 1
		}
	}

	if len(r.Items) > 1 {
		orders := make([]int, len(r.Items))
		for i, item := range r.Items {
			orders[i] = item.OrderIndex
		}
		if err := domain.ValidatePlanItemOrders(orders); err != nil {
			return err
		}
	}

	return nil
}

func (r *PlanUpdateRequest) validateSpecialized() error {
	planType := r.PlanType
	if planType == 0 {
		planType = r.Plan.PlanType
	}

	switch planType {
	case int(domain.PlanTypeAgenda):
		if r.Agenda != nil {
			if r.Agenda.Recurrence < 1 || r.Agenda.Recurrence > 3 {
				return fmt.Errorf("recorrência inválida, deve ser 1 (semanal), 2 (quinzenal) ou 3 (mensal)")
			}
			if r.Agenda.WeekDay < 1 || r.Agenda.WeekDay > 7 {
				return fmt.Errorf("dia da semana inválido, deve ser entre 1 (domingo) e 7 (sábado)")
			}
			if r.Agenda.TermType < 1 || r.Agenda.TermType > 2 {
				return fmt.Errorf("tipo de vigência inválido, deve ser 1 (pré-pago) ou 2 (pós-pago)")
			}
			if r.Agenda.PaymentDay < 1 || r.Agenda.PaymentDay > 31 {
				return fmt.Errorf("dia de pagamento deve ser entre 1 e 31")
			}
			if r.Agenda.MonthlyServiceLimit != nil && *r.Agenda.MonthlyServiceLimit <= 0 {
				return fmt.Errorf("limite de sessões mensal deve ser maior que zero")
			}
		}
	case int(domain.PlanTypePackage):
		if r.Package != nil {
			if r.Package.Quantity <= 0 {
				return fmt.Errorf("quantidade de sessões do pacote deve ser maior que zero")
			}
			if r.Package.PaymentDate != "" {
				if _, err := time.Parse("2006-01-02", r.Package.PaymentDate); err != nil {
					return fmt.Errorf("data de pagamento do pacote inválida, esperado AAAA-MM-DD")
				}
			}
		}
	case int(domain.PlanTypeNotebook):
		if r.Notebook != nil {
			hasPaymentDay := r.Notebook.PaymentDay != nil
			hasPaymentTerm := r.Notebook.PaymentTerm != nil
			if !hasPaymentDay && !hasPaymentTerm {
				return fmt.Errorf("plano caderneta deve ter uma e apenas uma das opções preenchidas: dia de pagamento ou prazo de pagamento")
			}
			if hasPaymentDay && hasPaymentTerm {
				return fmt.Errorf("plano caderneta deve ter apenas uma opção preenchida (dia de pagamento OU prazo de pagamento), não ambas")
			}
			if hasPaymentDay && (*r.Notebook.PaymentDay < 1 || *r.Notebook.PaymentDay > 31) {
				return fmt.Errorf("dia de pagamento na caderneta deve ser entre 1 e 31")
			}
			if hasPaymentTerm && *r.Notebook.PaymentTerm < 0 {
				return fmt.Errorf("prazo de pagamento na caderneta deve ser maior ou igual a zero")
			}
		}
	}
	return nil
}

func (r *PlanUpdateRequest) validateServicePeriodOverlap(repo port.Repository) error {
	if repo == nil {
		return nil
	}
	customerID := r.CustomerID
	if customerID == 0 {
		customerID = r.Plan.CustomerID
	}

	startTime := r.Plan.PlanStart
	if r.PlanStart != "" {
		st, err := time.Parse("2006-01-02", r.PlanStart)
		if err == nil {
			startTime = st
		}
	}

	var endTime *time.Time
	if r.ClearPlanEnd {
		endTime = nil
	} else if r.PlanEnd != nil && *r.PlanEnd != "" {
		et, err := time.Parse("2006-01-02", *r.PlanEnd)
		if err == nil {
			endTime = &et
		}
	} else {
		endTime = r.Plan.PlanEnd
	}

	// Service IDs to check: from updated items if provided, or from existing items
	serviceIDs := make([]int64, 0)
	if len(r.Items) > 0 {
		for _, it := range r.Items {
			serviceIDs = append(serviceIDs, it.ServiceID)
		}
	} else {
		for _, it := range r.Plan.Items {
			serviceIDs = append(serviceIDs, it.ServiceID)
		}
	}

	if len(serviceIDs) == 0 {
		return nil
	}

	existingPlans, err := repo.FindCustomerActivePlansWithServices(customerID, serviceIDs, r.ID)
	if err != nil {
		return fmt.Errorf("erro ao verificar coincidência de serviços em planos ativos: %v", err)
	}

	for _, ep := range existingPlans {
		if domain.PeriodsOverlap(startTime, endTime, ep.PlanStart, ep.PlanEnd) {
			for _, epItem := range ep.Items {
				for _, sid := range serviceIDs {
					if epItem.ServiceID == sid {
						svcName := fmt.Sprintf("ID %d", epItem.ServiceID)
						if epItem.Service != nil {
							svcName = epItem.Service.Name
						}
						epEndStr := "em andamento/indefinido"
						if ep.PlanEnd != nil {
							epEndStr = ep.PlanEnd.Format("02/01/2006")
						}
						return fmt.Errorf("conflito de período: o cliente já possui o plano %d ativo com o serviço '%s' no período de %s até %s",
							ep.ID, svcName, ep.PlanStart.Format("02/01/2006"), epEndStr)
					}
				}
			}
		}
	}
	return nil
}

func (r *PlanUpdateRequest) validatePrices() error {
	var targetPrice *float64
	if r.ClearPrice {
		targetPrice = nil
	} else if r.Price != nil {
		targetPrice = r.Price
	} else {
		targetPrice = r.Plan.Price
	}

	var itemPrices []*float64
	if len(r.Items) > 0 {
		itemPrices = make([]*float64, len(r.Items))
		for i, it := range r.Items {
			itemPrices[i] = it.Price
		}
	} else {
		itemPrices = make([]*float64, len(r.Plan.Items))
		for i, it := range r.Plan.Items {
			itemPrices[i] = it.Price
		}
	}

	return domain.ValidatePlanPrices(targetPrice, itemPrices)
}

// Reset clears the fields
func (r *PlanUpdateRequest) Reset() {
	r.ID = 0
	r.CustomerID = 0
	r.Nickname = ""
	r.PlanStart = ""
	r.PlanEnd = nil
	r.ClearPlanEnd = false
	r.PlanType = 0
	r.Price = nil
	r.ClearPrice = false
	r.Items = nil
	r.Agenda = nil
	r.Package = nil
	r.Notebook = nil
}
