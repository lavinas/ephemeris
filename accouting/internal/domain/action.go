package domain

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Action types
const (
	ActionIgnoreTransaction = 1 // 1 - filtrar transações (desconsiderar linha do extrato na importação)
)

// Rule Fields (bank_statement columns to analyze)
const (
	RuleFieldDate        = 1 // 1 - date (sdate)
	RuleFieldTransaction = 2 // 2 - transaction
	RuleFieldComplement  = 3 // 3 - complement
	RuleFieldDocument    = 4 // 4 - document
	RuleFieldAmount      = 5 // 5 - amount
)

// Rule Operators
const (
	RuleOpEquals      = 1 // 1 - equals
	RuleOpContains    = 2 // 2 - contains
	RuleOpStartsWith  = 3 // 3 - starts_with
	RuleOpEndsWith    = 4 // 4 - ends_with
	RuleOpRegex       = 5 // 5 - regex
	RuleOpGreaterThan = 6 // 6 - greater_than
	RuleOpLessThan    = 7 // 7 - less_than
)

// BankAccountAction represents an action attached to a bank account
type BankAccountAction struct {
	ID        int64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int64                   `gorm:"column:account_id;not null;index" json:"account_id"`
	Action    int                     `gorm:"column:action;not null" json:"action"`
	CreatedAt time.Time               `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time               `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt *time.Time              `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Relationships
	Rules []BankAccountActionRule `gorm:"foreignKey:ActionID" json:"rules,omitempty"`
}

func (BankAccountAction) TableName() string {
	return "bank_account_action"
}

// BankAccountActionRule represents an individual condition rule of an action
type BankAccountActionRule struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ActionID      int64      `gorm:"column:action_id;not null;index" json:"action_id"`
	Field         int        `gorm:"column:field;not null" json:"field"`
	Operator      int        `gorm:"column:operator;not null" json:"operator"`
	Value         string     `gorm:"column:value;size:255;not null" json:"value"`
	CaseSensitive bool       `gorm:"column:case_sensitive;default:false" json:"case_sensitive"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`
}

func (BankAccountActionRule) TableName() string {
	return "bank_account_action_rule"
}

// TransactionEvalContext holds transaction field values for rule evaluation
type TransactionEvalContext struct {
	Date        time.Time
	Transaction string
	Complement  string
	Document    string
	Amount      float64
}

// Matches checks whether this rule is satisfied by the transaction context
func (r BankAccountActionRule) Matches(ctx TransactionEvalContext) bool {
	switch r.Field {
	case RuleFieldDate:
		return r.matchDate(ctx.Date)
	case RuleFieldTransaction:
		return r.matchString(ctx.Transaction)
	case RuleFieldComplement:
		return r.matchString(ctx.Complement)
	case RuleFieldDocument:
		return r.matchString(ctx.Document)
	case RuleFieldAmount:
		return r.matchAmount(ctx.Amount)
	default:
		return false
	}
}

func (r BankAccountActionRule) matchString(target string) bool {
	s := target
	val := r.Value

	if !r.CaseSensitive {
		s = strings.ToLower(s)
		val = strings.ToLower(val)
	}

	switch r.Operator {
	case RuleOpEquals:
		return s == val
	case RuleOpContains:
		return strings.Contains(s, val)
	case RuleOpStartsWith:
		return strings.HasPrefix(s, val)
	case RuleOpEndsWith:
		return strings.HasSuffix(s, val)
	case RuleOpRegex:
		pattern := r.Value
		if !r.CaseSensitive && !strings.HasPrefix(pattern, "(?i)") {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(target)
	case RuleOpGreaterThan:
		return s > val
	case RuleOpLessThan:
		return s < val
	default:
		return false
	}
}

func (r BankAccountActionRule) matchAmount(amount float64) bool {
	cleanVal := strings.ReplaceAll(strings.TrimSpace(r.Value), ",", ".")
	valFloat, err := strconv.ParseFloat(cleanVal, 64)
	if err != nil {
		// Fallback to string comparison if not a valid float
		return r.matchString(strconv.FormatFloat(amount, 'f', 2, 64))
	}

	switch r.Operator {
	case RuleOpEquals:
		return math.Abs(amount-valFloat) < 0.001
	case RuleOpGreaterThan:
		return amount > valFloat
	case RuleOpLessThan:
		return amount < valFloat
	case RuleOpContains, RuleOpStartsWith, RuleOpEndsWith:
		amtStr := strconv.FormatFloat(amount, 'f', 2, 64)
		return r.matchString(amtStr)
	default:
		return false
	}
}

func (r BankAccountActionRule) matchDate(targetDate time.Time) bool {
	dateStr := targetDate.Format("2006-01-02")
	ruleVal := strings.TrimSpace(r.Value)

	// Try parsing rule value as YYYY-MM-DD or DD/MM/YYYY
	var ruleDate time.Time
	var parsed bool
	for _, fmtStr := range []string{"2006-01-02", "02/01/2006"} {
		if t, err := time.Parse(fmtStr, ruleVal); err == nil {
			ruleDate = t
			parsed = true
			break
		}
	}

	if parsed {
		targetNormalized := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
		ruleNormalized := time.Date(ruleDate.Year(), ruleDate.Month(), ruleDate.Day(), 0, 0, 0, 0, time.UTC)

		switch r.Operator {
		case RuleOpEquals:
			return targetNormalized.Equal(ruleNormalized)
		case RuleOpGreaterThan:
			return targetNormalized.After(ruleNormalized)
		case RuleOpLessThan:
			return targetNormalized.Before(ruleNormalized)
		}
	}

	// Fallback to string comparison (also works for contains, starts_with, ends_with)
	return r.matchString(dateStr)
}

// Matches evaluates all rules of this action. Rules within an action combine as AND.
func (a BankAccountAction) Matches(ctx TransactionEvalContext) bool {
	if len(a.Rules) == 0 {
		return false
	}
	for _, r := range a.Rules {
		if !r.Matches(ctx) {
			return false
		}
	}
	return true
}

// ActionExecutor defines the interface for an action processor
type ActionExecutor interface {
	ActionType() int
	ShouldIgnore(actions []BankAccountAction, ctx TransactionEvalContext) bool
}

// IgnoreTransactionExecutor handles Action 1 (filtrar transações)
type IgnoreTransactionExecutor struct{}

func (e *IgnoreTransactionExecutor) ActionType() int {
	return ActionIgnoreTransaction
}

// ShouldIgnore returns true if ANY action of type ActionIgnoreTransaction matches the context (OR combination across actions)
func (e *IgnoreTransactionExecutor) ShouldIgnore(actions []BankAccountAction, ctx TransactionEvalContext) bool {
	for _, a := range actions {
		if a.Action == ActionIgnoreTransaction && a.Matches(ctx) {
			return true
		}
	}
	return false
}

// ActionEngine coordinates evaluation across multiple registered actions
type ActionEngine struct {
	ignoreExecutor ActionExecutor
}

func NewActionEngine() *ActionEngine {
	return &ActionEngine{
		ignoreExecutor: &IgnoreTransactionExecutor{},
	}
}

// ShouldIgnore checks if any ActionIgnoreTransaction for the account matches the transaction
func (e *ActionEngine) ShouldIgnore(actions []BankAccountAction, ctx TransactionEvalContext) bool {
	if e == nil || e.ignoreExecutor == nil {
		return false
	}
	return e.ignoreExecutor.ShouldIgnore(actions, ctx)
}

// Helper functions for human-readable labels
func FieldName(field int) string {
	switch field {
	case RuleFieldDate:
		return "Data"
	case RuleFieldTransaction:
		return "Transação"
	case RuleFieldComplement:
		return "Complemento"
	case RuleFieldDocument:
		return "Documento"
	case RuleFieldAmount:
		return "Valor"
	default:
		return "Desconhecido"
	}
}

func OperatorName(op int) string {
	switch op {
	case RuleOpEquals:
		return "Igual a"
	case RuleOpContains:
		return "Contém"
	case RuleOpStartsWith:
		return "Inicia com"
	case RuleOpEndsWith:
		return "Termina com"
	case RuleOpRegex:
		return "Expressão Regular"
	case RuleOpGreaterThan:
		return "Maior que"
	case RuleOpLessThan:
		return "Menor que"
	default:
		return "Desconhecido"
	}
}

func ActionName(action int) string {
	switch action {
	case ActionIgnoreTransaction:
		return "Filtrar / Desconsiderar Transação"
	default:
		return "Ação Desconhecida"
	}
}

