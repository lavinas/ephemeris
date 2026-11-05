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

		case "/bank_account_information/v1/banks/90400888000142/accounts":
			// Test fallback when accounts endpoint returns 404
			w.WriteHeader(http.StatusNotFound)

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
				"transactions": [
					{
						"transactionId": "tx-123456",
						"transactionSign": "-",
						"transactionName": "PIX ENVIADO",
						"transactionAmount": 150.75,
						"transactionDate": "2026-11-04"
					}
				],
				"hasMore": false
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
	if txs[0].TransactionID != "tx-123456" {
		t.Errorf("expected tx-123456, got %s", txs[0].TransactionID)
	}
	if txs[0].TransactionAmount != 150.75 {
		t.Errorf("expected 150.75, got %f", txs[0].TransactionAmount)
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

func TestClient_ParseStatementResponse_BothFormats(t *testing.T) {
	c := &Client{}

	// Format 1: Real Corporate Extrato response from Santander
	corporateJSON := []byte(`{
		"_content": [
			{
				"creditDebitType": "DEBITO",
				"transactionName": "TRANSF VALOR P/ CONTA DIF TITULAR",
				"historicComplement": "04417932824",
				"amount": "70.00",
				"transactionDate": "03/11/2026"
			},
			{
				"creditDebitType": "CREDITO",
				"transactionName": "PIX RECEBIDO",
				"historicComplement": "15271970850",
				"amount": "480.00",
				"transactionDate": "03/11/2026"
			}
		],
		"_pageable": {
			"totalPages": "1",
			"totalRecords": "2"
		}
	}`)

	txs1, hasMore1, err := c.parseStatementResponse(corporateJSON)
	if err != nil {
		t.Fatalf("failed to parse corporate response: %v", err)
	}
	if len(txs1) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txs1))
	}
	if hasMore1 {
		t.Errorf("expected hasMore false")
	}
	if txs1[0].TransactionSign != "-" {
		t.Errorf("expected debit sign '-', got %s", txs1[0].TransactionSign)
	}
	if txs1[0].PartieDocument != "04417932824" {
		t.Errorf("expected counterparty doc 04417932824, got %s", txs1[0].PartieDocument)
	}
	if txs1[0].TransactionAmount != 70.00 {
		t.Errorf("expected amount 70.00, got %f", txs1[0].TransactionAmount)
	}
	if txs1[1].TransactionSign != "+" {
		t.Errorf("expected credit sign '+', got %s", txs1[1].TransactionSign)
	}
	if txs1[1].TransactionAmount != 480.00 {
		t.Errorf("expected amount 480.00, got %f", txs1[1].TransactionAmount)
	}

	// Format 2: Open Finance response from Santander Developer Portal
	openFinanceJSON := []byte(`{
		"_pageable": {
			"_moreElements": true
		},
		"_content": [
			{
				"transactionId": "1654119654b128318c0bbb905ebbb95b8c9c2f841c1a0c",
				"completedAuthorisedPaymentType": "TRANSACAO_EFETIVADA",
				"creditDebitType": "DEBITO",
				"transactionName": "Você transferiu para...",
				"type": "PIX",
				"amount": -0.1,
				"transactionCurrency": "BRL",
				"transactionDate": "2022-08-12",
				"partieCnpjCpf": "11608308863",
				"partiePersonType": "PESSOA_NATURAL",
				"partieCompeCode": "0001",
				"partieBranchCode": "0001",
				"partieNumber": "01015243",
				"partieCheckDigit": "3"
			}
		]
	}`)

	txs2, hasMore2, err := c.parseStatementResponse(openFinanceJSON)
	if err != nil {
		t.Fatalf("failed to parse open finance response: %v", err)
	}
	if len(txs2) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs2))
	}
	if !hasMore2 {
		t.Errorf("expected hasMore true")
	}
	if txs2[0].TransactionID != "1654119654b128318c0bbb905ebbb95b8c9c2f841c1a0c" {
		t.Errorf("expected transactionId, got %s", txs2[0].TransactionID)
	}
	if txs2[0].TransactionSign != "-" {
		t.Errorf("expected debit sign '-', got %s", txs2[0].TransactionSign)
	}
	if txs2[0].TransactionAmount != 0.1 {
		t.Errorf("expected abs amount 0.1, got %f", txs2[0].TransactionAmount)
	}
	if txs2[0].PartieDocument != "11608308863" {
		t.Errorf("expected partie document 11608308863, got %s", txs2[0].PartieDocument)
	}
	if txs2[0].PartieAccount != "01015243-3" {
		t.Errorf("expected partie account 01015243-3, got %s", txs2[0].PartieAccount)
	}
}

func TestClient_OpenFinanceDetailedRoute(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/oauth/v2/token":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"mock_jwt_token","token_type":"Bearer","expires_in":3600}`))

		case "/bank_account_information/v1/banks/90400888000142/accounts":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"_content": [
					{
						"accountId": "ACC-999888",
						"branchCode": "0985",
						"accountNumber": "000130010014"
					}
				]
			}`))

		case "/bank_account_information/v1/banks/90400888000142/statements":
			if r.URL.Query().Get("accountId") != "ACC-999888" {
				t.Errorf("expected accountId ACC-999888, got %s", r.URL.Query().Get("accountId"))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"_content": [
					{
						"transactionId": "tx-open-finance-1",
						"creditDebitType": "DEBITO",
						"transactionName": "PIX ENVIADO",
						"type": "PIX",
						"amount": 50.0,
						"transactionDate": "2026-11-04",
						"partieCnpjCpf": "12345678900"
					}
				],
				"_pageable": {
					"_moreElements": false
				}
			}`))

		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
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
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 tx, got %d", len(txs))
	}
	if txs[0].TransactionID != "tx-open-finance-1" {
		t.Errorf("expected tx-open-finance-1, got %s", txs[0].TransactionID)
	}
	if txs[0].PartieDocument != "12345678900" {
		t.Errorf("expected 12345678900, got %s", txs[0].PartieDocument)
	}
}
