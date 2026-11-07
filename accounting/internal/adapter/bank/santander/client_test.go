package santander

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"accounting/internal/adapter/config"
	"accounting/internal/domain"
)

func TestClient_GetAccessTokenAndStatements(t *testing.T) {
	tokenReceived := false
	statementsReceived := false

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/oauth/v2/token":
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("unexpected content type: %s", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("client_id") != "test_client_id" {
				t.Errorf("unexpected client_id header: %s", r.Header.Get("client_id"))
			}
			if r.Header.Get("client_secret") != "test_secret" {
				t.Errorf("unexpected client_secret header: %s", r.Header.Get("client_secret"))
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("failed to parse form: %v", err)
			}
			if r.Form.Get("grant_type") != "client_credentials" {
				t.Errorf("unexpected grant_type: %s", r.Form.Get("grant_type"))
			}
			tokenReceived = true
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"mock_jwt_token","token_type":"Bearer","expires_in":3600}`))

		case "/bank_account_information/v1/banks/90400888000142/statements/0985.000130010014":
			if r.Header.Get("Authorization") != "Bearer mock_jwt_token" {
				t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("X-Application-Key") != "test_client_id" {
				t.Errorf("unexpected X-Application-Key header: %s", r.Header.Get("X-Application-Key"))
			}
			statementsReceived = true
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"_content": [
					{
						"creditDebitType": "DEBITO",
						"transactionName": "PIX ENVIADO",
						"historicComplement": "04417932824",
						"amount": "150.75",
						"transactionDate": "04/11/2026",
						"documentNumber": "DOC123"
					}
				],
				"_pageable": {
					"_moreElements": false
				}
			}`))

		default:
			t.Errorf("unexpected request path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cfg := config.SantanderConfig{
		Environment:    "production",
		ProdBaseURL:    ts.URL,
		TokenURL:       ts.URL + "/auth/oauth/v2/token",
		BankID:         "90400888000142",
		TimeoutSeconds: 5,
	}

	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	account := &domain.Account{
		ID:            1,
		AccountBranch: "0985",
		AccountNumber: "13001001-4",
	}
	creds := &domain.AccountSantanderAPI{
		ID:           1,
		ClientID:     "test_client_id",
		ClientSecret: "test_secret",
	}

	txs, err := client.FetchStatements(context.Background(), account, creds, time.Now().AddDate(0, 0, -1), time.Now())
	if err != nil {
		t.Fatalf("unexpected error fetching statements: %v", err)
	}

	if !tokenReceived {
		t.Errorf("token endpoint was not called")
	}
	if !statementsReceived {
		t.Errorf("statements endpoint was not called")
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
	if txs[0].CreditDebitType != "DEBITO" {
		t.Errorf("expected DEBITO, got %s", txs[0].CreditDebitType)
	}
	if txs[0].TransactionName != "PIX ENVIADO" {
		t.Errorf("expected PIX ENVIADO, got %s", txs[0].TransactionName)
	}
	if txs[0].Amount != 150.75 {
		t.Errorf("expected 150.75, got %f", txs[0].Amount)
	}
	if txs[0].DocumentNumber != "DOC123" {
		t.Errorf("expected DOC123, got %s", txs[0].DocumentNumber)
	}
}

func TestClient_FormatAccountBranchConta(t *testing.T) {
	c := &Client{}
	tests := []struct {
		branch   string
		number   string
		expected string
	}{
		{
			branch:   "0985",
			number:   "13001001-4",
			expected: "0985.000130010014",
		},
		{
			branch:   "985",
			number:   "13.001001-4",
			expected: "0985.000130010014",
		},
		{
			branch:   "0000",
			number:   "000011112222",
			expected: "0000.000011112222",
		},
		{
			branch:   "1",
			number:   "2",
			expected: "0001.000000000002",
		},
	}

	for _, tc := range tests {
		actual := c.formatAccountBranchConta(tc.branch, tc.number)
		if actual != tc.expected {
			t.Errorf("formatAccountBranchConta(%q, %q) = %q; expected %q", tc.branch, tc.number, actual, tc.expected)
		}
	}
}

func TestClient_ParseStatementResponse(t *testing.T) {
	c := &Client{}

	// Real Open API Contas Correntes response from Santander
	corporateJSON := []byte(`{
		"_content": [
			{
				"creditDebitType": "DEBITO",
				"transactionName": "TRANSF VALOR P/ CONTA DIF TITULAR",
				"historicComplement": "04417932824",
				"amount": "70.00",
				"transactionDate": "03/11/2026",
				"documentNumber": "12345"
			},
			{
				"creditDebitType": "CREDITO",
				"transactionName": "PIX RECEBIDO",
				"historicComplement": "15271970850",
				"amount": "480.00",
				"transactionDate": "03/11/2026",
				"documentNumber": "67890"
			}
		],
		"_pageable": {
			"totalPages": "1",
			"totalRecords": "2"
		}
	}`)

	txs, hasMore, err := c.parseStatementResponse(corporateJSON)
	if err != nil {
		t.Fatalf("failed to parse corporate response: %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txs))
	}
	if hasMore {
		t.Errorf("expected hasMore false")
	}
	if txs[0].CreditDebitType != "DEBITO" {
		t.Errorf("expected DEBITO, got %s", txs[0].CreditDebitType)
	}
	if txs[0].HistoricComplement != "04417932824" {
		t.Errorf("expected complement 04417932824, got %s", txs[0].HistoricComplement)
	}
	if txs[0].Amount != 70.00 {
		t.Errorf("expected amount 70.00, got %f", txs[0].Amount)
	}
	if txs[0].DocumentNumber != "12345" {
		t.Errorf("expected document 12345, got %s", txs[0].DocumentNumber)
	}

	if txs[1].CreditDebitType != "CREDITO" {
		t.Errorf("expected CREDITO, got %s", txs[1].CreditDebitType)
	}
	if txs[1].Amount != 480.00 {
		t.Errorf("expected amount 480.00, got %f", txs[1].Amount)
	}
}
