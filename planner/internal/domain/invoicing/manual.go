package invoicing

import (
	"planner/internal/domain"
)

// ManualStrategy represents the passive invoicing strategy for manual plans.
// As defined in adjust3.md, this strategy takes no action during plan invoice generation.
type ManualStrategy struct{}

// NewManualStrategy creates a new ManualStrategy instance.
func NewManualStrategy() *ManualStrategy {
	return &ManualStrategy{}
}

// Generate does nothing and returns nil because Manual plans are passive.
func (s *ManualStrategy) Generate(plan *domain.Plan) (*domain.Invoice, error) {
	return nil, nil
}
