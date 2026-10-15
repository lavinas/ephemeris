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

	// Invoicing date: max(2026-09-01, 2026-08-15) -> 2026-09-01
	expectedInvoicingDate := monthStart
	if !inv.InvoicingDate.Equal(expectedInvoicingDate) {
		t.Errorf("expected InvoicingDate %v, got %v", expectedInvoicingDate, inv.InvoicingDate)
	}

	expectedDueDate := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
	if !inv.DueDate.Equal(expectedDueDate) {
		t.Errorf("expected DueDate %v, got %v", expectedDueDate, inv.DueDate)
	}

	expectedPeriodStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodStart.Equal(expectedPeriodStart) {
		t.Errorf("expected PeriodStart %v, got %v", expectedPeriodStart, inv.PeriodStart)
	}
	expectedPeriodEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodEnd.Equal(expectedPeriodEnd) {
		t.Errorf("expected PeriodEnd %v, got %v", expectedPeriodEnd, inv.PeriodEnd)
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

	expectedPeriodStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodStart.Equal(expectedPeriodStart) {
		t.Errorf("expected PeriodStart %v, got %v", expectedPeriodStart, inv.PeriodStart)
	}
	expectedPeriodEnd := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	if !inv.PeriodEnd.Equal(expectedPeriodEnd) {
		t.Errorf("expected PeriodEnd %v, got %v", expectedPeriodEnd, inv.PeriodEnd)
	}

	if inv.Items[0].InvoicedQuantity != 5 {
		t.Errorf("expected InvoicedQuantity 5, got %d", inv.Items[0].InvoicedQuantity)
	}
	if inv.Amount != 500.0 {
		t.Errorf("expected Amount 500.0, got %.2f", inv.Amount)
	}

	// For post-paid, calculation month is September 2026 (monthStart = 2026-09-01).
	// Invoicing date = max(monthStart, planStart) -> max(2026-09-01, 2026-08-01) = 2026-09-01
	expectedInvoicingDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if !inv.InvoicingDate.Equal(expectedInvoicingDate) {
		t.Errorf("expected InvoicingDate %v, got %v", expectedInvoicingDate, inv.InvoicingDate)
	}
}

func TestAgendaStrategy_InvoicingDateRule(t *testing.T) {
	strat := NewAgendaStrategy()

	t.Run("Pre-paid: planStart in middle of month is greater than first day of current month", func(t *testing.T) {
		planStart := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
		plan := &domain.Plan{
			ID:        51,
			PlanStart: planStart,
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePrePaid),
				PaymentDay: 5,
			},
			Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0)}},
		}
		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// max(2026-09-01, 2026-09-15) -> 2026-09-15
		expected := planStart
		if !inv.InvoicingDate.Equal(expected) {
			t.Errorf("expected InvoicingDate %v, got %v", expected, inv.InvoicingDate)
		}
	})

	t.Run("Pre-paid: planStart before current month is smaller than first day of current month", func(t *testing.T) {
		planStart := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
		plan := &domain.Plan{
			ID:        52,
			PlanStart: planStart,
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePrePaid),
				PaymentDay: 5,
			},
			Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0)}},
		}
		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// max(2026-09-01, 2026-08-20) -> 2026-09-01
		expected := monthStart
		if !inv.InvoicingDate.Equal(expected) {
			t.Errorf("expected InvoicingDate %v, got %v", expected, inv.InvoicingDate)
		}
	})

	t.Run("Post-paid: plan starting mid month is not active in that same month, billed in subsequent month", func(t *testing.T) {
		planStart := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
		plan := &domain.Plan{
			ID:        53,
			PlanStart: planStart,
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePostPaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0)}},
		}

		// In September 2026, target month is August 2026 (plan was not active yet) -> nil invoice
		monthStartSept := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		invSept, err := strat.CalculateMonth(plan, monthStartSept, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invSept != nil {
			t.Errorf("expected nil invoice for post-paid plan in its start month, got %+v", invSept)
		}

		// In October 2026, target month is September 2026 (plan active from 15/09 to 30/09)
		// September 2026 Mondays after 15th: Sept 21, Sept 28 -> 2 Mondays
		monthStartOct := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		invOct, err := strat.CalculateMonth(plan, monthStartOct, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invOct == nil {
			t.Fatal("expected invoice in October for September sessions, got nil")
		}
		if invOct.Items[0].InvoicedQuantity != 2 {
			t.Errorf("expected 2 invoiced sessions, got %d", invOct.Items[0].InvoicedQuantity)
		}
		expectedDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
		if !invOct.InvoicingDate.Equal(expectedDate) {
			t.Errorf("expected InvoicingDate %v, got %v", expectedDate, invOct.InvoicingDate)
		}
	})

	t.Run("Post-paid: calculation month is current month; planStart prior to current month", func(t *testing.T) {
		planStart := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
		plan := &domain.Plan{
			ID:        54,
			PlanStart: planStart,
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePostPaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0)}},
		}
		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// max(2026-09-01, 2026-07-10) -> 2026-09-01
		expected := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		if !inv.InvoicingDate.Equal(expected) {
			t.Errorf("expected InvoicingDate %v, got %v", expected, inv.InvoicingDate)
		}
	})

	t.Run("Post-paid: plan ends in previous month (February), billed in calculation month (March)", func(t *testing.T) {
		planStart := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
		planEnd := time.Date(2027, 2, 5, 0, 0, 0, 0, time.UTC)
		plan := &domain.Plan{
			ID:        80,
			PlanStart: planStart,
			PlanEnd:   &planEnd,
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday), // Monday=2
				TermType:   int(domain.TermTypePostPaid),
				PaymentDay: 15,
			},
			Items: []domain.PlanItem{{ServiceID: 5, OrderIndex: 1, Price: floatPtr(100.0)}},
		}

		// March 2027 invoicing: bills February 2027 sessions (2027-02-01 is Monday, <= 2027-02-05 -> 1 session)
		monthStartMar := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
		invMar, err := strat.CalculateMonth(plan, monthStartMar, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invMar == nil {
			t.Fatal("expected invoice in March for February session, got nil")
		}
		if invMar.Items[0].InvoicedQuantity != 1 {
			t.Errorf("expected 1 invoiced session (01/02/2027), got %d", invMar.Items[0].InvoicedQuantity)
		}
		if invMar.Amount != 100.0 {
			t.Errorf("expected Amount 100.0, got %.2f", invMar.Amount)
		}
		expectedInvoicingDate := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
		if !invMar.InvoicingDate.Equal(expectedInvoicingDate) {
			t.Errorf("expected InvoicingDate %v, got %v", expectedInvoicingDate, invMar.InvoicingDate)
		}
		expectedDueDate := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC)
		if !invMar.DueDate.Equal(expectedDueDate) {
			t.Errorf("expected DueDate %v, got %v", expectedDueDate, invMar.DueDate)
		}

		// April 2027 invoicing: bills March 2027 sessions (plan already ended on 2027-02-05 -> inactive, no invoice)
		monthStartApr := time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC)
		invApr, err := strat.CalculateMonth(plan, monthStartApr, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invApr != nil {
			t.Errorf("expected nil invoice in April because plan ended in February, got %+v", invApr)
		}
	})
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

