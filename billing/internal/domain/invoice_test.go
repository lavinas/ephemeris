package domain

import (
	"testing"
	"time"
)

func TestInvoice_IsOverdue(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)

	paidDate := today

	tests := []struct {
		name        string
		dueDate     time.Time
		paymentDate *time.Time
		wantOverdue bool
	}{
		{
			name:        "due yesterday, unpaid -> overdue",
			dueDate:     yesterday,
			paymentDate: nil,
			wantOverdue: true,
		},
		{
			name:        "due today (00:00), unpaid -> overdue",
			dueDate:     today,
			paymentDate: nil,
			wantOverdue: true,
		},
		{
			name:        "due tomorrow, unpaid -> not overdue",
			dueDate:     tomorrow,
			paymentDate: nil,
			wantOverdue: false,
		},
		{
			name:        "due yesterday, paid -> not overdue",
			dueDate:     yesterday,
			paymentDate: &paidDate,
			wantOverdue: false,
		},
		{
			name:        "due today, paid -> not overdue",
			dueDate:     today,
			paymentDate: &paidDate,
			wantOverdue: false,
		},
		{
			name:        "due tomorrow, paid -> not overdue",
			dueDate:     tomorrow,
			paymentDate: &paidDate,
			wantOverdue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &Invoice{
				DueDate:     tt.dueDate,
				PaymentDate: tt.paymentDate,
			}
			if got := inv.IsOverdue(); got != tt.wantOverdue {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.wantOverdue)
			}
		})
	}
}

func TestInvoice_CanDelete(t *testing.T) {
	now := time.Now()

	t.Run("all nil -> can delete", func(t *testing.T) {
		inv := &Invoice{}
		if !inv.CanDelete() {
			t.Errorf("expected CanDelete() = true, got false")
		}
	})

	t.Run("PaymentDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{PaymentDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("EmailSentDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{EmailSentDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("WhatsappSentDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{WhatsappSentDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("EmailReceiptDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{EmailReceiptDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("WhatsappReceiptDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{WhatsappReceiptDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("TaxDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{TaxDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})
}

func TestInvoice_CanPay(t *testing.T) {
	now := time.Now()

	t.Run("PaymentDate nil -> can pay", func(t *testing.T) {
		inv := &Invoice{PaymentDate: nil}
		if !inv.CanPay() {
			t.Errorf("expected CanPay() = true, got false")
		}
	})

	t.Run("PaymentDate not nil -> cannot pay", func(t *testing.T) {
		inv := &Invoice{PaymentDate: &now}
		if inv.CanPay() {
			t.Errorf("expected CanPay() = false, got true")
		}
	})
}

