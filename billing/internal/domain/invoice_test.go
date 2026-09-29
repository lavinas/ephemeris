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
