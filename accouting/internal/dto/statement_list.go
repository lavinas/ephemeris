package dto

import (
	"errors"
	"time"

	"accounting/internal/domain"
	"accounting/internal/port"
)

// StatementListRequest is the input for listing bank statements
type StatementListRequest struct {
	AccountID int64  `json:"account_id"`
	StartDate string `json:"start_date"` // YYYY-MM-DD
	EndDate   string `json:"end_date"`   // YYYY-MM-DD
	Type      string `json:"type"`       // all, credit, debit
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

func (r *StatementListRequest) Validate(repo port.Repository) error {
	if r.AccountID <= 0 {
		return errors.New("account_id é obrigatório")
	}
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 20
	}
	if r.Type == "" {
		r.Type = "all"
	}
	if r.StartDate != "" {
		if _, err := time.Parse("2006-01-02", r.StartDate); err != nil {
			return errors.New("start_date inválida, utilize formato YYYY-MM-DD")
		}
	}
	if r.EndDate != "" {
		if _, err := time.Parse("2006-01-02", r.EndDate); err != nil {
			return errors.New("end_date inválida, utilize formato YYYY-MM-DD")
		}
	}
	return nil
}

// StatementItemResponse represents a row returned in the statement list
type StatementItemResponse struct {
	ID          int64      `json:"id"`
	AccountID   int64      `json:"account_id"`
	SDate       *time.Time `json:"sdate"`
	Transaction string     `json:"transaction"`
	Complement  string     `json:"complement"`
	Document    string     `json:"document"`
	Amount      float64    `json:"amount"`
	Balance     float64    `json:"balance"`
	Key         string     `json:"key"`
	CreatedAt   time.Time  `json:"created_at"`
}

// StatementListResponse is the paginated response
type StatementListResponse struct {
	ResponseBase
	AccountID  int64                   `json:"account_id"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
	Data       []StatementItemResponse `json:"data"`
}

// ToStatementItemResponse converts domain.Statement to DTO item
func ToStatementItemResponse(s domain.Statement) StatementItemResponse {
	return StatementItemResponse{
		ID:          s.ID,
		AccountID:   s.AccountID,
		SDate:       s.SDate,
		Transaction: s.Transaction,
		Complement:  s.Complement,
		Document:    s.Document,
		Amount:      s.Amount,
		Balance:     s.Balance,
		Key:         s.Key,
		CreatedAt:   s.CreatedAt,
	}
}

