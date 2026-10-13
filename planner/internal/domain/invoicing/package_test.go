package invoicing

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func floatPtr(v float64) *float64 { return &v }
func stringPtr(v string) *string  { return &v }
func intPtr(v int) *int          { return &v }

func TestPackageStrategy_Calculate(t *testing.T) {
	strat := NewPackageStrategy()

	planStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	paymentDate := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	refDate := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)

	desc1 := "Aula Canto 30"
	desc2 := "Aula Piano 45"
	desc3 := "Aula Teoria"

	plan := &domain.Plan{
		ID:         10,
		CustomerID: 100,
		PlanStart:  planStart,
		PlanType:   2,
		Package: &domain.PlanPackage{
			ID:          10,
			Quantity:    10,
			PaymentDate: paymentDate,
		},
		Items: []domain.PlanItem{
			{ServiceID: 1, OrderIndex: 1, Price: floatPtr(50.0), Service: &domain.Service{Description: &desc1}},
			{ServiceID: 2, OrderIndex: 2, Price: floatPtr(60.0), Service: &domain.Service{Description: &desc2}},
			{ServiceID: 3, OrderIndex: 3, Price: floatPtr(40.0), Service: &domain.Service{Description: &desc3}},
		},
	}

	inv, err := strat.Calculate(plan, refDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inv.PlanID != 10 {
		t.Errorf("expected PlanID 10, got %d", inv.PlanID)
	}
	if !inv.InvoicingDate.Equal(planStart) {
		t.Errorf("expected InvoicingDate %v, got %v", planStart, inv.InvoicingDate)
	}
	if !inv.DueDate.Equal(paymentDate) {
		t.Errorf("expected DueDate %v, got %v", paymentDate, inv.DueDate)
	}
	if len(inv.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(inv.Items))
	}

	// 10 distributed among 3 items: 4, 3, 3
	if inv.Items[0].InvoicedQuantity != 4 {
		t.Errorf("expected item 1 quantity 4, got %d", inv.Items[0].InvoicedQuantity)
	}
	if inv.Items[1].InvoicedQuantity != 3 {
		t.Errorf("expected item 2 quantity 3, got %d", inv.Items[1].InvoicedQuantity)
	}
	if inv.Items[2].InvoicedQuantity != 3 {
		t.Errorf("expected item 3 quantity 3, got %d", inv.Items[2].InvoicedQuantity)
	}

	// Amount: 4*50 + 3*60 + 3*40 = 500
	if inv.Amount != 500.0 {
		t.Errorf("expected Amount 500.0, got %.2f", inv.Amount)
	}
}

func TestPackageStrategy_Validations(t *testing.T) {
	strat := NewPackageStrategy()
	refDate := time.Now()

	t.Run("nil plan", func(t *testing.T) {
		_, err := strat.Calculate(nil, refDate)
		if err == nil {
			t.Fatal("expected error for nil plan")
		}
	})

	t.Run("nil package", func(t *testing.T) {
		_, err := strat.Calculate(&domain.Plan{ID: 1}, refDate)
		if err == nil {
			t.Fatal("expected error for nil package")
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := strat.Calculate(&domain.Plan{
			ID:      1,
			Package: &domain.PlanPackage{Quantity: 5},
		}, refDate)
		if err == nil {
			t.Fatal("expected error for empty items")
		}
	})

	t.Run("quantity zero or negative", func(t *testing.T) {
		_, err := strat.Calculate(&domain.Plan{
			ID:      1,
			Package: &domain.PlanPackage{Quantity: 0},
			Items:   []domain.PlanItem{{ServiceID: 1}},
		}, refDate)
		if err == nil {
			t.Fatal("expected error for quantity <= 0")
		}
	})

	t.Run("more items than quantity", func(t *testing.T) {
		_, err := strat.Calculate(&domain.Plan{
			ID:      1,
			Package: &domain.PlanPackage{Quantity: 1},
			Items:   []domain.PlanItem{{ServiceID: 1}, {ServiceID: 2}},
		}, refDate)
		if err == nil {
			t.Fatal("expected error when items > quantity")
		}
	})
}
