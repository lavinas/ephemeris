package domain_test

import (
	"testing"
	"time"

	"accounting/internal/domain"
)

func TestActionRule_Matches(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-11-04")
	ctx := domain.TransactionEvalContext{
		Date:        date,
		Transaction: "PIX ENVIADO JOAO SILVA",
		Complement:  "04417932824",
		Document:    "DOC-12345",
		Amount:      -150.50,
	}

	tests := []struct {
		name     string
		rule     domain.BankAccountActionRule
		expected bool
	}{
		{
			name: "Transaction contains joao (case insensitive)",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldTransaction,
				Operator:      domain.RuleOpContains,
				Value:         "joao",
				CaseSensitive: false,
			},
			expected: true,
		},
		{
			name: "Transaction contains joao (case sensitive - should fail)",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldTransaction,
				Operator:      domain.RuleOpContains,
				Value:         "joao",
				CaseSensitive: true,
			},
			expected: false,
		},
		{
			name: "Transaction starts with PIX",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldTransaction,
				Operator:      domain.RuleOpStartsWith,
				Value:         "PIX",
				CaseSensitive: true,
			},
			expected: true,
		},
		{
			name: "Complement starts with 044",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldComplement,
				Operator:      domain.RuleOpStartsWith,
				Value:         "044",
				CaseSensitive: false,
			},
			expected: true,
		},
		{
			name: "Document ends with 12345",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldDocument,
				Operator:      domain.RuleOpEndsWith,
				Value:         "12345",
				CaseSensitive: false,
			},
			expected: true,
		},
		{
			name: "Transaction regex",
			rule: domain.BankAccountActionRule{
				Field:         domain.RuleFieldTransaction,
				Operator:      domain.RuleOpRegex,
				Value:         `JOAO\s+SILVA`,
				CaseSensitive: false,
			},
			expected: true,
		},
		{
			name: "Amount less than -100",
			rule: domain.BankAccountActionRule{
				Field:    domain.RuleFieldAmount,
				Operator: domain.RuleOpLessThan,
				Value:    "-100.00",
			},
			expected: true, // -150.50 < -100.00 is true
		},
		{
			name: "Amount greater than -100",
			rule: domain.BankAccountActionRule{
				Field:    domain.RuleFieldAmount,
				Operator: domain.RuleOpGreaterThan,
				Value:    "-100.00",
			},
			expected: false, // -150.50 > -100.00 is false
		},
		{
			name: "Date equals 2026-11-04",
			rule: domain.BankAccountActionRule{
				Field:    domain.RuleFieldDate,
				Operator: domain.RuleOpEquals,
				Value:    "2026-11-04",
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.rule.Matches(ctx)
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestAction_AndLogic_And_ActionEngine_OrLogic(t *testing.T) {
	date, _ := time.Parse("2006-01-02", "2026-11-04")
	ctx := domain.TransactionEvalContext{
		Date:        date,
		Transaction: "TRANSF VALOR P/ CONTA DIF TITULAR",
		Complement:  "04417932824",
		Document:    "DOC-999",
		Amount:      -70.00,
	}

	// Action 1: Requires BOTH transaction contains "TRANSF" AND complement starts with "044" -> Matches!
	action1 := domain.BankAccountAction{
		ID:        1,
		AccountID: 10,
		Action:    domain.ActionIgnoreTransaction,
		Rules: []domain.BankAccountActionRule{
			{
				Field:    domain.RuleFieldTransaction,
				Operator: domain.RuleOpContains,
				Value:    "TRANSF",
			},
			{
				Field:    domain.RuleFieldComplement,
				Operator: domain.RuleOpStartsWith,
				Value:    "044",
			},
		},
	}
	if !action1.Matches(ctx) {
		t.Errorf("expected action1 to match with AND logic")
	}

	// Action 2: Requires transaction contains "TRANSF" AND complement starts with "999" (fails second rule)
	action2 := domain.BankAccountAction{
		ID:        2,
		AccountID: 10,
		Action:    domain.ActionIgnoreTransaction,
		Rules: []domain.BankAccountActionRule{
			{
				Field:    domain.RuleFieldTransaction,
				Operator: domain.RuleOpContains,
				Value:    "TRANSF",
			},
			{
				Field:    domain.RuleFieldComplement,
				Operator: domain.RuleOpStartsWith,
				Value:    "999",
			},
		},
	}
	if action2.Matches(ctx) {
		t.Errorf("expected action2 to NOT match because one rule fails")
	}

	engine := domain.NewActionEngine()

	// When evaluating actions: action1 matches, so engine.ShouldIgnore returns true
	if !engine.ShouldIgnore([]domain.BankAccountAction{action2, action1}, ctx) {
		t.Errorf("expected engine.ShouldIgnore to be true (OR combination across actions)")
	}

	// Only action2: neither matches -> false
	if engine.ShouldIgnore([]domain.BankAccountAction{action2}, ctx) {
		t.Errorf("expected engine.ShouldIgnore to be false when no action matches")
	}
}

