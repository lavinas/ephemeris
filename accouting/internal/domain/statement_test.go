package domain_test

import (
	"testing"
	"time"

	"accounting/internal/domain"
)

func TestGenerateChecksum(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-11-03")
	cs1 := domain.GenerateChecksum(1, &date, 1500.50, "Pix Recebido", "0001")
	cs2 := domain.GenerateChecksum(1, &date, 1500.50, "Pix Recebido", "0001")
	cs3 := domain.GenerateChecksum(1, &date, -1500.50, "Pix Recebido", "0001")

	if cs1 == "" {
		t.Errorf("checksum should not be empty")
	}
	if cs1 != cs2 {
		t.Errorf("identical transactions should produce identical checksum")
	}
	if cs1 == cs3 {
		t.Errorf("different amounts should produce different checksum")
	}
}

