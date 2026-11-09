package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"pluggy/internal/db"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

type APIHandler struct {
	db     *db.DB
	client *pluggy.Client
	syncer *service.Syncer
}

func NewAPIHandler(database *db.DB, client *pluggy.Client, syncer *service.Syncer) *APIHandler {
	return &APIHandler{
		db:     database,
		client: client,
		syncer: syncer,
	}
}

// CreateConnectToken gera o token para o Pluggy Connect Widget
func (a *APIHandler) CreateConnectToken(w http.ResponseWriter, r *http.Request) {
	clientUserID := r.URL.Query().Get("clientUserId")
	if clientUserID == "" {
		clientUserID = "user-ephemeris"
	}

	token, err := a.client.CreateConnectToken(r.Context(), clientUserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao gerar Connect Token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"accessToken":  token,
		"connectToken": token,
	})
}

// ConnectItem salva um item recém-conectado pelo widget e dispara o primeiro sync
func (a *APIHandler) ConnectItem(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ItemID string `json:"itemId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
		http.Error(w, "itemId é obrigatório", http.StatusBadRequest)
		return
	}

	stats, err := a.syncer.SyncItem(r.Context(), req.ItemID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao sincronizar novo item: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "connected",
		"itemId": req.ItemID,
		"stats":  stats,
	})
}

// SyncAll dispara sincronização de todas as instituições cadastradas
func (a *APIHandler) SyncAll(w http.ResponseWriter, r *http.Request) {
	stats, err := a.syncer.SyncAll(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro na sincronização: %v", err), http.StatusInternalServerError)
		return
	}

	// Se for chamada via HTMX, podemos retornar um alerta estilizado
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<div class="p-4 mb-4 text-sm text-green-800 rounded-lg bg-green-50 border border-green-200 flex items-center justify-between" role="alert">
				<div class="flex items-center">
					<svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg>
					<span><strong>Sincronização concluída com sucesso!</strong> %d contas, %d transações, %d faturas e %d investimentos atualizados.</span>
				</div>
				<button type="button" onclick="this.parentElement.remove()" class="text-green-800 hover:text-green-900">&times;</button>
			</div>
		`, stats.AccountsCount, stats.TransactionsCount, stats.CreditCardBillsCount, stats.InvestmentsCount)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// SyncItem dispara sincronização de um item específico
func (a *APIHandler) SyncItem(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("id")
	if itemID == "" {
		http.Error(w, "Parâmetro id é obrigatório", http.StatusBadRequest)
		return
	}

	stats, err := a.syncer.SyncItem(r.Context(), itemID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao sincronizar item: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<div class="p-3 my-2 text-sm text-emerald-800 rounded-lg bg-emerald-50 border border-emerald-200">
				✓ Sincronizado: %d contas e %d transações atualizadas.
			</div>
		`, stats.AccountsCount, stats.TransactionsCount)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

// DeleteItem desconecta e apaga o item
func (a *APIHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("id")
	if itemID == "" {
		http.Error(w, "Parâmetro id é obrigatório", http.StatusBadRequest)
		return
	}

	// Tenta apagar na API do Pluggy
	_ = a.client.DeleteItem(r.Context(), itemID)

	// Apaga da base local
	if err := a.db.DeleteItem(r.Context(), itemID); err != nil {
		http.Error(w, fmt.Sprintf("Erro ao excluir do banco: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "itemDeleted")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": itemID})
}

// RegisterWebhook registra uma nova URL de webhook no Pluggy
func (a *APIHandler) RegisterWebhook(w http.ResponseWriter, r *http.Request) {
	webhookURL := r.FormValue("url")
	event := r.FormValue("event")
	if webhookURL == "" {
		http.Error(w, "URL é obrigatória", http.StatusBadRequest)
		return
	}
	if event == "" {
		event = "all"
	}

	reg, err := a.client.CreateWebhook(r.Context(), webhookURL, event)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erro ao registrar webhook no Pluggy: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<div class="p-4 mb-4 text-sm text-emerald-800 bg-emerald-50 rounded-lg border border-emerald-200">
				✓ Webhook <strong>%s</strong> registrado com sucesso na Pluggy! (ID: %s)
			</div>
		`, reg.URL, reg.ID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reg)
}

// DeleteWebhook exclui um webhook cadastrado no Pluggy
func (a *APIHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	webhookID := r.URL.Query().Get("id")
	if webhookID == "" {
		http.Error(w, "id é obrigatório", http.StatusBadRequest)
		return
	}

	if err := a.client.DeleteWebhook(r.Context(), webhookID); err != nil {
		http.Error(w, fmt.Sprintf("Erro ao remover webhook: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

