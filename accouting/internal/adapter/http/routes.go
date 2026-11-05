package http

import (
	"net/http"
	"strings"

	"accounting/internal/port"
	"accounting/internal/service"
)

// NewRoutes configures HTTP routing for API, HTMX and static files
func NewRoutes(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
	actionService *service.AccountActionService,
	templateData []byte,
) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	// 1. Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. API Routes
	apiHandler := NewHandlerApi(repo, logger, listService, syncService, actionService)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/ping", apiHandler.Ping)
	apiMux.HandleFunc("/healthz", apiHandler.Ping)
	apiMux.HandleFunc("/statement/list", apiHandler.StatementList)
	apiMux.HandleFunc("/statement/sync", apiHandler.StatementSync)

	// API Account Actions
	apiMux.HandleFunc("/account/action", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				apiHandler.AccountActionDetail(w, r)
			} else {
				apiHandler.AccountActionList(w, r)
			}
		case http.MethodPost:
			apiHandler.AccountActionCreate(w, r)
		case http.MethodPut:
			apiHandler.AccountActionUpdate(w, r)
		case http.MethodDelete:
			apiHandler.AccountActionDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	apiMux.HandleFunc("/account/action/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/account/action/")
		if idStr != "" {
			q := r.URL.Query()
			q.Set("id", idStr)
			r.URL.RawQuery = q.Encode()
		}
		switch r.Method {
		case http.MethodGet:
			apiHandler.AccountActionDetail(w, r)
		case http.MethodPut, http.MethodPost:
			apiHandler.AccountActionUpdate(w, r)
		case http.MethodDelete:
			apiHandler.AccountActionDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// 3. HTML / HTMX Routes
	htmlHandler, err := NewHandlerHtml(repo, logger, listService, syncService, actionService, templateData)
	if err != nil {
		return nil, err
	}

	htmlMux := http.NewServeMux()
	htmlMux.HandleFunc("/statements", htmlHandler.Index)
	htmlMux.HandleFunc("/statements/", htmlHandler.Index)
	htmlMux.HandleFunc("/statements/tabela", htmlHandler.Table)
	htmlMux.HandleFunc("/statements/tabela/reset", htmlHandler.TableReset)
	htmlMux.HandleFunc("/extratos", htmlHandler.Index)
	htmlMux.HandleFunc("/extratos/", htmlHandler.Index)
	htmlMux.HandleFunc("/extratos/tabela", htmlHandler.Table)
	htmlMux.HandleFunc("/extratos/tabela/reset", htmlHandler.TableReset)
	htmlMux.HandleFunc("/statements/sync", htmlHandler.Sync)
	htmlMux.HandleFunc("/extratos/sync", htmlHandler.Sync)

	// Action modal routes
	htmlMux.HandleFunc("/statements/actions", htmlHandler.ActionsModal)
	htmlMux.HandleFunc("/statements/actions/list", htmlHandler.ActionsList)
	htmlMux.HandleFunc("/statements/actions/new", htmlHandler.ActionsNew)
	htmlMux.HandleFunc("/statements/actions/edit", htmlHandler.ActionsEdit)
	htmlMux.HandleFunc("/statements/actions/save", htmlHandler.ActionsSave)
	htmlMux.HandleFunc("/statements/actions/delete", htmlHandler.ActionsDelete)

	htmlMux.HandleFunc("/extratos/actions", htmlHandler.ActionsModal)
	htmlMux.HandleFunc("/extratos/actions/list", htmlHandler.ActionsList)
	htmlMux.HandleFunc("/extratos/actions/new", htmlHandler.ActionsNew)
	htmlMux.HandleFunc("/extratos/actions/edit", htmlHandler.ActionsEdit)
	htmlMux.HandleFunc("/extratos/actions/save", htmlHandler.ActionsSave)
	htmlMux.HandleFunc("/extratos/actions/delete", htmlHandler.ActionsDelete)

	// Registra tanto com /html/ quanto sem prefixo para compatibilidade total com proxies
	mux.Handle("/html/", http.StripPrefix("/html", htmlMux))
	mux.Handle("/statements", htmlMux)
	mux.Handle("/statements/", htmlMux)
	mux.Handle("/extratos", htmlMux)
	mux.Handle("/extratos/", htmlMux)

	// Root redirects or renders index
	mux.HandleFunc("/", htmlHandler.Index)

	return mux, nil
}

