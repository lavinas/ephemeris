package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"pluggy/internal/db"
	"pluggy/internal/logger"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

type WebhookHandler struct {
	db     *db.DB
	syncer *service.Syncer
	logger *logger.Logger
}

func NewWebhookHandler(database *db.DB, syncer *service.Syncer, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		db:     database,
		syncer: syncer,
		logger: log,
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
		if h.logger != nil {
			h.logger.Errorf("[Webhook] Payload JSON inválido recebido: %s (%v)", string(bodyBytes), err)
		}
		_ = h.db.SaveWebhookEvent(r.Context(), "unknown", "", "", "", "", string(bodyBytes), "ERROR", err.Error())
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if h.logger != nil {
		h.logger.Infof("[Webhook] Evento recebido: %s (Item: %s, EventID: %s, TriggeredBy: %s)",
			payload.Event, payload.ItemID, payload.EventID, payload.TriggeredBy)
	}

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
	if err != nil && h.logger != nil {
		h.logger.Errorf("[Webhook] Erro ao gravar evento no banco: %v", err)
	}

	// Processamento automático assíncrono desativado (aguardando nova especificação do sistema).
	// O evento é apenas registrado no banco de dados para auditoria em pluggy_webhook_event.
	if payload.ItemID != "" {
		switch payload.Event {
		case "item/created", "item/updated", "transactions/created", "transactions/updated", "investments/created", "investments/updated":
			if h.logger != nil {
				h.logger.Infof("[Webhook] Evento %s recebido para item %s (sincronismo automático desativado)", payload.Event, payload.ItemID)
			}
		case "item/deleted":
			go func(itemID string) {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
				defer cancel()
				if err := h.db.DeleteItem(ctx, itemID); err != nil && h.logger != nil {
					h.logger.Errorf("[Webhook] Erro ao deletar item local %s após webhook: %v", itemID, err)
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
