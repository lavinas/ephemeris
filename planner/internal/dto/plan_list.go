package dto

import (
	"errors"
	"strings"
	"time"

	"planner/internal/port"
)

// PlanListRequest represents a request to list plans with filters and pagination
type PlanListRequest struct {
	Page        int     `json:"page" validate:"required,gt=0"`
	PageSize    int     `json:"page_size" validate:"required,gt=0"`
	PlanID      int64   `json:"plan_id,omitempty"`
	Nickname    string  `json:"nickname,omitempty"`
	CustomerIDs []int64 `json:"-"`
	PlanType    int     `json:"plan_type,omitempty"`
	DateStart   string  `json:"date_start,omitempty"`
	DateEnd     string  `json:"date_end,omitempty"`
	ServiceID   int64   `json:"service_id,omitempty"`
}

// PlanListResponse represents the response containing a list of plans
type PlanListResponse struct {
	ResponseBase
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
	TotalCount int64     `json:"total_count"`
	Plans      []PlanDTO `json:"plans"`
}

// PlanDTO represents a plan and all its detailed data
type PlanDTO struct {
	ID           int64            `json:"id"`
	CustomerID   int64            `json:"customer_id"`
	Nickname     string           `json:"nickname"`
	CustomerName string           `json:"customer_name"`
	PlanStart    string           `json:"plan_start"`
	PlanEnd      *string          `json:"plan_end,omitempty"`
	PlanType     int              `json:"plan_type"`
	PlanTypeName string           `json:"plan_type_name"`
	Price        *float64         `json:"price,omitempty"`
	Items        []PlanItemDTO    `json:"items"`
	Agenda       *PlanAgendaDTO   `json:"agenda,omitempty"`
	Package      *PlanPackageDTO  `json:"package,omitempty"`
	Notebook     *PlanNotebookDTO `json:"notebook,omitempty"`
}

// PlanItemDTO represents a service item within a plan
type PlanItemDTO struct {
	ID          int64    `json:"id"`
	ServiceID   int64    `json:"service_id"`
	ServiceName string   `json:"service_name"`
	OrderIndex  int      `json:"order_index"`
	Price       *float64 `json:"price,omitempty"`
}

// PlanAgendaDTO represents agenda specialized details
type PlanAgendaDTO struct {
	Recurrence          int    `json:"recurrence"`
	RecurrenceName      string `json:"recurrence_name"`
	WeekDay             int    `json:"week_day"`
	WeekDayName         string `json:"week_day_name"`
	TermType            int    `json:"term_type"`
	TermTypeName        string `json:"term_type_name"`
	PaymentDay          int    `json:"payment_day"`
	MonthlyServiceLimit *int   `json:"monthly_service_limit,omitempty"`
}

// PlanPackageDTO represents package specialized details
type PlanPackageDTO struct {
	Quantity    int    `json:"quantity"`
	PaymentDate string `json:"payment_date"`
}

// PlanNotebookDTO represents notebook specialized details
type PlanNotebookDTO struct {
	PaymentDay  *int `json:"payment_day,omitempty"`
	PaymentTerm *int `json:"payment_term,omitempty"`
}

// NewPlanListResponse creates a new instance of PlanListResponse
func NewPlanListResponse(statusCode int, statusMessage string, errorMessage string, plans []PlanDTO,
	page int, pageSize int, totalPages int, totalCount int64) *PlanListResponse {
	return &PlanListResponse{
		ResponseBase: ResponseBase{
			HttpCode: statusCode,
			Status:   statusMessage,
			Message:  errorMessage,
		},
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalCount: totalCount,
		Plans:      plans,
	}
}

// Validate checks if the PlanListRequest has valid data
func (r *PlanListRequest) Validate(repo port.Repository) error {
	var errs []error
	if err := r.validatePlanID(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateNickname(repo); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateDates(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePagination(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validatePlanType(); err != nil {
		errs = append(errs, err)
	}
	if err := r.validateServiceID(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		err := errors.Join(errs...)
		return errors.New(strings.ReplaceAll(err.Error(), "\n", "; "))
	}
	return nil
}

func (r *PlanListRequest) validatePlanID() error {
	if r.PlanID < 0 {
		return errors.New("plan_id não pode ser negativo")
	}
	return nil
}

func (r *PlanListRequest) validateNickname(repo port.Repository) error {
	if r.Nickname == "" {
		r.CustomerIDs = nil
		return nil
	}
	if repo == nil {
		return nil
	}
	customers, err := repo.FindCustomers(0, 0, 1, nil, &r.Nickname, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	if len(customers) == 0 {
		r.CustomerIDs = []int64{-1}
		return nil
	}
	r.CustomerIDs = make([]int64, len(customers))
	for i, c := range customers {
		r.CustomerIDs[i] = c.ID
	}
	return nil
}

func (r *PlanListRequest) validateDates() error {
	var sd *time.Time
	if r.DateStart != "" {
		startDate, err := time.Parse("2006-01-02", r.DateStart)
		if err != nil {
			return errors.New("formato inválido para data de início, esperado AAAA-MM-DD")
		}
		sd = &startDate
	}
	if r.DateEnd != "" {
		endDate, err := time.Parse("2006-01-02", r.DateEnd)
		if err != nil {
			return errors.New("formato inválido para data de fim, esperado AAAA-MM-DD")
		}
		if sd != nil && endDate.Before(*sd) {
			return errors.New("data de fim não pode ser anterior à data de início")
		}
	}
	return nil
}

func (r *PlanListRequest) validatePagination() error {
	var errs []error
	if r.Page <= 0 {
		errs = append(errs, errors.New("page deve ser maior que 0"))
	}
	if r.PageSize <= 0 {
		errs = append(errs, errors.New("page_size deve ser maior que 0"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (r *PlanListRequest) validatePlanType() error {
	if r.PlanType < 0 || r.PlanType > 3 {
		return errors.New("plan_type inválido, deve ser entre 1 e 3")
	}
	return nil
}

func (r *PlanListRequest) validateServiceID() error {
	if r.ServiceID < 0 {
		return errors.New("service_id não pode ser negativo")
	}
	return nil
}

// Reset clears the fields of PlanListRequest
func (r *PlanListRequest) Reset() {
	r.Page = 1
	r.PageSize = 10
	r.PlanID = 0
	r.Nickname = ""
	r.CustomerIDs = nil
	r.PlanType = 0
	r.DateStart = ""
	r.DateEnd = ""
	r.ServiceID = 0
}