func TestAgendaStrategy_ItemDescriptionFormat(t *testing.T) {
	strat := NewAgendaStrategy()

	t.Run("Pre-paid with service description includes current month and year", func(t *testing.T) {
		serviceDesc := "Aula de piano online"
		plan := &domain.Plan{
			ID:        101,
			PlanStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePrePaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  1,
					OrderIndex: 1,
					Price:      floatPtr(100.0),
					Service:    &domain.Service{ID: 1, Description: &serviceDesc},
				},
			},
		}

		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "Aula de piano online - setembro de 2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})

	t.Run("Pre-paid with nil service description formats only month and year", func(t *testing.T) {
		plan := &domain.Plan{
			ID:        102,
			PlanStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePrePaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  1,
					OrderIndex: 1,
					Price:      floatPtr(100.0),
					Service:    nil,
				},
			},
		}

		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "setembro de 2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})

	t.Run("Post-paid with service description includes previous month and year", func(t *testing.T) {
		serviceDesc := "Aula de canto online"
		plan := &domain.Plan{
			ID:        103,
			PlanStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePostPaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  2,
					OrderIndex: 1,
					Price:      floatPtr(100.0),
					Service:    &domain.Service{ID: 2, Description: &serviceDesc},
				},
			},
		}

		// Calculation month is September 2026 -> previous month (sessions) is August 2026
		monthStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "Aula de canto online - agosto de 2026"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})

	t.Run("Post-paid with nil service description formats only previous month and year", func(t *testing.T) {
		plan := &domain.Plan{
			ID:        104,
			PlanStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			PlanType:  1,
			Agenda: &domain.PlanAgenda{
				Recurrence: int(domain.RecurrenceWeekly),
				WeekDay:    int(domain.WeekDayMonday),
				TermType:   int(domain.TermTypePostPaid),
				PaymentDay: 10,
			},
			Items: []domain.PlanItem{
				{
					ServiceID:  2,
					OrderIndex: 1,
					Price:      floatPtr(100.0),
					Service:    nil,
				},
			},
		}

		// Calculation month is March 2027 -> previous month (sessions) is February 2027
		monthStart := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
		inv, err := strat.CalculateMonth(plan, monthStart, time.Now())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv == nil || len(inv.Items) == 0 {
			t.Fatal("expected invoice with items")
		}
		expectedDesc := "fevereiro de 2027"
		if inv.Items[0].Description == nil || *inv.Items[0].Description != expectedDesc {
			t.Errorf("expected Description %q, got %v", expectedDesc, inv.Items[0].Description)
		}
	})
}
