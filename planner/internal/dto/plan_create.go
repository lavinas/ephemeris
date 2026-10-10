package dto

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

// PlanItemRequest represents an item (service) within a plan request
type PlanItemRequest struct {
	ServiceID  int64    `json:"service_id"`
	OrderIndex int      `json:"order_index"`
	Price      *float64 `json:"price,omitempty"`
}

// PlanAgendaRequest represents details for recurring agenda plan
type PlanAgendaRequest struct {
	Recurrence          int  `json:"recurrence"`
	WeekDay             int  `json:"week_day"`
	TermType            int  `json:"term_type"`
	PaymentDay          int  `json:"payment_day"`
	MonthlyServiceLimit *int `json:"monthly_service_limit,omitempty"`
}

// PlanPackageRequest represents details for session package plan
type PlanPackageRequest struct {
	Quantity    int    `json:"quantity"`
	PaymentDate string `json:"payment_date"`
}

// PlanNotebookRequest represents details for notebook plan
type PlanNotebookRequest struct {
	PaymentDay  *int `json:"payment_day,omitempty"`
	PaymentTerm *int `json:"payment_term,omitempty"`
}

// PlanCreateRequest represents a request to create a new plan
type PlanCreateRequest struct {
	CustomerID int64                `json:"customer_id,omitempty"`
	Nickname   string               `json:"nickname" validate:"required"`
	PlanStart  string               `json:"plan_start" validate:"required"`
	PlanEnd    *string              `json:"plan_end,omitempty"`
	PlanType   int                  `json:"plan_type" validate:"required"`
	Price      *float64             `json:"price,omitempty"`
	Items      []PlanItemRequest    `json:"items" validate:"required"`
	Agenda     *PlanAgendaRequest   `json:"agenda,omitempty"`
	Package    *PlanPackageRequest  `json:"package,omitempty"`
	Notebook   *PlanNotebookRequest `json:"notebook,omitempty"`
}

// PlanCreateResponse represents response after creating a new plan
type PlanCreateResponse struct {
	ResponseBase
	PlanID int64 `json:"plan_id"`
}

// NewPlanCreateResponse creates a new instance of PlanCreateResponse
func NewPlanCreateResponse(statusCode int, statusMessage string, errorMessage string, planID int64) *PlanCreateResponse {
	return &PlanCreateResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		PlanID: planID,
	}
}

// Validate checks if the PlanCreateRequest has valid data
func (r *PlanCreateRequest) Validate(repo port.Repository) error {
	var errs []error
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

	// Validate period overlap with existing active plans for the same customer & services
	if err := r.validateServicePeriodOverlap(repo); err != nil {
		return err
	}

	return nil
}

func (r *PlanCreateRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		return fmt.Errorf("nickname é obrigatório")
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
			return fmt.Errorf("cliente com nickname '%s' não encontrado na lista de clientes cadastrados", r.Nickname)
		}
		r.CustomerID = customer.ID
	}
	return nil
}

func (r *PlanCreateRequest) validateDates() error {
	if r.PlanStart == "" {
		return fmt.Errorf("data de início (plan_start) é obrigatória")
	}
	startTime, err := time.Parse("2006-01-02", r.PlanStart)
	if err != nil {
		return fmt.Errorf("formato inválido para data de início, esperado AAAA-MM-DD")
	}
	if startTime.IsZero() {
		return fmt.Errorf("data de início não pode ser nula")
	}

	if r.PlanEnd != nil && *r.PlanEnd != "" {
		endTime, err := time.Parse("2006-01-02", *r.PlanEnd)
		if err != nil {
			return fmt.Errorf("formato inválido para data de fim, esperado AAAA-MM-DD")
		}
		if endTime.Before(startTime) {
			return fmt.Errorf("data de término não pode ser anterior à data de início")
		}
	}
	return nil
}

func (r *PlanCreateRequest) validatePlanType() error {
	if r.PlanType < 1 || r.PlanType > 3 {
		return fmt.Errorf("tipo de plano inválido (%d), deve ser 1 (agenda), 2 (pacote) ou 3 (caderneta)", r.PlanType)
	}
	return nil
}

