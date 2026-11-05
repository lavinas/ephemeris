package domain_test

import (
	"testing"
	"time"

	"accounting/internal/domain"
)

func TestKeys(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-11-03")
	k1 := domain.GenerateSantanderKey("DEBITO", "APLICACAO CONTAMAX", "04417932824", 340.00, &date, "")
	k2 := domain.GenerateSantanderKey("DEBITO", "APLICACAO CONTAMAX", "04417932824", 340.00, &date, "")
	k3 := domain.GenerateSantanderKey("CREDITO", "APLICACAO CONTAMAX", "04417932824", 340.00, &date, "")

	if k1 == "" {
		t.Errorf("key should not be empty")
	}
	if k1 != k2 {
		t.Errorf("identical fields should produce identical key")
	}
	if k1 == k3 {
		t.Errorf("different credit_debit_type should produce different key")
	}

	stKey1 := domain.GenerateStatementKey(1, 1, k1)
	stKey2 := domain.GenerateStatementKey(1, 2, k1)
	stKey3 := domain.GenerateStatementKey(1, 1, k1)

	if stKey1 == "" {
		t.Errorf("statement key should not be empty")
	}
	if stKey1 != stKey3 {
		t.Errorf("same account, occurrence, and santander key should match")
	}
	if stKey1 == stKey2 {
		t.Errorf("different occurrences should produce different statement keys")
	}
}
