package domain

import (
	"sort"
	"strings"
	"time"
)

// Service represents the service domain model.
type Service struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	VendorID       int64     `gorm:"not null;index"`
	Name           string    `gorm:"not null"`
	Description    *string   `gorm:"type:text"`
	SessionMinutes *int      `gorm:"column:session_minutes"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
}

// NewService creates a new Service instance with the provided details.
func NewService(vendorID int64, name string, description *string, sessionMinutes *int) *Service {
	return &Service{
		VendorID:       vendorID,
		Name:           name,
		Description:    description,
		SessionMinutes: sessionMinutes,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// TableName specifies the table name for Service model.
func (Service) TableName() string {
	return "service"
}

// SortServicesByName returns a new slice of services sorted alphabetically by Name (case-insensitive, then ID).
func SortServicesByName(services []Service) []Service {
	if len(services) == 0 {
		return nil
	}
	sorted := make([]Service, len(services))
	copy(sorted, services)
	sort.Slice(sorted, func(i, j int) bool {
		nameI := strings.ToLower(strings.TrimSpace(sorted[i].Name))
		nameJ := strings.ToLower(strings.TrimSpace(sorted[j].Name))
		if nameI != nameJ {
			return nameI < nameJ
		}
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

// ResolveServicesForCustomerPlans determines the services available for a customer based on their plans:
// - If the customer has one or more active plans at refDate (with available services), returns only the services in those plans, ordered by name.
// - If the customer has no active plans (or no services in active plans), returns all available services, ordered by name.
func ResolveServicesForCustomerPlans(plans []Plan, allServices []Service, refDate time.Time) []Service {
	serviceMap := make(map[int64]Service, len(allServices))
	for _, s := range allServices {
		serviceMap[s.ID] = s
	}

	seen := make(map[int64]bool)
	var planServices []Service
	for i := range plans {
		plan := &plans[i]
		if !plan.IsActiveAt(refDate) {
			continue
		}
		for _, item := range plan.Items {
			if item.DeletedAt != nil || item.ServiceID <= 0 || seen[item.ServiceID] {
				continue
			}
			var svc Service
			var found bool
			if item.Service != nil && item.Service.ID > 0 && item.Service.Name != "" {
				svc = *item.Service
				found = true
			} else if s, ok := serviceMap[item.ServiceID]; ok {
				svc = s
				found = true
			}
			if found {
				seen[item.ServiceID] = true
				planServices = append(planServices, svc)
			}
		}
	}

	if len(planServices) > 0 {
		return SortServicesByName(planServices)
	}
	return SortServicesByName(allServices)
}
