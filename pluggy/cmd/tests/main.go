package main

import (
	"fmt"
	"io"
	"net/http"
	"bytes"
	"encoding/json"
)

var ( 
	clientID = "1fe09f72-ca31-4139-904b-5fa9172767ce"
	clientSecret = "BSH8yKgOZmeCfpkzBYjsMQijOMQeA-KyxujQmNiZ-jU"
)

func auth() string {
	json_auth := map[string]string{
		"clientId": clientID,
		"clientSecret": clientSecret,
	}
	json_auth_bytes, _ := json.Marshal(json_auth)

	//body := bytes.NewBufferString(`{ "clientId": "f8c9b8f0-b8e2-4f0f-b8e2-4f0f8e2f0f8e2", "clientSecret": "UZzp2n7eMThpfZ74Xf7" }`)
	body := bytes.NewBuffer(json_auth_bytes)

	req, err := http.NewRequest("POST", "https://api.pluggy.ai/auth", body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)

	var json_auth_response map[string]string
	json.Unmarshal(data, &json_auth_response)
	apiKey := json_auth_response["apiKey"]
	return apiKey
}

func items(apiKey string) string {
	body := bytes.NewBufferString(`{ "connectorId": 200, "parameters": { "user": "user-ok", "password": "password-ok" }, "webhookUrl": "https://weak-squids-matter.loca.lt/api/webhook" }`)

	req, err := http.NewRequest("POST", "https://api.pluggy.ai/items", body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)

	var json_response map[string]string
	json.Unmarshal(data, &json_response)

	id := json_response["id"]
	return id
}


// Accounts
func accounts (apiKey string, itemId string) {

	url := fmt.Sprintf("https://api.pluggy.ai/accounts?itemId=%s", itemId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)
	
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Println(string(data))
}

// connectors
func connectors(apiKey string) {
	req, err := http.NewRequest("GET", "https://api.pluggy.ai/connectors", nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	data, _ := io.ReadAll(res.Body)
	fmt.Println(string(data))
}


// main 
func main() {
	apiKey := auth()
	itemId := items(apiKey)
	accounts(apiKey, itemId)
}