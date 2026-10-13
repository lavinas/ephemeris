package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestNotebookStrategy_CalculateMonth(t *testing.T) {
	strat := NewNotebookStrategy()

	planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:         20,
		CustomerID: 200,
		PlanStart:  planStart,
		PlanType:   3,
		Notebook: &domain.PlanNotebook{
			ID:         20,
			PaymentDay: 10,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(80.0)},
		},
	}

	sessions := []domain.Session{
		{ID: 101, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)},
		{ID: 102, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC)},
		{ID: 103, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)},
	}

	sessionsByService := map[int64][]domain.Session{
		1: sessions,
	}

	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	inv, err := strat.CalculateMonth(plan, monthStart, sessionsByService, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv == nil {
		t.Fatal("expected invoice, got nil")
	}

	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}

	item := inv.Items[0]
	if item.InvoicedQuantity != 3 {
		t.Errorf("expected InvoicedQuantity 3, got %d", item.InvoicedQuantity)
	}
	if item.RealizedQuantity != 3 {
		t.Errorf("expected RealizedQuantity 3, got %d", item.RealizedQuantity)
	}
	if len(item.ItemSessions) != 3 {
		t.Errorf("expected 3 session links, got %d", len(item.ItemSessions))
	}
	if inv.Amount != 240.0 {
		t.Errorf("expected Amount 240.0, got %.2f", inv.Amount)
	}
}

func TestNotebookStrategy_NoSessionsAndNoFixedPrice(t *testing.T) {
	strat := NewNotebookStrategy()

	plan := &domain.Plan{
		ID:        20,
		PlanStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PlanType:  3,
		Items:     []domain.PlanItem{{ServiceID: 1, Price: floatPtr(50.0)}},
	}

	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	sessionsByService := map[int64][]domain.Session{1: {}}

	inv, err := strat.CalculateMonth(plan, monthStart, sessionsByService, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv != nil {
		t.Errorf("expected nil invoice when 0 sessions and no fixed plan price, got %+v", inv)
	}
}
