package http

import (
	"net/http"
	"planner/internal/port"
)

// NewRoutes creates a new instance of Routes with the provided logger and repository.
func NewRoutes(repo port.Repository, logger port.Logger, htmlTemplate []byte) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	apiRoutes := NewAPIRoutes(repo, logger)
	mux.Handle("/api/", http.StripPrefix("/api", apiRoutes))
	htmlRoutes, htmlHandler, err := NewHTMLRoutes(repo, logger, htmlTemplate)
	if err != nil {
		return nil, err
	}
	mux.Handle("/html/", http.StripPrefix("/html", htmlRoutes))
	mux.HandleFunc("/", htmlHandler.Index)
	return mux, nil
}

// NewAPIRoutes creates a new instance of API routes with the provided logger and repository.
func NewAPIRoutes(repo port.Repository, logger port.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	handler := NewHandlerApi(repo, logger)
	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/session/create", handler.SessionCreate)
	mux.HandleFunc("/session/list", handler.SessionList)
	mux.HandleFunc("/session/delete", handler.SessionDelete)
	mux.HandleFunc("/session/update", handler.SessionUpdate)
	mux.HandleFunc("/session/users", handler.SessionUsers)
	mux.HandleFunc("/ping2", handler.Ping)
	return mux
}

// NewHTMLRoutes creates a new instance of HTML routes with the provided logger and repository.
func NewHTMLRoutes(repo port.Repository, logger port.Logger, htmlTemplate []byte) (*http.ServeMux, *HandlerHtml, error) {
	mux := http.NewServeMux()
	logger.IPrintf(0, "Initializing HTML routes")
	handler, err := NewHandlerHtml(repo, logger, htmlTemplate)
	if err != nil {
		logger.IPrintf(0, "Failed to initialize HTML handler: %v", err)
		return nil, nil, err
	}
	logger.IPrintf(0, "HTML handler initialized successfully")
	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/sessoes", handler.Sessions)
	mux.HandleFunc("/sessoes/", handler.Sessions)
	mux.HandleFunc("/sessoes/novo", handler.SessionsCreate)
	mux.HandleFunc("/sessoes/salvar", handler.SessionsSave)
	mux.HandleFunc("/sessoes/tabela/reset", handler.SessionsTableReset)
	mux.HandleFunc("/sessoes/bloco/limpar", handler.SessionsBlockClear)
	mux.HandleFunc("/sessoes/deletar-aviso", handler.SessionsDeleteWarning)
	mux.HandleFunc("/sessoes/cancelar-edicao", handler.SessionsCancelEdit)
	mux.HandleFunc("/sessoes/deletar", handler.SessionsDelete)
	mux.HandleFunc("/sessoes/editar", handler.SessionsEdit)
	mux.HandleFunc("/sessoes/atualizar", handler.SessionsUpdate)
	mux.HandleFunc("/sessoes/tabela", handler.SessionsTable)
	return mux, handler, nil
}
