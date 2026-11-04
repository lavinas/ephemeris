package http

import (
	"net/http"

	"accounting/internal/port"
	"accounting/internal/service"
)

// NewRoutes configures HTTP routing for API, HTMX and static files
func NewRoutes(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
	templateData []byte,
) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	// 1. Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. API Routes
	apiHandler := NewHandlerApi(repo, logger, listService, syncService)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/ping", apiHandler.Ping)
	apiMux.HandleFunc("/healthz", apiHandler.Ping)
	apiMux.HandleFunc("/statement/list", apiHandler.StatementList)
	apiMux.HandleFunc("/statement/sync", apiHandler.StatementSync)
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// 3. HTML / HTMX Routes
	htmlHandler, err := NewHandlerHtml(repo, logger, listService, syncService, templateData)
	if err != nil {
		return nil, err
	}

	htmlMux := http.NewServeMux()
	htmlMux.HandleFunc("/statements", htmlHandler.Index)
	htmlMux.HandleFunc("/statements/sync", htmlHandler.Sync)
	mux.Handle("/html/", http.StripPrefix("/html", htmlMux))

	// Root redirects or renders index
	mux.HandleFunc("/", htmlHandler.Index)

	return mux, nil
}

