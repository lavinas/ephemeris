package domain

import (
	"errors"
	"fmt"
	"time"
)

// PlanType represents the type of plan
type PlanType int

const (
	PlanTypeAgenda   PlanType = 1 // 1: plan_agenda
	PlanTypePackage  PlanType = 2 // 2: plan_package
	PlanTypeNotebook PlanType = 3 // 3: plan_notebook
)

// String returns human-readable label for PlanType
func (pt PlanType) String() string {
	switch pt {
	case PlanTypeAgenda:
		return "Agenda"
	case PlanTypePackage:
		return "Pacote"
	case PlanTypeNotebook:
		return "Caderneta"
	default:
		return "Desconhecido"
	}
}

// Recurrence represents recurrence interval in plan_agenda
type Recurrence int

const (
	RecurrenceWeekly   Recurrence = 1 // 1: weekly
	RecurrenceBiWeekly Recurrence = 2 // 2: bi-weekly
	RecurrenceMonthly  Recurrence = 3 // 3: monthly
)

func (r Recurrence) String() string {
	switch r {
	case RecurrenceWeekly:
		return "Semanal"
	case RecurrenceBiWeekly:
		return "Quinzenal"
	case RecurrenceMonthly:
		return "Mensal"
	default:
		return ""
	}
}

// WeekDay represents days of the week in plan_agenda
type WeekDay int

const (
	WeekDaySunday    WeekDay = 1
	WeekDayMonday    WeekDay = 2
	WeekDayTuesday   WeekDay = 3
	WeekDayWednesday WeekDay = 4
	WeekDayThursday  WeekDay = 5
	WeekDayFriday    WeekDay = 6
	WeekDaySaturday  WeekDay = 7
)

func (w WeekDay) String() string {
	switch w {
	case WeekDaySunday:
		return "Domingo"
	case WeekDayMonday:
		return "Segunda-feira"
	case WeekDayTuesday:
		return "Terça-feira"
	case WeekDayWednesday:
		return "Quarta-feira"
	case WeekDayThursday:
		return "Quinta-feira"
	case WeekDayFriday:
		return "Sexta-feira"
	case WeekDaySaturday:
		return "Sábado"
	default:
		return ""
	}
}

// TermType represents billing term type in plan_agenda
type TermType int

const (
	TermTypePrePaid  TermType = 1 // 1: pre-paid
	TermTypePostPaid TermType = 2 // 2: post-paid
)

func (t TermType) String() string {
	switch t {
	case TermTypePrePaid:
		return "Pré-pago"
	case TermTypePostPaid:
		return "Pós-pago"
	default:
		return ""
	}
}

// Plan represents the base subscription or session plan model.
type Plan struct {
	ID         int64         `gorm:"primaryKey;autoIncrement"`
	CustomerID int64         `gorm:"not null;index"`
	Customer   *Customer     `gorm:"foreignKey:CustomerID;references:ID"`
	PlanStart  time.Time     `gorm:"type:date;not null"`
	PlanEnd    *time.Time    `gorm:"type:date"`
	PlanType   int           `gorm:"not null"` // 1: agenda, 2: package, 3: notebook
	Price      *float64      `gorm:"type:numeric(15,2)"`
	CreatedAt  time.Time     `gorm:"not null"`
	UpdatedAt  time.Time     `gorm:"not null"`
	DeletedAt  *time.Time    `gorm:"index"`
	Items      []PlanItem    `gorm:"foreignKey:PlanID;references:ID"`
	Agenda     *PlanAgenda   `gorm:"foreignKey:ID;references:ID"`
	Package    *PlanPackage  `gorm:"foreignKey:ID;references:ID"`
	Notebook   *PlanNotebook `gorm:"foreignKey:ID;references:ID"`
}

// TableName specifies table name for Plan
func (Plan) TableName() string {
	return "plan"
}

// NewPlan creates a new Plan instance.
func NewPlan(customerID int64, planStart time.Time, planEnd *time.Time, planType int, price *float64) *Plan {
	now := time.Now()
	return &Plan{
		CustomerID: customerID,
		PlanStart:  planStart,
		PlanEnd:    planEnd,
		PlanType:   planType,
		Price:      price,
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  nil,
	}
}

