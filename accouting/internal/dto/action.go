package dto

import (
	"errors"
	"time"

	"accounting/internal/domain"
)

type ActionRuleDTO struct {
	ID            int64  `json:"id"`
	ActionID      int64  `json:"action_id"`
	Field         int    `json:"field"`
	FieldName     string `json:"field_name,omitempty"`
	Operator      int    `json:"operator"`
	OperatorName  string `json:"operator_name,omitempty"`
	Value         string `json:"value"`
	CaseSensitive bool   `json:"case_sensitive"`
}

type AccountActionItemDTO struct {
	ID         int64           `json:"id"`
	AccountID  int64           `json:"account_id"`
	Action     int             `json:"action"`
	ActionName string          `json:"action_name"`
	Rules      []ActionRuleDTO `json:"rules"`
	CreatedAt  time.Time       `json:"created_at"`
}

type AccountActionListResponse struct {
	ResponseBase
	AccountID int64                  `json:"account_id"`
	Data      []AccountActionItemDTO `json:"data"`
}

type AccountActionDetailResponse struct {
	ResponseBase
	Data *AccountActionItemDTO `json:"data,omitempty"`
}

type AccountActionCreateRequest struct {
	AccountID int64           `json:"account_id"`
	Action    int             `json:"action"`
	Rules     []ActionRuleDTO `json:"rules"`
}

func (r *AccountActionCreateRequest) Validate() error {
	if r.AccountID <= 0 {
		return errors.New("account_id é obrigatório")
	}
	if r.Action <= 0 {
		return errors.New("action é obrigatório")
	}
	if len(r.Rules) == 0 {
		return errors.New("ao menos uma regra deve ser informada")
	}
	for _, rule := range r.Rules {
		if rule.Field < 1 || rule.Field > 5 {
			return errors.New("campo (field) inválido")
		}
		if rule.Operator < 1 || rule.Operator > 7 {
			return errors.New("operador (operator) inválido")
		}
		if rule.Value == "" {
			return errors.New("valor (value) da regra é obrigatório")
		}
	}
	return nil
}

type AccountActionUpdateRequest struct {
	ID        int64           `json:"id"`
	AccountID int64           `json:"account_id"`
	Action    int             `json:"action"`
	Rules     []ActionRuleDTO `json:"rules"`
}

func (r *AccountActionUpdateRequest) Validate() error {
	if r.ID <= 0 {
		return errors.New("id é obrigatório")
	}
	if r.AccountID <= 0 {
		return errors.New("account_id é obrigatório")
	}
	if r.Action <= 0 {
		return errors.New("action é obrigatório")
	}
	if len(r.Rules) == 0 {
		return errors.New("ao menos uma regra deve ser informada")
	}
	for _, rule := range r.Rules {
		if rule.Field < 1 || rule.Field > 5 {
			return errors.New("campo (field) inválido")
		}
		if rule.Operator < 1 || rule.Operator > 7 {
			return errors.New("operador (operator) inválido")
		}
		if rule.Value == "" {
			return errors.New("valor (value) da regra é obrigatório")
		}
	}
	return nil
}

func ToAccountActionItemDTO(a domain.BankAccountAction) AccountActionItemDTO {
	rules := make([]ActionRuleDTO, len(a.Rules))
	for i, r := range a.Rules {
		rules[i] = ActionRuleDTO{
			ID:            r.ID,
			ActionID:      r.ActionID,
			Field:         r.Field,
			FieldName:     domain.FieldName(r.Field),
			Operator:      r.Operator,
			OperatorName:  domain.OperatorName(r.Operator),
			Value:         r.Value,
			CaseSensitive: r.CaseSensitive,
		}
	}
	return AccountActionItemDTO{
		ID:         a.ID,
		AccountID:  a.AccountID,
		Action:     a.Action,
		ActionName: domain.ActionName(a.Action),
		Rules:      rules,
		CreatedAt:  a.CreatedAt,
	}
}

