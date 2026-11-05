package http

import (
	"bytes"
	"os"
	"testing"
	"time"

	"accounting/internal/domain"
	"accounting/internal/dto"
)

func TestStatementsTemplate_ParseAndRender(t *testing.T) {
	data, err := os.ReadFile("../../../web/templates/statements.html")
	if err != nil {
		t.Fatalf("failed to read template file: %v", err)
	}

	handler, err := NewHandlerHtml(nil, nil, nil, nil, nil, data)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	now := time.Now()
	sampleData := ActionModalViewData{
		AccountID: 1,
		SelectedAccount: &domain.Account{
			ID:            1,
			BankCode:      "033",
			AccountBranch: "0001",
			AccountNumber: "12345678",
		},
		Actions: []dto.AccountActionItemDTO{
			{
				ID:        10,
				AccountID: 1,
				Action:    domain.ActionIgnoreTransaction,
				Rules: []dto.ActionRuleDTO{
					{Field: domain.RuleFieldTransaction, Operator: domain.RuleOpContains, Value: "TARIFA", CaseSensitive: false},
					{Field: domain.RuleFieldAmount, Operator: domain.RuleOpLessThan, Value: "0", CaseSensitive: false},
				},
				CreatedAt: now,
			},
		},
		Action: &dto.AccountActionItemDTO{
			ID:        10,
			AccountID: 1,
			Action:    domain.ActionIgnoreTransaction,
			Rules: []dto.ActionRuleDTO{
				{Field: domain.RuleFieldTransaction, Operator: domain.RuleOpContains, Value: "TARIFA", CaseSensitive: false},
			},
		},
	}

	templatesToTest := []string{"modal_actions", "modal_actions_list", "modal_action_form"}
	for _, tmplName := range templatesToTest {
		var buf bytes.Buffer
		if err := handler.tmpl.ExecuteTemplate(&buf, tmplName, sampleData); err != nil {
			t.Errorf("failed to render template %s: %v", tmplName, err)
		}
	}
}

