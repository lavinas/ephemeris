package domain

import (
	"testing"
	"time"
)

func TestNormalizeMonth(t *testing.T) {
	date := time.Date(2026, time.April, 15, 14, 30, 0, 0, time.UTC)
	normalized := NormalizeMonth(date)

	if normalized.Year() != 2026 || normalized.Month() != time.April || normalized.Day() != 1 {
		t.Errorf("expected 2026-04-01, got %v", normalized)
	}
	if normalized.Hour() != 0 || normalized.Minute() != 0 || normalized.Second() != 0 {
		t.Errorf("expected time zeroed, got %v", normalized)
	}
}

func TestEffectiveSessionMinutes(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		status   string
		expected int
	}{
		{"regular completed session", 60, string(StatusDone), 60},
		{"cancelled with charge", 45, string(StatusCancelCharge), 45},
		{"cancelled without charge", 50, string(StatusCancelNotCharge), 0},
		{"zero minutes session", 0, string(StatusDone), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EffectiveSessionMinutes(tt.minutes, tt.status)
			if got != tt.expected {
				t.Errorf("EffectiveSessionMinutes(%d, %q) = %d; expected %d",
					tt.minutes, tt.status, got, tt.expected)
			}
		})
	}
}

func TestComptimeBalanceCalculation(t *testing.T) {
	c := NewComptime(1, 10, 20, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC))
	if c.Balance != 0 || c.InvoicedMinutes != 0 || c.ConsumedMinutes != 0 || c.AdjustedMinutes != 0 {
		t.Errorf("new comptime should have zero values, got %+v", c)
	}

	c.InvoicedMinutes = 120
	c.ConsumedMinutes = 50
	c.AdjustedMinutes = -20
	c.RecalculateBalance(10) // previous month had 10 minutes balance

	expectedBalance := 10 + 120 - 50 - 20 // 60
	if c.Balance != expectedBalance {
		t.Errorf("expected balance %d, got %d", expectedBalance, c.Balance)
	}
	if c.MonthBalance() != 50 {
		t.Errorf("expected month balance 50, got %d", c.MonthBalance())
	}
}

func TestComptimeEntitiesTableNames(t *testing.T) {
	if (Comptime{}).TableName() != "comptime" {
		t.Errorf("unexpected table name for Comptime: %s", (Comptime{}).TableName())
	}
	if (ComptimeSession{}).TableName() != "comptime_session" {
		t.Errorf("unexpected table name for ComptimeSession: %s", (ComptimeSession{}).TableName())
	}
	if (ComptimeInvoiceItem{}).TableName() != "comptime_invoice_item" {
		t.Errorf("unexpected table name for ComptimeInvoiceItem: %s", (ComptimeInvoiceItem{}).TableName())
	}
	if (ComptimeAdjust{}).TableName() != "comptime_adjust" {
		t.Errorf("unexpected table name for ComptimeAdjust: %s", (ComptimeAdjust{}).TableName())
	}
}
