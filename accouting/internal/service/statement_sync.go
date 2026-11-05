package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"accounting/internal/domain"
	"accounting/internal/dto"
	"accounting/internal/port"
)

// StatementSyncService handles synchronization of bank statements across accounts
type StatementSyncService struct {
	repo              port.Repository
	santanderProvider port.BankStatementProvider
	logger            port.Logger
	defaultDaysBack   int
}

// NewStatementSyncService creates a new StatementSyncService
func NewStatementSyncService(
	repo port.Repository,
	santanderProvider port.BankStatementProvider,
	logger port.Logger,
	defaultDaysBack int,
) *StatementSyncService {
	if defaultDaysBack <= 0 {
		defaultDaysBack = 15
	}
	return &StatementSyncService{
		repo:              repo,
		santanderProvider: santanderProvider,
		logger:            logger,
		defaultDaysBack:   defaultDaysBack,
	}
}

// Run executes the synchronization process
func (s *StatementSyncService) Run(ctx context.Context, req *dto.SyncStatementsRequest) *dto.SyncStatementsResponse {
	daysBack := s.defaultDaysBack
	if req != nil && req.DaysBack > 0 {
		daysBack = req.DaysBack
	}

	finalDate := time.Now()
	initialDate := finalDate.AddDate(0, 0, -daysBack)

	if req != nil && req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			initialDate = t
		}
	}
	if req != nil && req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			finalDate = t
		}
	}

	var accountsToProcess []domain.Account

	if req != nil && req.AccountID > 0 {
		acc, err := s.repo.GetAccountByID(req.AccountID)
		if err != nil {
			s.logger.IPrintf(2, "Error fetching account %d: %v", req.AccountID, err)
			return &dto.SyncStatementsResponse{
				ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("error fetching account: %v", err)),
			}
		}
		if acc == nil {
			return &dto.SyncStatementsResponse{
				ResponseBase: dto.NewResponseBase(404, "error", "account not found"),
			}
		}
		accountsToProcess = append(accountsToProcess, *acc)
	} else {
		activeAccounts, err := s.repo.GetAccounts(true)
		if err != nil {
			s.logger.IPrintf(2, "Error fetching active accounts: %v", err)
			return &dto.SyncStatementsResponse{
				ResponseBase: dto.NewResponseBase(500, "error", fmt.Sprintf("error fetching active accounts: %v", err)),
			}
		}
		accountsToProcess = activeAccounts
	}

	totalAccounts := 0
	totalFetched := 0
	totalImported := 0
	totalIgnored := 0

	s.logger.IPrintf(0, "Starting statement sync for %d account(s) (window: %s to %s)",
		len(accountsToProcess), initialDate.Format("2006-01-02"), finalDate.Format("2006-01-02"))

	for _, acc := range accountsToProcess {
		totalAccounts++
		// Select provider for this account
		provider, creds := s.selectProviderForAccount(&acc)
		if provider == nil {
			s.logger.IPrintf(0, "Account ID %d (Bank %s, Agency %s, Num %s) has no API credentials or provider configured. Skipping API fetch.",
				acc.ID, acc.BankCode, acc.AccountBranch, acc.AccountNumber)
			continue
		}

		s.logger.IPrintf(0, "Fetching statements for Account ID %d via provider %s", acc.ID, provider.ProviderName())
		txs, err := provider.FetchStatements(ctx, &acc, creds, initialDate, finalDate)
		if err != nil {
			s.logger.IPrintf(1, "Warning: failed to fetch statements for Account ID %d: %v", acc.ID, err)
			continue
		}

		totalFetched += len(txs)

		// Sort transactions chronologically for sequential balance calculation
		sort.SliceStable(txs, func(i, j int) bool {
			return txs[i].TransactionDate.Before(txs[j].TransactionDate)
		})

		occurrences := make(map[string]int)

		for _, tx := range txs {
			imported, err := s.processSantanderTransaction(&acc, tx, occurrences)
			if err != nil {
				s.logger.IPrintf(1, "Warning: error processing transaction (%s %s %.2f): %v",
					tx.CreditDebitType, tx.TransactionName, tx.Amount, err)
				totalIgnored++
				continue
			}
			if imported {
				totalImported++
			} else {
				totalIgnored++
			}
		}
	}

	s.logger.IPrintf(0, "Completed statement sync: %d accounts, %d fetched, %d imported, %d ignored",
		totalAccounts, totalFetched, totalImported, totalIgnored)

	return &dto.SyncStatementsResponse{
		ResponseBase:      dto.NewResponseBase(200, "success", "Sync completed"),
		ProcessedAccounts: totalAccounts,
		TotalFetched:      totalFetched,
		TotalImported:     totalImported,
		TotalIgnored:      totalIgnored,
	}
}

