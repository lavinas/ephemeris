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
	if inv.Items[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", inv.Items[0].Quantity)
	}

	// Due date: 2026-09-10 + 5 days = 2026-09-15
	expectedDueDate := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}

	if !inv.Items[0].PeriodStart.Equal(sessDate) {
		t.Errorf("expected PeriodStart %v, got %v", sessDate, inv.Items[0].PeriodStart)
	}
	if !inv.Items[0].PeriodEnd.Equal(sessDate) {
		t.Errorf("expected PeriodEnd %v, got %v", sessDate, inv.Items[0].PeriodEnd)
	}

	expectedDesc := "10/09/2026"
	if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
		t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
	}
}

func TestPayPerViewStrategy_ItemDescriptionFormat(t *testing.T) {
	strat := NewPayPerViewStrategy()
	planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)
	sessDate := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)

	t.Run("With service description formats service.description - sessionDate", func(t *testing.T) {
		serviceDesc := "Consulta Psicológica Avulsa"
		plan := &domain.Plan{
			ID:        31,
			PlanStart: planStart,
			PlanType:  4,
			PayPerView: &domain.PlanPayPerView{
				ID:          31,
				PaymentTerm: 0,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  1,
					OrderIndex: 1,
					Price:      floatPtr(150.0),
					Service:    &domain.Service{ID: 1, Description: &serviceDesc},
				},
			},
		}

		session := domain.Session{
			ID:            301,
			CustomerID:    300,
			ServiceID:     1,
			SessionDate:   sessDate,
			SessionStatus: "realizada",
		}

		inv, err := strat.CalculateSession(plan, plan.Items[0], session, refDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "Consulta Psicológica Avulsa - 10/09/2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expected {
			t.Errorf("expected Description %q, got %v", expected, inv.Items[0].Description)
		}
	})

	t.Run("With nil service description formats only sessionDate", func(t *testing.T) {
		plan := &domain.Plan{
			ID:        32,
			PlanStart: planStart,
			PlanType:  4,
			PayPerView: &domain.PlanPayPerView{
				ID:          32,
				PaymentTerm: 0,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  1,
					OrderIndex: 1,
					Price:      floatPtr(150.0),
					Service:    nil,
				},
			},
		}

		session := domain.Session{
			ID:            302,
			CustomerID:    300,
			ServiceID:     1,
			SessionDate:   sessDate,
			SessionStatus: "realizada",
		}

		inv, err := strat.CalculateSession(plan, plan.Items[0], session, refDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "10/09/2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expected {
			t.Errorf("expected Description %q, got %v", expected, inv.Items[0].Description)
		}
	})
}
