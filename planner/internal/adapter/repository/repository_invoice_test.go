package repository

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestInvoiceRepository_Integration(t *testing.T) {
	// Connect to local test database if available
	repo, err := NewRepository("localhost", "root", "root", "planner", "disable", "America/Sao_Paulo", 5433, 2, "planner")
	if err != nil {
		t.Skipf("Postgres planner_db not reachable on 5433 (%v), skipping integration test", err)
		return
	}
	defer repo.Close()

	// 1. Test FindActivePlansForInvoicing
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	plans, err := repo.FindActivePlansForInvoicing(nil, nil, nil, nil, startDate, endDate)
	if err != nil {
		t.Fatalf("FindActivePlansForInvoicing failed: %v", err)
	}
	if len(plans) == 0 {
		t.Fatalf("expected active plans in DB, got 0")
	}
	firstPlan := plans[0]
	t.Logf("Found %d active plans for invoicing. First plan ID: %d, Customer: %+v, Type: %d, Items: %d",
		len(plans), firstPlan.ID, firstPlan.Customer, firstPlan.PlanType, len(firstPlan.Items))

	// 2. Test SaveInvoice
	now := time.Now()
	desc := "Aula Teste Fatura"
	testInvoice := &domain.Invoice{
		PlanID:        firstPlan.ID,
		InvoicingDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		Amount:        150.00,
		CreatedAt:     now,
		UpdatedAt:     now,
		Items: []domain.InvoiceItem{
			{
				ServiceID:        firstPlan.Items[0].ServiceID,
				Price:            150.00,
				Description:      &desc,
				InvoicedQuantity: 1,
				RealizedQuantity: 1,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
	}

	if err := repo.SaveInvoice(testInvoice); err != nil {
		t.Fatalf("SaveInvoice failed: %v", err)
	}
	if testInvoice.ID == 0 {
		t.Fatalf("expected testInvoice.ID > 0 after save, got 0")
	}
	t.Logf("Created test invoice with ID: %d", testInvoice.ID)

	// 3. Test GetInvoiceByID
	inv, err := repo.GetInvoiceByID(testInvoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID failed: %v", err)
	}
	if inv == nil {
		t.Fatalf("expected invoice %d, got nil", testInvoice.ID)
	}
	if len(inv.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(inv.Items))
	}
	if inv.Amount != 150.00 {
		t.Errorf("expected amount 150.00, got %.2f", inv.Amount)
	}

	// 4. Test FindInvoices
	found, err := repo.FindInvoices(map[string]interface{}{
		"plan_id = ?":        firstPlan.ID,
		"deleted_at IS NULL": nil,
	})
	if err != nil {
		t.Fatalf("FindInvoices failed: %v", err)
	}
	if len(found) == 0 {
		t.Errorf("expected at least 1 invoice for plan %d", firstPlan.ID)
	}

	// 5. Test DeleteInvoice (soft delete)
	if err := repo.DeleteInvoice(testInvoice.ID); err != nil {
		t.Fatalf("DeleteInvoice failed: %v", err)
	}

	deletedInv, err := repo.GetInvoiceByID(testInvoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after delete failed: %v", err)
	}
	if deletedInv != nil {
		t.Errorf("expected deleted invoice to not be returned, got %+v", deletedInv)
	}
	t.Logf("Successfully deleted test invoice with ID: %d", testInvoice.ID)
}
