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
	accounts      []domain.Account
	statements    []domain.Statement
	createdSt     []*domain.Statement
	createdSantSt []*domain.StatementSantanderAPI
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

func (m *MockRepository) ListStatements(accountID int64, startDate, endDate *time.Time, txType string, offset, limit int) ([]domain.Statement, int64, error) {
	var filtered []domain.Statement
	for _, s := range m.statements {
		if s.AccountID != accountID {
			continue
		}
		if txType == "credit" && s.Amount <= 0 {
			continue
		}
		if txType == "debit" && s.Amount >= 0 {
			continue
		}
		filtered = append(filtered, s)
	}
	return filtered, int64(len(filtered)), nil
}

func (m *MockRepository) GetStatementByKey(key string) (*domain.Statement, error) {
	for _, s := range m.statements {
		if s.Key == key {
			return &s, nil
		}
	}
	for _, s := range m.createdSt {
		if s.Key == key {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MockRepository) GetSantanderKeyCountForAccountAndDate(accountID int64, date time.Time, santanderKey string) (int, error) {
	count := 0
	dateStr := date.Format("2006-01-02")
	for i, s := range m.createdSt {
		if s.AccountID == accountID && s.SDate != nil && s.SDate.Format("2006-01-02") == dateStr {
			if i < len(m.createdSantSt) && m.createdSantSt[i].Key == santanderKey {
				count++
			}
		}
	}
	return count, nil
}

func (m *MockRepository) GetLastStatementBeforeOrOnDate(accountID int64, date time.Time) (*domain.Statement, error) {
	var last *domain.Statement
	dateStr := date.Format("2006-01-02")

	for _, s := range m.createdSt {
		if s.AccountID != accountID || s.SDate == nil {
			continue
		}
		sDateStr := s.SDate.Format("2006-01-02")
		if sDateStr <= dateStr {
			if last == nil || sDateStr > last.SDate.Format("2006-01-02") || (sDateStr == last.SDate.Format("2006-01-02") && s.ID > last.ID) {
				last = s
			}
		}
	}
	return last, nil
}

func (m *MockRepository) CreateSantanderStatement(st *domain.Statement, santanderSt *domain.StatementSantanderAPI) error {
	st.ID = int64(len(m.createdSt) + 1)
	santanderSt.ID = st.ID
	m.createdSt = append(m.createdSt, st)
	m.createdSantSt = append(m.createdSantSt, santanderSt)

	// Update posterior records
	stDateStr := st.SDate.Format("2006-01-02")
	for _, other := range m.createdSt {
		if other.ID == st.ID || other.AccountID != st.AccountID || other.SDate == nil {
			continue
		}
		otherDateStr := other.SDate.Format("2006-01-02")
		if otherDateStr > stDateStr || (otherDateStr == stDateStr && other.ID > st.ID) {
			other.Balance += st.Amount
		}
	}

	return nil
}

func (m *MockRepository) CreateStatement(st *domain.Statement) error {
	st.ID = int64(len(m.createdSt) + 1)
	m.createdSt = append(m.createdSt, st)

	stDateStr := st.SDate.Format("2006-01-02")
	for _, other := range m.createdSt {
		if other.ID == st.ID || other.AccountID != st.AccountID || other.SDate == nil {
			continue
		}
		otherDateStr := other.SDate.Format("2006-01-02")
		if otherDateStr > stDateStr || (otherDateStr == stDateStr && other.ID > st.ID) {
			other.Balance += st.Amount
		}
	}

	return nil
}

func (m *MockRepository) Close() {}

// MockProvider implements port.BankStatementProvider
type MockProvider struct {
	txs []port.BankTransaction
}

func (m *MockProvider) ProviderName() string { return "santander_api" }
func (m *MockProvider) FetchStatements(ctx context.Context, account *domain.Account, creds *domain.AccountSantanderAPI, initialDate, finalDate time.Time) ([]port.BankTransaction, error) {
	return m.txs, nil
}

// MockLogger implements port.Logger
type MockLogger struct{}

func (m *MockLogger) IPrintf(level int, format string, v ...interface{}) {}
func (m *MockLogger) Close()                                            {}

func TestStatementSyncService_DeduplicationAndBalance(t *testing.T) {
	now := time.Now()
	account := domain.Account{
		ID:             1,
		BankCode:       "033",
		AccountNumber:  "13001001-4",
		AccountBranch:  "0985",
		Status:         1,
		InitialBalance: 1000.00,
		InitialDate:    now.AddDate(0, 0, -5),
		AccountSantanderAPI: &domain.AccountSantanderAPI{
			ID:           1,
			ClientID:     "mock_client_id",
			ClientSecret: "mock_secret",
		},
	}

	mockRepo := &MockRepository{
		accounts: []domain.Account{account},
	}

	mockProv := &MockProvider{
		txs: []port.BankTransaction{
			{
				CreditDebitType:    "CREDITO",
				TransactionName:    "TED Recebida",
				HistoricComplement: "DOC 1001",
				Amount:             250.00,
				TransactionDate:    now.AddDate(0, 0, -2),
				DocumentNumber:     "DOC1",
			},
			{
				CreditDebitType:    "DEBITO",
				TransactionName:    "Pagamento Boleto",
				HistoricComplement: "DOC 1002",
				Amount:             100.00,
				TransactionDate:    now.AddDate(0, 0, -1),
				DocumentNumber:     "DOC2",
			},
			{
				// Outside account.initial_date: should be ignored!
				CreditDebitType:    "CREDITO",
				TransactionName:    "Deposito Antigo",
				HistoricComplement: "DOC 0001",
				Amount:             500.00,
				TransactionDate:    now.AddDate(0, 0, -10),
				DocumentNumber:     "DOC0",
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
		t.Errorf("expected 1 ignored (due to initial_date), got %d", resp.TotalIgnored)
	}

	// Verify balance and amount signs
	if len(mockRepo.createdSt) != 2 {
		t.Fatalf("expected 2 created statements, got %d", len(mockRepo.createdSt))
	}

	st1 := mockRepo.createdSt[0]
	if st1.Amount != 250.00 {
		t.Errorf("expected amount 250.00, got %f", st1.Amount)
	}
	// InitialBalance (1000) + 250 = 1250
	if st1.Balance != 1250.00 {
		t.Errorf("expected balance 1250.00, got %f", st1.Balance)
	}

	st2 := mockRepo.createdSt[1]
	if st2.Amount != -100.00 {
		t.Errorf("expected amount -100.00, got %f", st2.Amount)
	}
	// 1250 - 100 = 1150
	if st2.Balance != 1150.00 {
		t.Errorf("expected balance 1150.00, got %f", st2.Balance)
	}

	// Test Deduplication: second run with the exact same transactions
	resp2 := syncService.Run(context.Background(), &dto.SyncStatementsRequest{})
	if resp2.TotalImported != 0 {
		t.Errorf("expected 0 imported on re-sync, got %d", resp2.TotalImported)
	}
	if resp2.TotalIgnored != 3 {
		t.Errorf("expected 3 ignored on re-sync, got %d", resp2.TotalIgnored)
	}
}
