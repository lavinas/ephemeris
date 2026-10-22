package repository

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestSessionUpdate_ServiceChange(t *testing.T) {
	repo, err := NewRepository("localhost", "root", "root", "planner", "disable", "America/Sao_Paulo", 5433, 2, "planner")
	if err != nil {
		t.Skipf("Postgres planner_db not reachable on 5433 (%v), skipping integration test", err)
		return
	}
	defer repo.Close()

	// 1. Create a test session with ServiceID = 1
	dtime := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	testSession := domain.NewSession(182, 1, dtime, 30, "realizada", nil)
	if err := repo.Save(testSession); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}
	defer func() {
		now := time.Now()
		testSession.DeletedAt = &now
		testSession.Customer = nil
		testSession.Service = nil
		_ = repo.Save(testSession)
	}()

	t.Logf("Created test session with ID: %d, ServiceID: %d", testSession.ID, testSession.ServiceID)

	// 2. Fetch it using Find (which preloads Customer and Service)
	sessDomain := domain.Session{}
	found, _, err := sessDomain.Find(repo, 1, 1, testSession.ID, nil, time.Time{}, time.Time{}, 0, 0, "", "")
	if err != nil || len(found) == 0 {
		t.Fatalf("Failed to find test session: %v", err)
	}
	s := found[0]
	if s.ServiceID != 1 {
		t.Fatalf("Expected ServiceID 1, got %d", s.ServiceID)
	}
	if s.Service == nil || s.Service.ID != 1 {
		t.Fatalf("Expected preloaded Service with ID 1, got %+v", s.Service)
	}

	// 3. Now simulate what updateSessionModel does without fix vs with fix:
	// We want to change ServiceID to 3
	s.ServiceID = 3
	// Note: s.Service is still pointing to Service with ID 1
	if err := repo.Save(&s); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Re-fetch from DB:
	foundAfter, _, err := sessDomain.Find(repo, 1, 1, testSession.ID, nil, time.Time{}, time.Time{}, 0, 0, "", "")
	if err != nil || len(foundAfter) == 0 {
		t.Fatalf("Failed to re-find session: %v", err)
	}

	t.Logf("After save: ServiceID in DB is %d (s.Service was %+v)", foundAfter[0].ServiceID, foundAfter[0].Service)
	if foundAfter[0].ServiceID != 3 {
		t.Errorf("BUG CONFIRMED: Expected ServiceID to be 3, but got %d!", foundAfter[0].ServiceID)
	}
}
