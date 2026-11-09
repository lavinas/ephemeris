package handler

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"pluggy/internal/db"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

type WebhookHandler struct {
	db     *db.DB
	syncer *service.Syncer
}

func NewWebhookHandler(database *db.DB, syncer *service.Syncer) *WebhookHandler {
	return &WebhookHandler{
		db:     database,
		syncer: syncer,
	}
}

// HandleWebhook processa eventos HTTP POST enviados pelo Pluggy
func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload pluggy.WebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("[Webhook] Payload JSON inválido recebido: %s", string(bodyBytes))
		_ = h.db.SaveWebhookEvent(r.Context(), "unknown", "", "", "", "", string(bodyBytes), "ERROR", err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("[Webhook] Evento recebido: %s (Item: %s, EventID: %s, TriggeredBy: %s)",
		payload.Event, payload.ItemID, payload.EventID, payload.TriggeredBy)

	// Salva o evento no histórico de auditoria
	err = h.db.SaveWebhookEvent(
		r.Context(),
		payload.Event,
		payload.EventID,
		payload.ItemID,
		payload.AccountID,
		payload.TriggeredBy,
		payload,
		"PROCESSED",
		"",
	)
	if err != nil {
		log.Printf("[Webhook] Erro ao gravar evento no banco: %v", err)
	}

	// Processamento assíncrono com base no tipo de evento
	if payload.ItemID != "" {
		switch payload.Event {
		case "item/created", "item/updated", "transactions/created", "transactions/updated", "investments/created", "investments/updated":
			go func(itemID string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				if _, err := h.syncer.SyncItem(ctx, itemID); err != nil {
					log.Printf("[Webhook] Erro no sync disparado por webhook para item %s: %v", itemID, err)
				}
			}(payload.ItemID)

		case "item/deleted":
			go func(itemID string) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				if err := h.db.DeleteItem(ctx, itemID); err != nil {
					log.Printf("[Webhook] Erro ao deletar item %s após webhook: %v", itemID, err)
				}
			}(payload.ItemID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "ok",
		"event":   payload.Event,
		"message": "Webhook processed successfully",
	})
}
