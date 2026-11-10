package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var (
	clientID     = "1fe09f72-ca31-4139-904b-5fa9172767ce"
	clientSecret = "BSH8yKgOZmeCfpkzBYjsMQijOMQeA-KyxujQmNiZ-jU"

	// connectorID 2 = "Pluggy Bank" (sandbox, aceita user-ok / password-ok).
	// Não usar 200 (MeuPluggy): é OAuth e fica em WAITING_USER_INPUT sem retornar contas.
	connectorID = 2
	webhookURL  = "https://weak-squids-matter.loca.lt/api/webhook"

	pollInterval = 3 * time.Second
	pollTimeout  = 2 * time.Minute
)

// doRequest executa a chamada HTTP e retorna status e corpo, falhando em erro de rede
func doRequest(method, url, apiKey string, body io.Reader) (int, []byte) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return res.StatusCode, data
}

// fail imprime o erro com o corpo da resposta da Pluggy e encerra
func fail(step string, status int, data []byte) {
	fmt.Fprintf(os.Stderr, "[%s] falhou: HTTP %d\n%s\n", step, status, string(data))
	os.Exit(1)
}

func auth() string {
	json_auth := map[string]string{
		"clientId":     clientID,
		"clientSecret": clientSecret,
	}
	json_auth_bytes, _ := json.Marshal(json_auth)

	//body := bytes.NewBufferString(`{ "clientId": "f8c9b8f0-b8e2-4f0f-b8e2-4f0f8e2f0f8e2", "clientSecret": "UZzp2n7eMThpfZ74Xf7" }`)
	body := bytes.NewBuffer(json_auth_bytes)

	status, data := doRequest("POST", "https://api.pluggy.ai/auth", "", body)
	if status != http.StatusOK {
		fail("auth", status, data)
	}

	var json_auth_response struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.Unmarshal(data, &json_auth_response); err != nil || json_auth_response.APIKey == "" {
		fail("auth (apiKey ausente)", status, data)
	}
	fmt.Println("[auth] apiKey obtida")
	return json_auth_response.APIKey
}

func items(apiKey string) string {
	payload := map[string]any{
		"connectorId": connectorID,
		"parameters": map[string]string{
			"user":     "user-ok",
			"password": "password-ok",
		},
		"webhookUrl": webhookURL,
	}
	payloadBytes, _ := json.Marshal(payload)

	status, data := doRequest("POST", "https://api.pluggy.ai/items", apiKey, bytes.NewReader(payloadBytes))
	if status != http.StatusOK && status != http.StatusCreated {
		fail("items", status, data)
	}

	// struct tipada: a resposta contém objetos e números, não apenas strings
	var json_response struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &json_response); err != nil || json_response.ID == "" {
		fail("items (id ausente)", status, data)
	}
	fmt.Printf("[items] item criado: %s (status=%s)\n", json_response.ID, json_response.Status)
	return json_response.ID
}

// waitItem aguarda o processamento assíncrono do item até status final.
// As contas só ficam disponíveis quando o item chega em UPDATED.
func waitItem(apiKey string, itemId string) {
	url := fmt.Sprintf("https://api.pluggy.ai/items/%s", itemId)
	deadline := time.Now().Add(pollTimeout)

	for time.Now().Before(deadline) {
		status, data := doRequest("GET", url, apiKey, nil)
		if status != http.StatusOK {
			fail("waitItem", status, data)
		}

		var item struct {
			Status          string `json:"status"`
			ExecutionStatus string `json:"executionStatus"`
			Error           *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(data, &item); err != nil {
			fail("waitItem (json inválido)", status, data)
		}
		fmt.Printf("[waitItem] status=%s executionStatus=%s\n", item.Status, item.ExecutionStatus)

		switch item.Status {
		case "UPDATED":
			return
		case "LOGIN_ERROR", "OUTDATED", "WAITING_USER_INPUT":
			msg := ""
			if item.Error != nil {
				msg = fmt.Sprintf(" (%s: %s)", item.Error.Code, item.Error.Message)
			}
			fmt.Fprintf(os.Stderr, "[waitItem] item terminou em %s%s\n", item.Status, msg)
			os.Exit(1)
		}
		time.Sleep(pollInterval)
	}

	fmt.Fprintf(os.Stderr, "[waitItem] timeout de %s aguardando item %s\n", pollTimeout, itemId)
	os.Exit(1)
}

// Accounts
func accounts(apiKey string, itemId string) {

	url := fmt.Sprintf("https://api.pluggy.ai/accounts?itemId=%s", itemId)

	status, data := doRequest("GET", url, apiKey, nil)
	if status != http.StatusOK {
		fail("accounts", status, data)
	}

	var resp struct {
		Total   int `json:"total"`
		Results []struct {
			ID      string  `json:"id"`
			Type    string  `json:"type"`
			Subtype string  `json:"subtype"`
			Name    string  `json:"name"`
			Number  string  `json:"number"`
			Balance float64 `json:"balance"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		fail("accounts (json inválido)", status, data)
	}

	fmt.Printf("[accounts] %d conta(s) encontrada(s)\n", resp.Total)
	for _, acc := range resp.Results {
		fmt.Printf("  - %-6s %-16s %-20s nº %-14s saldo %.2f (id %s)\n",
			acc.Type, acc.Subtype, acc.Name, acc.Number, acc.Balance, acc.ID)
	}
}

// connectors
func connectors(apiKey string) {
	status, data := doRequest("GET", "https://api.pluggy.ai/connectors?sandbox=true", apiKey, nil)
	if status != http.StatusOK {
		fail("connectors", status, data)
	}
	fmt.Println(string(data))
}

// main
func main() {
	apiKey := auth()
	itemId := items(apiKey)
	waitItem(apiKey, itemId)
	accounts(apiKey, itemId)
}
