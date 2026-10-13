package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestAgendaStrategy_CalculateMonth(t *testing.T) {
	strat := NewAgendaStrategy()

	planStart := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	// Pre-paid, Weekly on Tuesdays (WeekDay 3 in 1..7: Tuesday=3)
	// Payment day: 5th of month
	plan := &domain.Plan{
		ID:         40,
		CustomerID: 400,
		PlanStart:  planStart,
		PlanType:   1,
		Agenda: &domain.PlanAgenda{
			ID:                  40,
			Recurrence:          int(domain.RecurrenceWeekly),
			WeekDay:             int(domain.WeekDayTuesday),
			TermType:            int(domain.TermTypePrePaid),
			PaymentDay:          5,
			MonthlyServiceLimit: intPtr(4),
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(75.0)},
		},
	}

	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	inv, err := strat.CalculateMonth(plan, monthStart, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv == nil {
		t.Fatal("expected invoice, got nil")
	}

	// Invoicing date: min(2026-09-01, 2026-08-15) -> 2026-08-15
	if !inv.InvoicingDate.Equal(planStart) {
		t.Errorf("expected InvoicingDate %v, got %v", planStart, inv.InvoicingDate)
	}

	expectedDueDate := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}

	if len(inv.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(inv.Items))
	}

	// Sept 2026 has 5 Tuesdays, but limit is 4
	if inv.Items[0].InvoicedQuantity != 4 {
		t.Errorf("expected InvoicedQuantity 4, got %d", inv.Items[0].InvoicedQuantity)
	}
	if inv.Amount != 300.0 {
		t.Errorf("expected Amount 300.0, got %.2f", inv.Amount)
	}
}

func TestAgendaStrategy_PostPaid(t *testing.T) {
	strat := NewAgendaStrategy()

	planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:         41,
		CustomerID: 400,
		PlanStart:  planStart,
		PlanType:   1,
		Agenda: &domain.PlanAgenda{
			ID:         41,
			Recurrence: int(domain.RecurrenceWeekly),
			WeekDay:    int(domain.WeekDayMonday), // Monday=2
			TermType:   int(domain.TermTypePostPaid),
			PaymentDay: 10,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(100.0)},
		},
	}

	// Month: September 2026. For post-paid, calculates sessions in August 2026.
	// August 2026 Mondays: Aug 3, 10, 17, 24, 31 (5 Mondays).
	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	inv, err := strat.CalculateMonth(plan, monthStart, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv == nil {
		t.Fatal("expected invoice, got nil")
	}

	if inv.Items[0].InvoicedQuantity != 5 {
		t.Errorf("expected InvoicedQuantity 5, got %d", inv.Items[0].InvoicedQuantity)
	}
	if inv.Amount != 500.0 {
		t.Errorf("expected Amount 500.0, got %.2f", inv.Amount)
	}
}

func TestAgendaStrategy_InactiveMonth(t *testing.T) {
	strat := NewAgendaStrategy()

	planStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	plan := &domain.Plan{
		ID:        42,
		PlanStart: planStart,
		Agenda: &domain.PlanAgenda{
			Recurrence: int(domain.RecurrenceWeekly),
			WeekDay:    int(domain.WeekDayMonday),
			PaymentDay: 5,
		},
		Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1}},
	}

	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Now()

	inv, err := strat.CalculateMonth(plan, monthStart, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv != nil {
		t.Errorf("expected nil invoice for inactive month, got %+v", inv)
	}
}
