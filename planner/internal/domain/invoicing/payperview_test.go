package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestPayPerViewStrategy_CalculateSession(t *testing.T) {
	strat := NewPayPerViewStrategy()

	planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:         30,
		CustomerID: 300,
		PlanStart:  planStart,
		PlanType:   4,
		PayPerView: &domain.PlanPayPerView{
			ID:          30,
			PaymentTerm: 5,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(120.0)},
		},
	}

	sessDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	session := domain.Session{
		ID:            201,
		CustomerID:    300,
		ServiceID:     1,
		SessionDate:   sessDate,
		SessionStatus: "realizada",
	}

	refDate := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)

	inv, err := strat.CalculateSession(plan, plan.Items[0], session, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.Amount != 120.0 {
		t.Errorf("expected Amount 120.0, got %.2f", inv.Amount)
	}
	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}
	if inv.Items[0].InvoicedQuantity != 1 || inv.Items[0].RealizedQuantity != 1 {
		t.Errorf("expected quantities 1, got inv=%d, real=%d",
			inv.Items[0].InvoicedQuantity, inv.Items[0].RealizedQuantity)
	}
	if len(inv.Items[0].ItemSessions) != 1 || inv.Items[0].ItemSessions[0].SessionID != 201 {
		t.Errorf("expected session link to 201, got %v", inv.Items[0].ItemSessions)
	}

	// Due date: 2026-09-10 + 5 days = 2026-09-15
	expectedDueDate := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}
}
