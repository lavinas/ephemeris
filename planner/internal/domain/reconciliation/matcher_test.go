package reconciliation

import (
	"testing"
	"time"

	"planner/internal/domain"
)

func TestIsSessionEligible(t *testing.T) {
	date := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	// Valid session
	sess := domain.NewSession(1, 10, date, 60, string(domain.StatusDone), nil)
	if !IsSessionEligible(sess) {
		t.Errorf("expected session to be eligible")
	}

	// Cancel charge session is eligible
	sessCancelCharge := domain.NewSession(1, 10, date, 60, string(domain.StatusCancelCharge), nil)
	if !IsSessionEligible(sessCancelCharge) {
		t.Errorf("expected cancel charge session to be eligible")
	}

	// Cancel not charge is NOT eligible
	sessCancelNoCharge := domain.NewSession(1, 10, date, 60, string(domain.StatusCancelNotCharge), nil)
	if IsSessionEligible(sessCancelNoCharge) {
		t.Errorf("expected cancel not charge session to NOT be eligible")
	}

	// Alternate text for cancel not charge
	sessCancelNoCharge2 := domain.NewSession(1, 10, date, 60, "Cancelada/Não Cobrar", nil)
	if IsSessionEligible(sessCancelNoCharge2) {
		t.Errorf("expected Cancelada/Não Cobrar session to NOT be eligible")
	}

	// Soft deleted is NOT eligible
	now := time.Now()
	sess.DeletedAt = &now
	if IsSessionEligible(sess) {
		t.Errorf("expected deleted session to NOT be eligible")
	}

	// Nil session
	if IsSessionEligible(nil) {
		t.Errorf("expected nil session to NOT be eligible")
	}
}

func TestMatchesInvoiceItem(t *testing.T) {
	sessionDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	sess := domain.NewSession(1, 10, sessionDate, 60, string(domain.StatusDone), nil)

	pStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	pEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	item := domain.NewInvoiceItem(100, 10, 50.0, nil, 4, 0)

	// Matches perfectly
	if !MatchesInvoiceItem(sess, 1, item, pStart, pEnd) {
		t.Errorf("expected session to match item")
	}

	// Different customer
	if MatchesInvoiceItem(sess, 2, item, pStart, pEnd) {
		t.Errorf("expected different customer not to match")
	}

	// Different service
	itemDiffService := domain.NewInvoiceItem(100, 20, 50.0, nil, 4, 0)
	if MatchesInvoiceItem(sess, 1, itemDiffService, pStart, pEnd) {
		t.Errorf("expected different service not to match")
	}

	// Date outside period (before)
	pStartFuture := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	pEndFuture := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
	if MatchesInvoiceItem(sess, 1, item, pStartFuture, pEndFuture) {
		t.Errorf("expected session date outside period not to match")
	}

	// Date outside period (after)
	pStartPast := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	pEndPast := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	if MatchesInvoiceItem(sess, 1, item, pStartPast, pEndPast) {
		t.Errorf("expected session date outside past period not to match")
	}
}
