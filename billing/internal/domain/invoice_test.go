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
			name:        "due today (00:00), unpaid -> not overdue (until end of day)",
			dueDate:     today,
			paymentDate: nil,
			wantOverdue: false,
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

	t.Run("TaxDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{TaxDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})

	t.Run("CancellationDate not nil -> cannot delete", func(t *testing.T) {
		inv := &Invoice{CancellationDate: &now}
		if inv.CanDelete() {
			t.Errorf("expected CanDelete() = false, got true")
		}
	})
}

func TestInvoice_CanEdit(t *testing.T) {
	now := time.Now()

	t.Run("CancellationDate nil -> can edit", func(t *testing.T) {
		inv := &Invoice{}
		if !inv.CanEdit() {
			t.Errorf("expected CanEdit() = true, got false")
		}
	})

	t.Run("CancellationDate not nil -> cannot edit", func(t *testing.T) {
		inv := &Invoice{CancellationDate: &now}
		if inv.CanEdit() {
			t.Errorf("expected CanEdit() = false, got true")
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

func TestInvoice_HasNBS(t *testing.T) {
	nbs := int64(122051100)

	t.Run("no items -> false", func(t *testing.T) {
		inv := &Invoice{}
		if inv.HasNBS() {
			t.Errorf("expected HasNBS() = false, got true")
		}
	})

	t.Run("item with nil NBS -> false", func(t *testing.T) {
		inv := &Invoice{
			InvoiceItems: []InvoiceItem{
				{Description: "Item 1", NBSCode: nil},
			},
		}
		if inv.HasNBS() {
			t.Errorf("expected HasNBS() = false, got true")
		}
	})

	t.Run("all items with valid NBS -> true", func(t *testing.T) {
		inv := &Invoice{
			InvoiceItems: []InvoiceItem{
				{Description: "Item 1", NBSCode: &nbs},
			},
		}
		if !inv.HasNBS() {
			t.Errorf("expected HasNBS() = true, got false")
		}
	})
}

func TestInvoice_IsPF(t *testing.T) {
	cpf := "123.456.789-01"
	cnpj := "12.345.678/0001-90"
	cleanCPF := "12345678901"

	t.Run("CPF with mask -> true", func(t *testing.T) {
		inv := &Invoice{Customer: Customer{Document: &cpf}}
		if !inv.IsPF() {
			t.Errorf("expected IsPF() = true for CPF")
		}
	})

	t.Run("clean CPF -> true", func(t *testing.T) {
		inv := &Invoice{Customer: Customer{Document: &cleanCPF}}
		if !inv.IsPF() {
			t.Errorf("expected IsPF() = true for clean CPF")
		}
	})

	t.Run("CNPJ -> false", func(t *testing.T) {
		inv := &Invoice{Customer: Customer{Document: &cnpj}}
		if inv.IsPF() {
			t.Errorf("expected IsPF() = false for CNPJ")
		}
	})

	t.Run("nil document -> false", func(t *testing.T) {
		inv := &Invoice{Customer: Customer{Document: nil}}
		if inv.IsPF() {
			t.Errorf("expected IsPF() = false for nil doc")
		}
	})
}

func TestInvoice_CanTax(t *testing.T) {
	now := time.Now()
	cpf := "123.456.789-01"
	cnpj := "12.345.678/0001-90"
	nbs := int64(122051100)

	t.Run("valid paid PF with NBS -> can tax", func(t *testing.T) {
		inv := &Invoice{
			PaymentDate: &now,
			Customer:    Customer{Document: &cpf},
			InvoiceItems: []InvoiceItem{
				{Description: "Aula", NBSCode: &nbs},
			},
		}
		if !inv.CanTax() {
			t.Errorf("expected CanTax() = true, got false")
		}
	})

	t.Run("paid PF without NBS -> cannot tax", func(t *testing.T) {
		inv := &Invoice{
			PaymentDate: &now,
			Customer:    Customer{Document: &cpf},
			InvoiceItems: []InvoiceItem{
				{Description: "Aula", NBSCode: nil},
			},
		}
		if inv.CanTax() {
			t.Errorf("expected CanTax() = false when NBS is nil, got true")
		}
	})

	t.Run("paid PJ with NBS -> cannot tax", func(t *testing.T) {
		inv := &Invoice{
			PaymentDate: &now,
			Customer:    Customer{Document: &cnpj},
			InvoiceItems: []InvoiceItem{
				{Description: "Aula", NBSCode: &nbs},
			},
		}
		if inv.CanTax() {
			t.Errorf("expected CanTax() = false for PJ, got true")
		}
	})

	t.Run("unpaid PF with NBS -> cannot tax", func(t *testing.T) {
		inv := &Invoice{
			PaymentDate: nil,
			Customer:    Customer{Document: &cpf},
			InvoiceItems: []InvoiceItem{
				{Description: "Aula", NBSCode: &nbs},
			},
		}
		if inv.CanTax() {
			t.Errorf("expected CanTax() = false for unpaid invoice, got true")
		}
	})
}
