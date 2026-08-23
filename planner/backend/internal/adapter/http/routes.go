package http

import (
	"net/http"
	"planner/internal/port"
)

// NewRoutes creates a new instance of Routes with the provided logger and repository.
func NewRoutes(repo port.Repository, logger port.Logger, htmlTemplate []byte) *http.ServeMux {
	mux := http.NewServeMux()
	apiRoutes := NewAPIRoutes(repo, logger)
	mux.Handle("/api/", http.StripPrefix("/api", apiRoutes))
	htmlRoutes := NewHTMLRoutes(repo, logger, htmlTemplate)
	mux.Handle("/html/", http.StripPrefix("/html", htmlRoutes))
	return mux
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
func NewHTMLRoutes(repo port.Repository, logger port.Logger, htmlTemplate []byte) *http.ServeMux {
	mux := http.NewServeMux()
	handler := NewHandlerHtml(repo, logger, htmlTemplate)
	mux.HandleFunc("/ping", handler.Ping)
	return mux
}