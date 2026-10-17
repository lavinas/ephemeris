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
	plans, err := repo.FindActivePlansForInvoicing(nil, nil, nil, nil, nil, startDate, endDate)
	if err != nil {
		t.Fatalf("FindActivePlansForInvoicing failed: %v", err)
	}
	if len(plans) == 0 {
		t.Fatalf("expected active plans in DB, got 0")
	}
	firstPlan := plans[0]
	t.Logf("Found %d active plans for invoicing. First plan ID: %d, Customer: %+v, Type: %d, Items: %d",
		len(plans), firstPlan.ID, firstPlan.Customer, firstPlan.PlanType, len(firstPlan.Items))

	// Test post-paid plan ending in Feb 2027 is found in March 2027
	testEnd := time.Date(2027, 2, 5, 0, 0, 0, 0, time.UTC)
	testPlan := &domain.Plan{
		CustomerID: firstPlan.CustomerID,
		PlanStart:  time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
		PlanEnd:    &testEnd,
		PlanType:   1,
		Agenda: &domain.PlanAgenda{
			Recurrence: int(domain.RecurrenceWeekly),
			WeekDay:    int(domain.WeekDayMonday),
			TermType:   int(domain.TermTypePostPaid),
			PaymentDay: 15,
		},
		Items: []domain.PlanItem{
			{ServiceID: firstPlan.Items[0].ServiceID, OrderIndex: 1},
		},
	}
	if err := repo.SavePlan(testPlan); err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}
	defer repo.DeletePlan(testPlan.ID)

	plansMar, err := repo.FindActivePlansForInvoicing(&testPlan.ID, nil, nil, nil, nil, time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FindActivePlansForInvoicing for March 2027 failed: %v", err)
	}
	if len(plansMar) != 1 {
		t.Errorf("expected test plan to be found for March 2027 invoicing, got %d", len(plansMar))
	}

	// Test testPlan is NOT found for April 2027
	plansApr, err := repo.FindActivePlansForInvoicing(&testPlan.ID, nil, nil, nil, nil, time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC), time.Date(2027, 4, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("FindActivePlansForInvoicing for April 2027 failed: %v", err)
	}
	if len(plansApr) != 0 {
		t.Errorf("expected test plan to NOT be found for April 2027 invoicing, got %d", len(plansApr))
	}

	// 2. Test SaveInvoice
	now := time.Now()
	desc := "Aula Teste Fatura"
	testInvoice := &domain.Invoice{
		PlanID:        firstPlan.ID,
		InvoicingDate: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		PeriodStart:   time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
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
	if !inv.PeriodStart.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected PeriodStart 2026-09-01, got %v", inv.PeriodStart)
	}
	if !inv.PeriodEnd.Equal(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("expected PeriodEnd 2026-09-30, got %v", inv.PeriodEnd)
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

	// 5. Test UpdateInvoice
	newDueDate := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	testInvoice.DueDate = newDueDate
	if len(testInvoice.Items) > 0 {
		testInvoice.Items[0].InvoicedQuantity = 7
		testInvoice.Amount = testInvoice.Items[0].Price * 7
	}
	if err := repo.UpdateInvoice(testInvoice); err != nil {
		t.Fatalf("UpdateInvoice failed: %v", err)
	}

	updatedInv, err := repo.GetInvoiceByID(testInvoice.ID)
	if err != nil {
		t.Fatalf("GetInvoiceByID after update failed: %v", err)
	}
	if updatedInv == nil {
		t.Fatalf("expected updated invoice, got nil")
	}
	if !updatedInv.DueDate.Equal(newDueDate) {
		t.Errorf("expected DueDate %v, got %v", newDueDate, updatedInv.DueDate)
	}
	if len(updatedInv.Items) > 0 && updatedInv.Items[0].InvoicedQuantity != 7 {
		t.Errorf("expected InvoicedQuantity 7, got %d", updatedInv.Items[0].InvoicedQuantity)
	}

	// 6. Test DeleteInvoicesByIDs (soft delete in batch)
	testInvoice2 := &domain.Invoice{
		PlanID:        firstPlan.ID,
		InvoicingDate: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC),
		Amount:        90.0,
		Items: []domain.InvoiceItem{
			{ServiceID: firstPlan.Items[0].ServiceID, Price: 90.0, InvoicedQuantity: 1, RealizedQuantity: 0},
		},
	}
	if err := repo.SaveInvoice(testInvoice2); err != nil {
		t.Fatalf("SaveInvoice for testInvoice2 failed: %v", err)
	}

	if err := repo.DeleteInvoicesByIDs([]int64{testInvoice.ID, testInvoice2.ID}); err != nil {
		t.Fatalf("DeleteInvoicesByIDs failed: %v", err)
	}

	for _, id := range []int64{testInvoice.ID, testInvoice2.ID} {
		deletedInv, err := repo.GetInvoiceByID(id)
		if err != nil {
			t.Fatalf("GetInvoiceByID after DeleteInvoicesByIDs failed: %v", err)
		}
		if deletedInv != nil {
			t.Errorf("expected invoice %d to be soft deleted, got %+v", id, deletedInv)
		}
	}
	t.Logf("Successfully updated and batch-deleted test invoices: %d, %d", testInvoice.ID, testInvoice2.ID)
}
