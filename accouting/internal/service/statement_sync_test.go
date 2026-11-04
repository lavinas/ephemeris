package service_test

import (
	"context"
	"testing"
	"time"

	"accounting/internal/domain"
	"accounting/internal/dto"
	"accounting/internal/port"
	"accounting/internal/service"
)

// MockRepository implements port.Repository for tests
type MockRepository struct {
	accounts       []domain.Account
	statements     []domain.Statement
	santanderTxs   []domain.StatementSantanderAPI
	createdSt      []*domain.Statement
	createdSantSt  []*domain.StatementSantanderAPI
}

func (m *MockRepository) GetAccounts(activeOnly bool) ([]domain.Account, error) {
	return m.accounts, nil
}
func (m *MockRepository) GetAccountByID(id int64) (*domain.Account, error) {
	for _, a := range m.accounts {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) GetAccountSantanderAPI(accountID int64) (*domain.AccountSantanderAPI, error) {
	for _, a := range m.accounts {
		if a.ID == accountID && a.AccountSantanderAPI != nil {
			return a.AccountSantanderAPI, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) GetVendors() ([]domain.Vendor, error) {
	return nil, nil
}
func (m *MockRepository) ListStatements(accountID int64, startDate, endDate *time.Time, offset, limit int) ([]domain.Statement, int64, error) {
	var filtered []domain.Statement
	for _, s := range m.statements {
		if s.AccountID == accountID {
			filtered = append(filtered, s)
		}
	}
	return filtered, int64(len(filtered)), nil
}
func (m *MockRepository) GetStatementByChecksum(checksum string) (*domain.Statement, error) {
	for _, s := range m.statements {
		if s.Checksum == checksum {
			return &s, nil
		}
	}
	for _, s := range m.createdSt {
		if s.Checksum == checksum {
			return s, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) GetSantanderStatementByTransactionID(transactionID string) (*domain.StatementSantanderAPI, error) {
	for _, s := range m.santanderTxs {
		if s.TransactionID == transactionID {
			return &s, nil
		}
	}
	for _, s := range m.createdSantSt {
		if s.TransactionID == transactionID {
			return s, nil
		}
	}
	return nil, nil
}
func (m *MockRepository) CreateSantanderStatement(st *domain.Statement, santanderSt *domain.StatementSantanderAPI) error {
	st.ID = int64(len(m.createdSt) + 1)
	santanderSt.ID = st.ID
	m.createdSt = append(m.createdSt, st)
	m.createdSantSt = append(m.createdSantSt, santanderSt)
	return nil
}
func (m *MockRepository) CreateStatement(st *domain.Statement) error {
	st.ID = int64(len(m.createdSt) + 1)
	m.createdSt = append(m.createdSt, st)
	return nil
}
func (m *MockRepository) Close() {}

// MockProvider implements port.BankStatementProvider
type MockProvider struct {
	txs []port.BankTransaction
}

func (m *MockProvider) ProviderName() string { return "mock_santander" }
func (m *MockProvider) FetchStatements(ctx context.Context, account *domain.Account, creds *domain.AccountSantanderAPI, initialDate, finalDate time.Time) ([]port.BankTransaction, error) {
	return m.txs, nil
}

// MockLogger implements port.Logger
type MockLogger struct{}
func (m *MockLogger) IPrintf(level int, format string, v ...interface{}) {}
func (m *MockLogger) Close()                                            {}

func TestStatementSyncService_DeduplicationAndDePara(t *testing.T) {
	now := time.Now()
	mockRepo := &MockRepository{
		accounts: []domain.Account{
			{
				ID:            1,
				BankCode:      "033",
				AccountNumber: "13001001-4",
				AccountBranch: "0985",
				Status:        1,
				AccountSantanderAPI: &domain.AccountSantanderAPI{
					ID:           1,
					ClientID:     "mock_client_id",
					ClientSecret: "mock_secret",
				},
			},
		},
	}

	mockProv := &MockProvider{
		txs: []port.BankTransaction{
			{
				TransactionID:     "TX_001",
				TransactionName:   "TED Recebida",
				TransactionAmount: 250.00,
				TransactionSign:   "+",
				TransactionDate:   now,
				PartieDocument:    "12345678901",
			},
			{
				TransactionID:     "TX_002",
				TransactionName:   "Pagamento Boleto",
				TransactionAmount: 100.00,
				TransactionSign:   "-",
				TransactionDate:   now,
				PartieDocument:    "98765432100",
			},
			{
				// Duplicate of TX_001
				TransactionID:     "TX_001",
				TransactionName:   "TED Recebida",
				TransactionAmount: 250.00,
				TransactionSign:   "+",
				TransactionDate:   now,
				PartieDocument:    "12345678901",
			},
		},
	}

	logger := &MockLogger{}
	syncService := service.NewStatementSyncService(mockRepo, mockProv, logger, 10)

	resp := syncService.Run(context.Background(), &dto.SyncStatementsRequest{})

	if resp.TotalFetched != 3 {
		t.Errorf("expected 3 fetched, got %d", resp.TotalFetched)
	}
	if resp.TotalImported != 2 {
		t.Errorf("expected 2 imported, got %d", resp.TotalImported)
	}
	if resp.TotalIgnored != 1 {
		t.Errorf("expected 1 ignored (duplicate), got %d", resp.TotalIgnored)
	}

	// Verify de-para and sign amounts
	if len(mockRepo.createdSt) != 2 {
		t.Fatalf("expected 2 created statements, got %d", len(mockRepo.createdSt))
	}

	stCredit := mockRepo.createdSt[0]
	if stCredit.Amount != 250.00 {
		t.Errorf("expected credit amount 250.00, got %f", stCredit.Amount)
	}

	stDebit := mockRepo.createdSt[1]
	if stDebit.Amount != -100.00 {
		t.Errorf("expected debit amount -100.00, got %f", stDebit.Amount)
	}
}

