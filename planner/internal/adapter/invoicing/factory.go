package invoicing

import (
	"fmt"

	"planner/internal/domain"
	"planner/internal/port"
)

// StrategyFactory creates the appropriate InvoicingStrategy adapter based on the plan type.
type StrategyFactory struct {
	repo   port.Repository
	logger port.Logger
}

// NewStrategyFactory creates a new StrategyFactory.
func NewStrategyFactory(repo port.Repository, logger port.Logger) *StrategyFactory {
	return &StrategyFactory{
		repo:   repo,
		logger: logger,
	}
}

// GetStrategy returns the invoicing strategy adapter for the given plan type.
func (f *StrategyFactory) GetStrategy(planType domain.PlanType) (port.InvoicingStrategy, error) {
	switch planType {
	case domain.PlanTypeAgenda:
		return NewAgendaStrategy(f.repo, f.logger), nil
	case domain.PlanTypePackage:
		return NewPackageStrategy(f.repo, f.logger), nil
	case domain.PlanTypeNotebook:
		return NewNotebookStrategy(f.repo, f.logger), nil
	case domain.PlanTypePayPerView:
		return NewPayPerViewStrategy(f.repo, f.logger), nil
	default:
		return nil, fmt.Errorf("tipo de plano desconhecido ou não suportado: %d", planType)
	}
}
