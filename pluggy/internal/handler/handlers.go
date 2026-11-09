package handler

import (
	"net/http"
	"time"

	"pluggy/internal/config"
	"pluggy/internal/db"
	"pluggy/internal/logger"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

func NewRouter(cfg *config.Config, database *db.DB, client *pluggy.Client, syncer *service.Syncer, log *logger.Logger, templatesDir string) (http.Handler, error) {
	mux := http.NewServeMux()

	webHandler, err := NewWebHandler(cfg, database, client, syncer, templatesDir)
	if err != nil {
		return nil, err
	}

	apiHandler := NewAPIHandler(database, client, syncer)
	webhookHandler := NewWebhookHandler(database, syncer, log)

	// Rotas da Interface Web HTMX
	mux.HandleFunc("GET /", webHandler.RenderDashboard)
	mux.HandleFunc("GET /statements", webHandler.RenderStatements)
	mux.HandleFunc("GET /credit-cards", webHandler.RenderCreditCards)
	mux.HandleFunc("GET /investments", webHandler.RenderInvestments)
	mux.HandleFunc("GET /partials/investments/transactions", webHandler.RenderInvestmentTransactions)
	mux.HandleFunc("GET /connections", webHandler.RenderConnections)
	mux.HandleFunc("GET /webhooks", webHandler.RenderWebhooks)

	// Rotas da API e Integrações
	mux.HandleFunc("GET /api/connect-token", apiHandler.CreateConnectToken)
	mux.HandleFunc("POST /api/items/connect", apiHandler.ConnectItem)
	mux.HandleFunc("POST /api/sync/all", apiHandler.SyncAll)
	mux.HandleFunc("POST /api/sync/item", apiHandler.SyncItem)
	mux.HandleFunc("DELETE /api/items", apiHandler.DeleteItem)
	mux.HandleFunc("POST /api/webhooks/register", apiHandler.RegisterWebhook)
	mux.HandleFunc("DELETE /api/webhooks", apiHandler.DeleteWebhook)

	// Receptor de Webhook da Pluggy
	mux.HandleFunc("POST /api/webhook", webhookHandler.HandleWebhook)
	mux.HandleFunc("POST /webhook", webhookHandler.HandleWebhook)

	// Middleware estruturado de log HTTP
	loggedMux := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		mux.ServeHTTP(w, r)
		if log != nil {
			log.Infof("[HTTP] %s %s (%dms)", r.Method, r.URL.Path, time.Since(start).Milliseconds())
		}
	})

	return loggedMux, nil
}
