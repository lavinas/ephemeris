package dto

import "accounting/internal/port"

// SyncStatementsRequest requests synchronization for an account or all accounts
type SyncStatementsRequest struct {
	AccountID      int64  `json:"account_id"`                // Optional: 0 means all active accounts
	DaysBack       int    `json:"days_back"`                 // Optional: override configured days back
	StartDate      string `json:"start_date,omitempty"`      // Optional: YYYY-MM-DD
	EndDate        string `json:"end_date,omitempty"`        // Optional: YYYY-MM-DD
	DeleteExisting bool   `json:"delete_existing,omitempty"` // Optional: if true, deletes existing statements in date range before processing
}

func (r *SyncStatementsRequest) Validate(repo port.Repository) error {
	return nil
}

// SyncStatementsResponse returns results of synchronization
type SyncStatementsResponse struct {
	ResponseBase
	ProcessedAccounts int `json:"processed_accounts"`
	TotalDeleted      int `json:"total_deleted"`
	TotalFetched      int `json:"total_fetched"`
	TotalImported     int `json:"total_imported"`
	TotalIgnored      int `json:"total_ignored"`
}