// PlanItem represents a service included in a plan.
type PlanItem struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	PlanID     int64      `gorm:"not null;index"`
	ServiceID  int64      `gorm:"not null;index"`
	Service    *Service   `gorm:"foreignKey:ServiceID;references:ID"`
	OrderIndex int        `gorm:"not null"`
	Price      *float64   `gorm:"type:numeric(15,2)"`
	CreatedAt  time.Time  `gorm:"not null"`
	UpdatedAt  time.Time  `gorm:"not null"`
	DeletedAt  *time.Time `gorm:"index"`
}

// TableName specifies table name for PlanItem
func (PlanItem) TableName() string {
	return "plan_item"
}

// PlanAgenda represents specialization for recurring agenda plans (plan_type = 1).
type PlanAgenda struct {
	ID                  int64 `gorm:"primaryKey"`
	Recurrence          int   `gorm:"not null"` // 1: weekly, 2: bi-weekly, 3: monthly
	WeekDay             int   `gorm:"not null"` // 1..7 (1: sunday)
	TermType            int   `gorm:"not null"` // 1: pre-paid, 2: post-paid
	PaymentDay          int   `gorm:"not null"` // 1..31
	MonthlyServiceLimit *int  `gorm:"null"`     // limit number of sessions per month, if null no limit
}

// TableName specifies table name for PlanAgenda
func (PlanAgenda) TableName() string {
	return "plan_agenda"
}

// PlanPackage represents specialization for session package plans (plan_type = 2).
type PlanPackage struct {
	ID          int64     `gorm:"primaryKey"`
	Quantity    int       `gorm:"not null"` // quantity of sessions
	PaymentDate time.Time `gorm:"type:date;not null"`
}

// TableName specifies table name for PlanPackage
func (PlanPackage) TableName() string {
	return "plan_package"
}

// PlanNotebook represents specialization for notebook plans (plan_type = 3).
type PlanNotebook struct {
	ID          int64 `gorm:"primaryKey"`
	PaymentDay  *int  `gorm:"null"` // 1..31
	PaymentTerm *int  `gorm:"null"` // >= 0
}

// TableName specifies table name for PlanNotebook
func (PlanNotebook) TableName() string {
	return "plan_notebook"
}

// PeriodsOverlap checks if two date periods overlap.
// A nil end date is considered infinite (+inf).
// Two intervals [S1, E1] and [S2, E2] overlap if:
// (E2 == nil || S1 <= E2) && (E1 == nil || S2 <= E1)
func PeriodsOverlap(start1 time.Time, end1 *time.Time, start2 time.Time, end2 *time.Time) bool {
	cond1 := (end2 == nil || !start1.After(*end2))
	cond2 := (end1 == nil || !start2.After(*end1))
	return cond1 && cond2
}

// ValidatePlanPrices checks the business rules for pricing on a plan and its items:
// 1. Plan price and item prices are mutually exclusive (if plan price is non-null, item prices must be null, and vice-versa).
// 2. Either the plan or the items must have a defined price (cannot both be null).
// 3. If an item has a price, ALL items in the plan must have a price (and non-negative).
func ValidatePlanPrices(planPrice *float64, itemPrices []*float64) error {
	hasPlanPrice := planPrice != nil
	if hasPlanPrice && *planPrice < 0 {
		return errors.New("o valor do plano não pode ser negativo")
	}

	totalItems := len(itemPrices)
	itemsWithPrice := 0
	for i, ip := range itemPrices {
		if ip != nil {
			if *ip < 0 {
				return fmt.Errorf("item %d: preço não pode ser negativo", i+1)
			}
			itemsWithPrice++
		}
	}

	if hasPlanPrice && itemsWithPrice > 0 {
		return errors.New("os valores do plano e dos itens são mutuamente exclusivos: se o valor do plano for informado, os itens não devem ter valor")
	}

	if !hasPlanPrice && itemsWithPrice == 0 {
		return errors.New("o plano ou os itens devem ter valor definido")
	}

	if !hasPlanPrice && itemsWithPrice > 0 && itemsWithPrice < totalItems {
		return errors.New("caso os itens tenham valor, todos os itens do plano devem ter valor")
	}

	return nil
}

