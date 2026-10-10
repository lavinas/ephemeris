package repository

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestPlanRepository_Integration(t *testing.T) {
	// Connect to local test database if available
	repo, err := NewRepository("localhost", "root", "root", "planner", "disable", "America/Sao_Paulo", 5433, 2, "planner")
	if err != nil {
		t.Skipf("Postgres planner_db not reachable on 5433 (%v), skipping integration test", err)
		return
	}
	defer repo.Close()

	// 1. Test FindPlans
	plans, count, err := repo.FindPlans(1, 10, map[string]interface{}{"deleted_at IS NULL": nil})
	if err != nil {
		t.Fatalf("FindPlans failed: %v", err)
	}
	if count == 0 || len(plans) == 0 {
		t.Fatalf("expected plans in DB, got count=%d, len=%d", count, len(plans))
	}

	firstPlan := plans[0]
	t.Logf("Found %d plans in DB. First plan ID: %d, Customer: %+v, Type: %d, Items: %d",
		count, firstPlan.ID, firstPlan.Customer, firstPlan.PlanType, len(firstPlan.Items))

	// 2. Test GetPlanByID
	plan1, err := repo.GetPlanByID(firstPlan.ID)
	if err != nil {
		t.Fatalf("GetPlanByID failed: %v", err)
	}
	if plan1 == nil {
		t.Fatalf("expected plan with ID %d, got nil", firstPlan.ID)
	}
	if plan1.Customer == nil {
		t.Errorf("expected customer to be preloaded for plan %d", plan1.ID)
	}
	if len(plan1.Items) == 0 {
		t.Errorf("expected items to be preloaded for plan %d", plan1.ID)
	}

	// 3. Test FindCustomerActivePlansWithServices
	activePlans, err := repo.FindCustomerActivePlansWithServices(plan1.CustomerID, []int64{plan1.Items[0].ServiceID}, 0)
	if err != nil {
		t.Fatalf("FindCustomerActivePlansWithServices failed: %v", err)
	}
	if len(activePlans) == 0 {
		t.Errorf("expected at least 1 active plan for customer %d and service %d", plan1.CustomerID, plan1.Items[0].ServiceID)
	}

	// 4. Test SavePlan (create new plan in transaction and clean it up)
	newPlanStart := time.Now().AddDate(5, 0, 0) // far in future to avoid overlap
	testPlan := domain.NewPlan(plan1.CustomerID, newPlanStart, nil, 1, nil)
	testPlan.Items = []domain.PlanItem{
		{ServiceID: plan1.Items[0].ServiceID, OrderIndex: 1},
	}
	testPlan.Agenda = &domain.PlanAgenda{
		Recurrence:  1,
		WeekDay:     2,
		TermType:    1,
		PaymentDay: 15,
	}

	if err := repo.SavePlan(testPlan); err != nil {
		t.Fatalf("SavePlan failed: %v", err)
	}
	t.Logf("Successfully created test plan with ID: %d", testPlan.ID)

	// Clean up created test plan
	if err := repo.DeletePlan(testPlan.ID); err != nil {
		t.Fatalf("DeletePlan failed: %v", err)
	}
	t.Logf("Successfully deleted test plan with ID: %d", testPlan.ID)

	deletedPlan, err := repo.GetPlanByID(testPlan.ID)
	if err != nil {
		t.Fatalf("GetPlanByID after delete failed: %v", err)
	}
	if deletedPlan != nil {
		t.Errorf("expected deleted plan to not be returned, got %+v", deletedPlan)
	}
}
