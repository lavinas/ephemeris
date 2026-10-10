package domain

import (
	"strings"
	"testing"
	"time"
)

func parseDate(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func parseDatePtr(s string) *time.Time {
	t := parseDate(s)
	return &t
}

func TestPeriodsOverlap(t *testing.T) {
	tests := []struct {
		name     string
		start1   time.Time
		end1     *time.Time
		start2   time.Time
		end2     *time.Time
		expected bool
	}{
		{
			name:     "Both infinite end dates - overlap",
			start1:   parseDate("2026-01-01"),
			end1:     nil,
			start2:   parseDate("2026-06-01"),
			end2:     nil,
			expected: true,
		},
		{
			name:     "First has finite end, second infinite - no overlap",
			start1:   parseDate("2026-01-01"),
			end1:     parseDatePtr("2026-05-31"),
			start2:   parseDate("2026-06-01"),
			end2:     nil,
			expected: false,
		},
		{
			name:     "First has finite end, second infinite - overlap on boundary",
			start1:   parseDate("2026-01-01"),
			end1:     parseDatePtr("2026-06-01"),
			start2:   parseDate("2026-06-01"),
			end2:     nil,
			expected: true,
		},
		{
			name:     "First infinite, second finite before first start - no overlap",
			start1:   parseDate("2026-06-01"),
			end1:     nil,
			start2:   parseDate("2026-01-01"),
			end2:     parseDatePtr("2026-05-31"),
			expected: false,
		},
		{
			name:     "First infinite, second finite overlapping - overlap",
			start1:   parseDate("2026-01-01"),
			end1:     nil,
			start2:   parseDate("2026-03-01"),
			end2:     parseDatePtr("2026-08-01"),
			expected: true,
		},
		{
			name:     "Both finite - disjoint",
			start1:   parseDate("2026-01-01"),
			end1:     parseDatePtr("2026-03-31"),
			start2:   parseDate("2026-04-01"),
			end2:     parseDatePtr("2026-06-30"),
			expected: false,
		},
		{
			name:     "Both finite - overlapping",
			start1:   parseDate("2026-01-01"),
			end1:     parseDatePtr("2026-05-31"),
			start2:   parseDate("2026-03-01"),
			end2:     parseDatePtr("2026-08-31"),
			expected: true,
		},
		{
			name:     "Both finite - one enclosed in other",
			start1:   parseDate("2026-01-01"),
			end1:     parseDatePtr("2026-12-31"),
			start2:   parseDate("2026-04-01"),
			end2:     parseDatePtr("2026-05-01"),
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PeriodsOverlap(tc.start1, tc.end1, tc.start2, tc.end2)
			if got != tc.expected {
				t.Errorf("PeriodsOverlap() = %v, want %v", got, tc.expected)
			}
			// Commutativity check
			rev := PeriodsOverlap(tc.start2, tc.end2, tc.start1, tc.end1)
			if rev != tc.expected {
				t.Errorf("PeriodsOverlap() reversed = %v, want %v", rev, tc.expected)
			}
		})
	}
}

func TestPlanTypeString(t *testing.T) {
	if PlanTypeAgenda.String() != "Agenda" {
		t.Errorf("expected Agenda, got %s", PlanTypeAgenda.String())
	}
	if PlanTypePackage.String() != "Pacote" {
		t.Errorf("expected Pacote, got %s", PlanTypePackage.String())
	}
	if PlanTypeNotebook.String() != "Caderneta" {
		t.Errorf("expected Caderneta, got %s", PlanTypeNotebook.String())
	}
	if PlanType(99).String() != "Desconhecido" {
		t.Errorf("expected Desconhecido, got %s", PlanType(99).String())
	}
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestValidatePlanPrices(t *testing.T) {
	tests := []struct {
		name          string
		planPrice     *float64
		itemPrices    []*float64
		expectErr     bool
		errMsgContain string
	}{
		{
			name:       "Valid: plan price set, all items price nil",
			planPrice:  floatPtr(150.0),
			itemPrices: []*float64{nil, nil},
			expectErr:  false,
		},
		{
			name:       "Valid: plan price nil, all items have price",
			planPrice:  nil,
			itemPrices: []*float64{floatPtr(50.0), floatPtr(60.0)},
			expectErr:  false,
		},
		{
			name:          "Error: both plan price and item price set (mutual exclusivity)",
			planPrice:     floatPtr(150.0),
			itemPrices:    []*float64{floatPtr(50.0), nil},
			expectErr:     true,
			errMsgContain: "mutuamente exclusivos",
		},
		{
			name:          "Error: both plan price and all item prices set (mutual exclusivity)",
			planPrice:     floatPtr(150.0),
			itemPrices:    []*float64{floatPtr(50.0), floatPtr(60.0)},
			expectErr:     true,
			errMsgContain: "mutuamente exclusivos",
		},
		{
			name:          "Error: neither plan price nor item prices set",
			planPrice:     nil,
			itemPrices:    []*float64{nil, nil},
			expectErr:     true,
			errMsgContain: "valor definido",
		},
		{
			name:          "Error: plan price nil and only some items have price",
			planPrice:     nil,
			itemPrices:    []*float64{floatPtr(50.0), nil},
			expectErr:     true,
			errMsgContain: "todos os itens do plano devem ter valor",
		},
		{
			name:          "Error: negative plan price",
			planPrice:     floatPtr(-10.0),
			itemPrices:    []*float64{nil},
			expectErr:     true,
			errMsgContain: "não pode ser negativo",
		},
		{
			name:          "Error: negative item price",
			planPrice:     nil,
			itemPrices:    []*float64{floatPtr(-5.0)},
			expectErr:     true,
			errMsgContain: "não pode ser negativo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePlanPrices(tc.planPrice, tc.itemPrices)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error containing '%s', got nil", tc.errMsgContain)
				}
				if tc.errMsgContain != "" && !strings.Contains(err.Error(), tc.errMsgContain) {
					t.Fatalf("expected error containing '%s', got '%v'", tc.errMsgContain, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidatePlanItemOrders(t *testing.T) {
	tests := []struct {
		name          string
		orders        []int
		expectErr     bool
		errMsgContain string
	}{
		{
			name:      "Empty orders - valid",
			orders:    []int{},
			expectErr: false,
		},
		{
			name:      "Single item - valid",
			orders:    []int{1},
			expectErr: false,
		},
		{
			name:      "Multiple items with distinct orders - valid",
			orders:    []int{1, 2, 3},
			expectErr: false,
		},
		{
			name:          "Two items with same order - error",
			orders:        []int{1, 1},
			expectErr:     true,
			errMsgContain: "mesmo valor em order",
		},
		{
			name:          "Three items with duplicate order - error",
			orders:        []int{1, 2, 2},
			expectErr:     true,
			errMsgContain: "mesmo valor em order",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePlanItemOrders(tc.orders)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error containing '%s', got nil", tc.errMsgContain)
				}
				if tc.errMsgContain != "" && !strings.Contains(err.Error(), tc.errMsgContain) {
					t.Fatalf("expected error containing '%s', got '%v'", tc.errMsgContain, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}

