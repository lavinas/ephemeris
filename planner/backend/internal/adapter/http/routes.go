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
	htmlRoutes, err := NewHTMLRoutes(repo, logger, htmlTemplate)
	if err != nil {
		return nil, err
	}
	mux.Handle("/html/", http.StripPrefix("/html", htmlRoutes))
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
	return mux
}

// NewHTMLRoutes creates a new instance of HTML routes with the provided logger and repository.
func NewHTMLRoutes(repo port.Repository, logger port.Logger, htmlTemplate []byte) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	handler, err := NewHandlerHtml(repo, logger, htmlTemplate)
	if err != nil {
		return nil, err
	}
	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/sessoes", handler.Sessions)
	mux.HandleFunc("/sessoes/novo", handler.SessionsCreate)
	mux.HandleFunc("/sessoes/salvar", handler.SessionsSave)
	return mux, nil
}
