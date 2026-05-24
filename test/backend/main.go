package main

import (
	"encoding/json"
	"net/http"
)

// Estrutura para formatar os dados de envio
type Venda struct {
	Mes   string  `json:"mes"`
	Valor float64 `json:"valor"`
}

func main() {
	// Endpoint que o Jupyter vai chamar
	http.HandleFunc("/api/vendas", func(w http.ResponseWriter, r *http.Request) {
		// Configura o cabeçalho para permitir que qualquer página acesse (CORS)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Dados fictícios
		dados := []Venda{
			{Mes: "Janeiro", Valor: 1500.0},
			{Mes: "Fevereiro", Valor: 2200.0},
			{Mes: "Março", Valor: 1800.0},
			{Mes: "Abril", Valor: 3100.0},
		}

		// Envia a resposta em formato JSON
		json.NewEncoder(w).Encode(dados)
	})

	println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}