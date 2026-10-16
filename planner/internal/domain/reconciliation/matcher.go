package reconciliation

import (
	"strings"
	"time"

	"planner/internal/domain"
)

// IsSessionEligible checks if a session is valid and billable for reconciliation.
func IsSessionEligible(session *domain.Session) bool {
	if session == nil {
		return false
	}
	if session.DeletedAt != nil {
		return false
	}
	status := strings.ToLower(strings.TrimSpace(session.SessionStatus))
	if status == strings.ToLower(string(domain.StatusCancelNotCharge)) || status == "cancelada/não cobrar" || status == "cancelada_nao_cobrar" {
		return false
	}
	return true
}

// MatchesInvoiceItem checks if a session matches an invoice item and period.
func MatchesInvoiceItem(session *domain.Session, customerID int64, item *domain.InvoiceItem, periodStart, periodEnd time.Time) bool {
	if session == nil || item == nil {
		return false
	}
	if !IsSessionEligible(session) {
		return false
	}
	if session.CustomerID != customerID {
		return false
	}
	if session.ServiceID != item.ServiceID {
		return false
	}

	sessionDate := time.Date(session.SessionDate.Year(), session.SessionDate.Month(), session.SessionDate.Day(), 0, 0, 0, 0, time.UTC)
	pStart := time.Date(periodStart.Year(), periodStart.Month(), periodStart.Day(), 0, 0, 0, 0, time.UTC)
	pEnd := time.Date(periodEnd.Year(), periodEnd.Month(), periodEnd.Day(), 23, 59, 59, 999999999, time.UTC)

	return !sessionDate.Before(pStart) && !sessionDate.After(pEnd)
}
