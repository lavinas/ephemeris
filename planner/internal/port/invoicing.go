package port

import (
	"time"

	"planner/internal/domain"
)

// InvoicingStrategy defines the interface that all plan-specific invoicing strategy adapters must implement.
type InvoicingStrategy interface {
	Process(plan *domain.Plan, startDate, endDate time.Time, reprocess bool) error
}