// selectProviderForAccount selects the appropriate bank statement provider
func (s *StatementSyncService) selectProviderForAccount(acc *domain.Account) (port.BankStatementProvider, *domain.AccountSantanderAPI) {
	if acc.AccountSantanderAPI != nil && acc.AccountSantanderAPI.ClientID != "" {
		return s.santanderProvider, acc.AccountSantanderAPI
	}

	// Try querying repository directly if not preloaded
	if creds, err := s.repo.GetAccountSantanderAPI(acc.ID); err == nil && creds != nil && creds.ClientID != "" {
		return s.santanderProvider, creds
	}

	return nil, nil
}

// processSantanderTransaction performs deduplication, balance calculation and persistence
func (s *StatementSyncService) processSantanderTransaction(acc *domain.Account, tx port.BankTransaction, occurrences map[string]int) (bool, error) {
	// 1. Ignore transactions prior to account.initial_date if defined
	if !acc.InitialDate.IsZero() && tx.TransactionDate.Before(acc.InitialDate) {
		return false, nil
	}

	// 2. Generate Santander transaction key
	txDate := tx.TransactionDate
	santanderKey := domain.GenerateSantanderKey(
		tx.CreditDebitType,
		tx.TransactionName,
		tx.HistoricComplement,
		tx.Amount,
		&txDate,
		tx.DocumentNumber,
	)

	// 3. Track occurrence for this (date, santanderKey) on the day being processed
	dateKey := fmt.Sprintf("%s_%s", txDate.Format("2006-01-02"), santanderKey)
	occurrences[dateKey]++
	occurrence := occurrences[dateKey]

	// 4. Generate Statement key and deduplicate
	statementKey := domain.GenerateStatementKey(acc.ID, occurrence, santanderKey)
	existingSt, err := s.repo.GetStatementByKey(statementKey)
	if err != nil {
		return false, err
	}
	if existingSt != nil {
		// Duplicated statement, ignore
		return false, nil
	}

	// 5. Compute signed amount: positive for CREDITO (+), negative for DEBITO (-)
	absAmount := math.Abs(tx.Amount)
	signedAmount := absAmount
	if strings.EqualFold(tx.CreditDebitType, "DEBITO") || strings.EqualFold(tx.CreditDebitType, "DÉBITO") {
		signedAmount = -absAmount
	}

	// 6. Calculate running balance:
	// Find last record on or before transaction date for this account.
	// If none found, use account.InitialBalance.
	lastSt, err := s.repo.GetLastStatementBeforeOrOnDate(acc.ID, txDate)
	if err != nil {
		return false, fmt.Errorf("error finding prior statement for account %d: %w", acc.ID, err)
	}

	prevBalance := acc.InitialBalance
	if lastSt != nil {
		prevBalance = lastSt.Balance
	}
	currentBalance := prevBalance + signedAmount

	// 7. Create entities for persistence
	statement := &domain.Statement{
		AccountID:   acc.ID,
		SDate:       &txDate,
		Transaction: tx.TransactionName,
		Complement:  tx.HistoricComplement,
		Document:    tx.DocumentNumber,
		Amount:      signedAmount,
		Balance:     currentBalance,
		Key:         statementKey,
	}

	santanderStatement := &domain.StatementSantanderAPI{
		CreditDebitType:    tx.CreditDebitType,
		TransactionName:    tx.TransactionName,
		HistoricComplement: tx.HistoricComplement,
		Amount:             tx.Amount,
		TransactionDate:    &txDate,
		DocumentNumber:     tx.DocumentNumber,
		Key:                santanderKey,
	}

	if err := s.repo.CreateSantanderStatement(statement, santanderStatement); err != nil {
		return false, fmt.Errorf("failed to persist santander statement: %w", err)
	}

	return true, nil
}

