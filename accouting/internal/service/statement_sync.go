package service

import (
	"context"
	"fmt"
	"math"
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

		for _, tx := range txs {
			imported, err := s.processSantanderTransaction(&acc, tx)
			if err != nil {
				s.logger.IPrintf(1, "Warning: error processing transaction %s: %v", tx.TransactionID, err)
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

// processSantanderTransaction performs deduplication and persistence
func (s *StatementSyncService) processSantanderTransaction(acc *domain.Account, tx port.BankTransaction) (bool, error) {
	// 1. Santander specific deduplication by transaction_id
	if tx.TransactionID != "" {
		existing, err := s.repo.GetSantanderStatementByTransactionID(tx.TransactionID)
		if err != nil {
			return false, err
		}
		if existing != nil {
			// Already imported, ignore
			return false, nil
		}
	}

	// 2. Compute signed amount: positive for credit (+), negative for debit (-)
	// Na statement o valor pode ser positivo e negativo de acordo com o santander (debito/credito, valor)
	absAmount := math.Abs(tx.TransactionAmount)
	signedAmount := absAmount
	if tx.TransactionSign == "-" || strings.ToLower(tx.TransactionType) == "debito" || strings.ToLower(tx.TransactionType) == "débito" {
		signedAmount = -absAmount
	}

	// 3. Checksum deduplication for statement table
	txDate := tx.TransactionDate
	checksum := domain.GenerateChecksum(acc.ID, &txDate, signedAmount, tx.TransactionName, tx.PartieDocument)
	existingSt, err := s.repo.GetStatementByChecksum(checksum)
	if err != nil {
		return false, err
	}
	if existingSt != nil {
		// Duplicated statement, ignore
		return false, nil
	}

	// 4. Create entities for persistence
	statement := &domain.Statement{
		AccountID: acc.ID,
		SDate:     &txDate,
		Amount:    signedAmount,
		Memo:      tx.TransactionName,
		Partie:    tx.PartieDocument,
		Checksum:  checksum,
	}

	santanderStatement := &domain.StatementSantanderAPI{
		TransactionID:       tx.TransactionID,
		TransactionStatus:   tx.TransactionStatus,
		TransactionSign:     tx.TransactionSign,
		TransactionName:     tx.TransactionName,
		TransactionType:     tx.TransactionType,
		TransactionAmount:   tx.TransactionAmount,
		TransactionCurrency: tx.TransactionCurrency,
		TransactionDate:     &txDate,
		PartieDocument:      tx.PartieDocument,
		PartieType:          tx.PartieType,
		PartieBankCode:      tx.PartieBankCode,
		PartieBranchCode:    tx.PartieBranchCode,
		PartieAccount:       tx.PartieAccount,
	}

	if err := s.repo.CreateSantanderStatement(statement, santanderStatement); err != nil {
		return false, fmt.Errorf("failed to persist santander statement: %w", err)
	}

	return true, nil
}

