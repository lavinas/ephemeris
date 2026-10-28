package service

import (
	"testing"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
)

func TestSessionServices_CustomerWithActivePlans_ReturnsOnlyPlanServicesOrderedByName(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	min30, min45, min60 := 30, 45, 60
	repo.services = map[int64]*domain.Service{
		1: {ID: 1, VendorID: 1, Name: "online/piano/60", SessionMinutes: &min60},
		2: {ID: 2, VendorID: 1, Name: "online/canto/30", SessionMinutes: &min30},
		3: {ID: 3, VendorID: 1, Name: "online/canto/45", SessionMinutes: &min45},
		4: {ID: 4, VendorID: 1, Name: "online/violao/60", SessionMinutes: &min60},
	}

	today := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	planStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	planEnd := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)

	// Customer 1 ("cliente_1") has two active plans with services 1 ("online/piano/60") and 2 ("online/canto/30")
	repo.plans[10] = &domain.Plan{
		ID:         10,
		CustomerID: 1,
		PlanStart:  planStart,
		PlanEnd:    &planEnd,
		PlanType:   int(domain.PlanTypeAgenda),
		Items: []domain.PlanItem{
			{ID: 100, PlanID: 10, ServiceID: 1, OrderIndex: 1, Service: repo.services[1]},
		},
	}
	repo.plans[11] = &domain.Plan{
		ID:         11,
		CustomerID: 1,
		PlanStart:  planStart,
		PlanEnd:    nil, // open-ended active plan
		PlanType:   int(domain.PlanTypePackage),
		Items: []domain.PlanItem{
			{ID: 101, PlanID: 11, ServiceID: 2, OrderIndex: 1, Service: repo.services[2]},
		},
	}

	svc := NewSessionServices(repo, logger)
	svc.SetNowFunc(func() time.Time { return today })

	out := svc.Run(&dto.SessionServicesRequest{
		Nickname: "cliente_1",
	})
	if out.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", out.GetStatusCode(), out.GetMessage())
	}

	resp := out.(*dto.SessionServicesResponse)
	if len(resp.Services) != 2 {
		t.Fatalf("expected 2 services from active plans, got %d: %+v", len(resp.Services), resp.Services)
	}
	if resp.Services[0].Name != "online/canto/30" || resp.Services[1].Name != "online/piano/60" {
		t.Errorf("expected services ordered by name [online/canto/30, online/piano/60], got [%s, %s]",
			resp.Services[0].Name, resp.Services[1].Name)
	}
}

func TestSessionServices_CustomerWithoutActivePlans_ReturnsAllServicesOrderedByName(t *testing.T) {
	repo := newMockPlanRepoForService()
	logger := &mockLogger{}

	min30, min45, min60 := 30, 45, 60
	repo.services = map[int64]*domain.Service{
		1: {ID: 1, VendorID: 1, Name: "online/piano/60", SessionMinutes: &min60},
		2: {ID: 2, VendorID: 1, Name: "online/canto/30", SessionMinutes: &min30},
		3: {ID: 3, VendorID: 1, Name: "online/canto/45", SessionMinutes: &min45},
	}

	today := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	// Expired plan for customer 1
	expiredStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	expiredEnd := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	repo.plans[20] = &domain.Plan{
		ID:         20,
		CustomerID: 1,
		PlanStart:  expiredStart,
		PlanEnd:    &expiredEnd,
		PlanType:   int(domain.PlanTypeAgenda),
		Items: []domain.PlanItem{
			{ID: 200, PlanID: 20, ServiceID: 1, OrderIndex: 1, Service: repo.services[1]},
		},
	}

	svc := NewSessionServices(repo, logger)
	svc.SetNowFunc(func() time.Time { return today })

	// 1. Customer with only expired plan -> all 3 services ordered by name
	out := svc.Run(&dto.SessionServicesRequest{
		Nickname: "cliente_1",
	})
	if out.GetStatusCode() != 200 {
		t.Fatalf("expected status 200, got %d: %s", out.GetStatusCode(), out.GetMessage())
	}
	resp := out.(*dto.SessionServicesResponse)
	if len(resp.Services) != 3 {
		t.Fatalf("expected all 3 services when no active plan exists, got %d", len(resp.Services))
	}
	expectedOrder := []string{"online/canto/30", "online/canto/45", "online/piano/60"}
	for i, exp := range expectedOrder {
		if resp.Services[i].Name != exp {
			t.Errorf("index %d: expected %q, got %q", i, exp, resp.Services[i].Name)
		}
	}

	// 2. Empty nickname -> all 3 services ordered by name
	outEmpty := svc.Run(&dto.SessionServicesRequest{Nickname: ""})
	respEmpty := outEmpty.(*dto.SessionServicesResponse)
	if len(respEmpty.Services) != 3 {
		t.Fatalf("expected all 3 services when nickname is empty, got %d", len(respEmpty.Services))
	}
}

