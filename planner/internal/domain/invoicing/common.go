package invoicing

import (
	"math"
	"sort"
	"time"

	"planner/internal/domain"
)

// DaysInMonth returns the number of days in the specified year and month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// DistributeQuantityForService distributes a total session quantity across plan items
// ordered by OrderIndex (matching AgendaStrategy and PackageStrategy) and returns the share
// for the given serviceID. If the plan has at most 1 item or serviceID is not matched, returns totalQty.
func DistributeQuantityForService(items []domain.PlanItem, totalQty int, serviceID int64) int {
	if len(items) <= 1 || serviceID <= 0 {
		return totalQty
	}

	sorted := make([]domain.PlanItem, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OrderIndex < sorted[j].OrderIndex
	})

	nItems := len(sorted)
	baseQty := totalQty / nItems
	remainder := totalQty % nItems

	serviceQty := 0
	matched := false
	for i, item := range sorted {
		qty := baseQty
		if i < remainder {
			qty++
		}
		if item.ServiceID == serviceID {
			serviceQty += qty
			matched = true
		}
	}
	if !matched {
		return totalQty
	}
	return serviceQty
}

// CalculateProportionalSessions calculates the proportional number of sessions between startDate and currentDate
// relative to the total interval [startDate, endDate], rounded to the nearest integer session:
// round( ((currentDate - startDate) / (endDate - startDate)) * sessionsCount ).
func CalculateProportionalSessions(startDate, endDate, currentDate time.Time, sessionsCount int) int {
	if sessionsCount <= 0 {
		return 0
	}

	startNorm := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endNorm := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)
	currNorm := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC)

	if currNorm.Before(startNorm) {
		return 0
	}
	if !endNorm.After(startNorm) {
		return sessionsCount
	}
	if !currNorm.Before(endNorm) {
		return sessionsCount
	}

	elapsedDays := currNorm.Sub(startNorm).Hours() / 24.0
	totalDays := endNorm.Sub(startNorm).Hours() / 24.0
	ratio := elapsedDays / totalDays

	return int(math.Round(ratio * float64(sessionsCount)))
}

// CalculateProportionalMinutes calculates the proportional minutes between startDate and currentDate
// by first calculating the estimated number of sessions rounded to an integer and then multiplying
// by serviceMinutes: round( ((currentDate - startDate) / (endDate - startDate)) * sessionsCount ) * serviceMinutes.
func CalculateProportionalMinutes(startDate, endDate, currentDate time.Time, sessionsCount, serviceMinutes int) int {
	if serviceMinutes <= 0 {
		serviceMinutes = 60
	}
	estimatedSessions := CalculateProportionalSessions(startDate, endDate, currentDate, sessionsCount)
	if estimatedSessions <= 0 {
		return 0
	}
	return estimatedSessions * serviceMinutes
}

// CalculatePredictedMinutes calculates the predicted minutes (PredictMinutes) for the current month
// based on the customer's active plan for the service.
// For agenda plans, if an invoice already exists for the current period (invoicedSessions != nil),
// the estimated session count is capped at min(invoicedSessions, estimatedSessions) before multiplying by serviceMinutes.
func CalculatePredictedMinutes(plan *domain.Plan, serviceID int64, serviceMinutes int, doneMinutes int, currentDate time.Time, invoicedSessions *int) int {
	if plan == nil || !plan.IsActiveAt(currentDate) {
		return 0
	}
	if serviceMinutes <= 0 {
		serviceMinutes = 60
	}

	switch domain.PlanType(plan.PlanType) {
	case domain.PlanTypeNotebook, domain.PlanTypePayPerView:
		return doneMinutes

	case domain.PlanTypePackage:
		if plan.Package == nil || plan.PlanEnd == nil || plan.PlanEnd.IsZero() {
			return 0
		}
		qty := DistributeQuantityForService(plan.Items, plan.Package.Quantity, serviceID)
		return CalculateProportionalMinutes(plan.PlanStart, *plan.PlanEnd, currentDate, qty, serviceMinutes)

	case domain.PlanTypeAgenda:
		if plan.Agenda == nil {
			return 0
		}
		strategy := NewAgendaStrategy()
		sessionsCount, effStart, effEnd := strategy.CalculatePlannedSessionsForMonth(plan, serviceID, currentDate)
		if sessionsCount <= 0 {
			return 0
		}
		estimatedSessions := CalculateProportionalSessions(effStart, effEnd, currentDate, sessionsCount)
		if invoicedSessions != nil && *invoicedSessions < estimatedSessions {
			estimatedSessions = *invoicedSessions
		}
		if estimatedSessions <= 0 {
			return 0
		}
		return estimatedSessions * serviceMinutes

	default:
		return 0
	}
}

// MonthNameInPortuguese returns the lowercase Portuguese name for the given month.
func MonthNameInPortuguese(m time.Month) string {
	switch m {
	case time.January:
		return "janeiro"
	case time.February:
		return "fevereiro"
	case time.March:
		return "março"
	case time.April:
		return "abril"
	case time.May:
		return "maio"
	case time.June:
		return "junho"
	case time.July:
		return "julho"
	case time.August:
		return "agosto"
	case time.September:
		return "setembro"
	case time.October:
		return "outubro"
	case time.November:
		return "novembro"
	case time.December:
		return "dezembro"
	default:
		return ""
	}
}