func (r *PlanCreateRequest) validateItems(repo port.Repository) error {
	if len(r.Items) == 0 {
		return fmt.Errorf("o plano deve conter pelo menos um serviço (item)")
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
	return nil
}

func (r *PlanCreateRequest) validateSpecialized() error {
	switch r.PlanType {
	case int(domain.PlanTypeAgenda):
		if r.Agenda == nil {
			return fmt.Errorf("registro na tabela especializada plan_agenda é obrigatório para plano tipo agenda")
		}
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
	case int(domain.PlanTypePackage):
		if r.Package == nil {
			return fmt.Errorf("registro na tabela especializada plan_package é obrigatório para plano tipo pacote")
		}
		if r.Package.Quantity <= 0 {
			return fmt.Errorf("quantidade de sessões do pacote deve ser maior que zero")
		}
		if r.Package.PaymentDate == "" {
			return fmt.Errorf("data de pagamento do pacote é obrigatória")
		}
		if _, err := time.Parse("2006-01-02", r.Package.PaymentDate); err != nil {
			return fmt.Errorf("data de pagamento do pacote inválida, esperado AAAA-MM-DD")
		}
	case int(domain.PlanTypeNotebook):
		if r.Notebook == nil {
			return fmt.Errorf("registro na tabela especializada plan_notebook é obrigatório para plano tipo caderneta")
		}
		hasPaymentDay := r.Notebook.PaymentDay != nil
		hasPaymentTerm := r.Notebook.PaymentTerm != nil
		if !hasPaymentDay && !hasPaymentTerm {
			return fmt.Errorf("plano caderneta deve ter uma e apenas uma das opções preenchidas: dia de pagamento ou prazo de pagamento")
		}
		if hasPaymentDay && hasPaymentTerm {
			return fmt.Errorf("plano caderneta deve ter apenas uma opção preenchida (dia de pagamento OU prazo de pagamento), não ambas")
		}
		if hasPaymentDay {
			if *r.Notebook.PaymentDay < 1 || *r.Notebook.PaymentDay > 31 {
				return fmt.Errorf("dia de pagamento na caderneta deve ser entre 1 e 31")
			}
		}
		if hasPaymentTerm {
			if *r.Notebook.PaymentTerm < 0 {
				return fmt.Errorf("prazo de pagamento na caderneta deve ser maior ou igual a zero")
			}
		}
	}
	return nil
}

func (r *PlanCreateRequest) validatePrices() error {
	itemPrices := make([]*float64, len(r.Items))
	for i, item := range r.Items {
		itemPrices[i] = item.Price
	}
	return domain.ValidatePlanPrices(r.Price, itemPrices)
}

func (r *PlanCreateRequest) validateServicePeriodOverlap(repo port.Repository) error {
	if repo == nil || r.CustomerID == 0 {
		return nil
	}
	startTime, err := time.Parse("2006-01-02", r.PlanStart)
	if err != nil {
		return err
	}
	var endTime *time.Time
	if r.PlanEnd != nil && *r.PlanEnd != "" {
		t, err := time.Parse("2006-01-02", *r.PlanEnd)
		if err == nil {
			endTime = &t
		}
	}

	serviceIDs := make([]int64, len(r.Items))
	for i, item := range r.Items {
		serviceIDs[i] = item.ServiceID
	}

	existingPlans, err := repo.FindCustomerActivePlansWithServices(r.CustomerID, serviceIDs, 0)
	if err != nil {
		return fmt.Errorf("erro ao verificar coincidência de serviços em planos ativos: %v", err)
	}

	for _, ep := range existingPlans {
		if domain.PeriodsOverlap(startTime, endTime, ep.PlanStart, ep.PlanEnd) {
			// Find which service collided
			for _, epItem := range ep.Items {
				for _, reqItem := range r.Items {
					if epItem.ServiceID == reqItem.ServiceID {
						svcName := fmt.Sprintf("ID %d", epItem.ServiceID)
						if epItem.Service != nil {
							svcName = epItem.Service.Name
						}
						epEndStr := "em andamento/indefinido"
						if ep.PlanEnd != nil {
							epEndStr = ep.PlanEnd.Format("02/01/2006")
						}
						return fmt.Errorf("conflito de período: o cliente '%s' já possui o plano %d ativo com o serviço '%s' no período de %s até %s",
							r.Nickname, ep.ID, svcName, ep.PlanStart.Format("02/01/2006"), epEndStr)
					}
				}
			}
		}
	}
	return nil
}

// Reset clears the fields of the PlanCreateRequest
func (r *PlanCreateRequest) Reset() {
	r.CustomerID = 0
	r.Nickname = ""
	r.PlanStart = ""
	r.PlanEnd = nil
	r.PlanType = 0
	r.Price = nil
	r.Items = nil
	r.Agenda = nil
	r.Package = nil
	r.Notebook = nil
}
