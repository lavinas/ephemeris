package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestManualStrategy_Generate(t *testing.T) {
	strat := NewManualStrategy()
	plan := domain.NewPlan(1, time.Now(), nil, int(domain.PlanTypeManual), nil)

	inv, err := strat.Generate(plan)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if inv != nil {
		t.Fatalf("expected nil invoice for passive manual strategy, got %+v", inv)
	}
}
