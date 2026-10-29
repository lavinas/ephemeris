package repository

import (
	"context"
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/port"
)

func TestComptimeRepository_BasicOperations(t *testing.T) {
	// Connect to test repository using the same connection parameters
	repo, err := NewRepository("127.0.0.1", "root", "root", "planner", "disable",
		"America/Sao_Paulo", 5433, 10, "planner")
	if err != nil {
		t.Skipf("PostgreSQL database not available for integration test: %v", err)
		return
	}
	defer repo.Close()

	ctx := context.Background()

	// 1. Test FindOrCreateComptime with existing foreign keys
	vendorID := int64(1)
	customerID := int64(213)
	serviceID := int64(1)
	month := time.Date(2099, time.March, 1, 0, 0, 0, 0, time.UTC)
	nextMonth := time.Date(2099, time.April, 1, 0, 0, 0, 0, time.UTC)

	// Clean up any test artifacts first
	_ = repo.DB.Exec("DELETE FROM planner.comptime_session WHERE comptime_id IN (SELECT id FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01')", customerID, serviceID).Error
	_ = repo.DB.Exec("DELETE FROM planner.comptime_invoice_item WHERE comptime_id IN (SELECT id FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01')", customerID, serviceID).Error
	_ = repo.DB.Exec("DELETE FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01'", customerID, serviceID).Error

	defer func() {
		_ = repo.DB.Exec("DELETE FROM planner.comptime_session WHERE comptime_id IN (SELECT id FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01')", customerID, serviceID).Error
		_ = repo.DB.Exec("DELETE FROM planner.comptime_invoice_item WHERE comptime_id IN (SELECT id FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01')", customerID, serviceID).Error
		_ = repo.DB.Exec("DELETE FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month >= '2099-01-01'", customerID, serviceID).Error
	}()

	c1, err := repo.FindOrCreateComptime(ctx, vendorID, customerID, serviceID, month)
	if err != nil {
		t.Fatalf("FindOrCreateComptime failed: %v", err)
	}
	if c1.ID == 0 {
		t.Fatalf("expected created comptime to have ID > 0")
	}
	if c1.InvoicedMinutes != 0 || c1.ConsumedMinutes != 0 || c1.Balance != 0 {
		t.Errorf("expected zeroed initial values, got %+v", c1)
	}

	// Finding it again should return the same record (idempotency)
	c2, err := repo.FindOrCreateComptime(ctx, vendorID, customerID, serviceID, month)
	if err != nil {
		t.Fatalf("second FindOrCreateComptime failed: %v", err)
	}
	if c2.ID != c1.ID {
		t.Errorf("expected same ID %d, got %d", c1.ID, c2.ID)
	}

	// 2. Test SaveComptime and updates
	c1.InvoicedMinutes = 120
	c1.ConsumedMinutes = 50
	c1.Balance = 70
	if err := repo.SaveComptime(ctx, c1); err != nil {
		t.Fatalf("SaveComptime failed: %v", err)
	}

	cRetrieved, err := repo.GetComptime(ctx, c1.ID)
	if err != nil {
		t.Fatalf("GetComptime failed: %v", err)
	}
	if cRetrieved == nil || cRetrieved.Balance != 70 {
		t.Errorf("expected balance 70, got %+v", cRetrieved)
	}

	// 3. Test subsequent month initial balance propagation
	cNext, err := repo.FindOrCreateComptime(ctx, vendorID, customerID, serviceID, nextMonth)
	if err != nil {
		t.Fatalf("FindOrCreateComptime for next month failed: %v", err)
	}
	// Initial balance should be 70 (carried over from March)
	if cNext.Balance != 70 {
		t.Errorf("expected April initial balance to carry over 70, got %d", cNext.Balance)
	}

	// 4. Test GetSubsequentComptimes
	subsequents, err := repo.GetSubsequentComptimes(ctx, customerID, serviceID, month)
	if err != nil {
		t.Fatalf("GetSubsequentComptimes failed: %v", err)
	}
	if len(subsequents) != 1 || subsequents[0].ID != cNext.ID {
		t.Errorf("expected 1 subsequent comptime with ID %d, got %v", cNext.ID, subsequents)
	}

	// Clean up
	_ = repo.DB.Exec("DELETE FROM planner.comptime WHERE customer_id = ?", customerID).Error
}

func TestComptimeRepository_SessionAndListOperations(t *testing.T) {
	repo, err := NewRepository("127.0.0.1", "root", "root", "planner", "disable",
		"America/Sao_Paulo", 5433, 10, "planner")
	if err != nil {
		t.Skipf("PostgreSQL database not available: %v", err)
		return
	}
	defer repo.Close()

	ctx := context.Background()
	vendorID := int64(1)
	customerID := int64(198)
	serviceID := int64(1)
	sessionID := int64(9)
	month := time.Date(2098, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Clean up
	_ = repo.DB.Exec("DELETE FROM planner.comptime_session WHERE session_id = ?", sessionID).Error
	_ = repo.DB.Exec("DELETE FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month = '2098-01-01'", customerID, serviceID).Error
	defer func() {
		_ = repo.DB.Exec("DELETE FROM planner.comptime_session WHERE session_id = ?", sessionID).Error
		_ = repo.DB.Exec("DELETE FROM planner.comptime WHERE customer_id = ? AND service_id = ? AND month = '2098-01-01'", customerID, serviceID).Error
	}()

	comp, err := repo.FindOrCreateComptime(ctx, vendorID, customerID, serviceID, month)
	if err != nil {
		t.Fatalf("FindOrCreateComptime failed: %v", err)
	}

	cs := domain.NewComptimeSession(comp.ID, sessionID, month, 50)
	if err := repo.SaveComptimeSession(ctx, cs); err != nil {
		t.Fatalf("SaveComptimeSession failed: %v", err)
	}

	foundCS, err := repo.FindComptimeSessionBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("FindComptimeSessionBySessionID failed: %v", err)
	}
	if foundCS == nil || foundCS.ConsumedMinutes != 50 {
		t.Errorf("expected found comptime_session with 50 minutes, got %+v", foundCS)
	}

	// Test soft-delete
	if err := repo.DeleteComptimeSession(ctx, sessionID); err != nil {
		t.Fatalf("DeleteComptimeSession failed: %v", err)
	}
	afterDelete, err := repo.FindComptimeSessionBySessionID(ctx, sessionID)
	if err != nil {
		t.Fatalf("FindComptimeSessionBySessionID after delete failed: %v", err)
	}
	if afterDelete != nil {
		t.Errorf("expected nil after soft-delete, got %+v", afterDelete)
	}

	// Test ListComptime
	list, count, err := repo.ListComptime(ctx, port.ComptimeFilter{
		VendorID:   vendorID,
		CustomerID: &customerID,
		StartMonth: &month,
		EndMonth:   &month,
	})
	if err != nil {
		t.Fatalf("ListComptime failed: %v", err)
	}
	if count < 1 || len(list) < 1 {
		t.Errorf("expected at least 1 record in list, got count=%d len=%d", count, len(list))
	}
	if list[0].CustomerName == "" || list[0].ServiceName == "" {
		t.Errorf("expected joined CustomerName and ServiceName, got %+v", list[0])
	}
}
