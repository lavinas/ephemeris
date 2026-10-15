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
	expectedPeriodStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodStart.Equal(expectedPeriodStart) {
		t.Errorf("expected PeriodStart %v, got %v", expectedPeriodStart, inv.PeriodStart)
	}
	expectedPeriodEnd := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodEnd.Equal(expectedPeriodEnd) {
		t.Errorf("expected PeriodEnd %v, got %v", expectedPeriodEnd, inv.PeriodEnd)
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

func TestNotebookStrategy_DisregardsCancelledNotChargeSessions(t *testing.T) {
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
		{ID: 101, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC), SessionStatus: string(domain.StatusDone)},
		{ID: 102, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), SessionStatus: string(domain.StatusCancelCharge)},
		{ID: 103, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC), SessionStatus: string(domain.StatusCancelNotCharge)},
		{ID: 104, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC), SessionStatus: "Cancelada/Não Cobrar"},
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
	if item.InvoicedQuantity != 2 {
		t.Errorf("expected InvoicedQuantity 2 (excluding cancelada_nao_cobrar), got %d", item.InvoicedQuantity)
	}
	if item.RealizedQuantity != 2 {
		t.Errorf("expected RealizedQuantity 2, got %d", item.RealizedQuantity)
	}
	if len(item.ItemSessions) != 2 {
		t.Fatalf("expected 2 session links, got %d", len(item.ItemSessions))
	}
	if item.ItemSessions[0].SessionID != 101 || item.ItemSessions[1].SessionID != 102 {
		t.Errorf("expected session IDs 101 and 102, got %d and %d", item.ItemSessions[0].SessionID, item.ItemSessions[1].SessionID)
	}
	if inv.Amount != 160.0 {
		t.Errorf("expected Amount 160.0 (2 * 80.0), got %.2f", inv.Amount)
	}
}

func TestNotebookStrategy_AllSessionsCancelledNotCharge(t *testing.T) {
	strat := NewNotebookStrategy()

	planStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	planWithoutFixedPrice := &domain.Plan{
		ID:         20,
		CustomerID: 200,
		PlanStart:  planStart,
		PlanType:   3,
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(80.0)},
		},
	}

	sessions := []domain.Session{
		{ID: 101, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC), SessionStatus: string(domain.StatusCancelNotCharge)},
		{ID: 102, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), SessionStatus: "Cancelada/Não Cobrar"},
	}

	sessionsByService := map[int64][]domain.Session{
		1: sessions,
	}

	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	inv, err := strat.CalculateMonth(planWithoutFixedPrice, monthStart, sessionsByService, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv != nil {
		t.Errorf("expected nil invoice when all sessions are cancelada_nao_cobrar and no fixed plan price, got %+v", inv)
	}

	// With fixed price
	fixedPrice := 200.0
	planWithFixedPrice := &domain.Plan{
		ID:         21,
		CustomerID: 200,
		PlanStart:  planStart,
		PlanType:   3,
		Price:      &fixedPrice,
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(80.0)},
		},
	}

	invFixed, err := strat.CalculateMonth(planWithFixedPrice, monthStart, sessionsByService, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invFixed == nil {
		t.Fatal("expected invoice when plan has fixed price, got nil")
	}
	if invFixed.Amount != 200.0 {
		t.Errorf("expected Amount 200.0, got %.2f", invFixed.Amount)
	}
	if invFixed.Items[0].InvoicedQuantity != 0 {
		t.Errorf("expected InvoicedQuantity 0, got %d", invFixed.Items[0].InvoicedQuantity)
	}
}

func TestNotebookStrategy_ItemDescriptionFormat(t *testing.T) {
	strat := NewNotebookStrategy()

	serviceDesc := "Atendimento Psicológico"
	planWithServiceDesc := &domain.Plan{
		ID:        20,
		PlanStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PlanType:  3,
		Items: []domain.PlanItem{
			{
				ServiceID:  1,
				OrderIndex: 1,
				Price:      floatPtr(100.0),
				Service:    &domain.Service{ID: 1, Description: &serviceDesc},
			},
		},
	}

	planWithoutServiceDesc := &domain.Plan{
		ID:        21,
		PlanStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PlanType:  3,
		Items: []domain.PlanItem{
			{
				ServiceID:  1,
				OrderIndex: 1,
				Price:      floatPtr(100.0),
				Service:    nil,
			},
		},
	}

	sessions := []domain.Session{
		{ID: 101, CustomerID: 200, ServiceID: 1, SessionDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC), SessionStatus: string(domain.StatusDone)},
	}
	sessionsByService := map[int64][]domain.Session{1: sessions}

	// Calculation month is September 2026 -> considered sessions are from August 2026
	monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	t.Run("With service description includes description and month/year", func(t *testing.T) {
		inv, err := strat.CalculateMonth(planWithServiceDesc, monthStart, sessionsByService, refDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "Atendimento Psicológico - agosto de 2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})

	t.Run("With nil service description formats only month and year without hyphen", func(t *testing.T) {
		inv, err := strat.CalculateMonth(planWithoutServiceDesc, monthStart, sessionsByService, refDate)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "agosto de 2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})
}


