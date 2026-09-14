package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"planner/internal/port"
)

const (
	// ServerShutdownTimeout is the timeout duration for server shutdown
	ServerShutdownTimeout = 10 * time.Second

	htmlTemplatePath = "./web/templates/sessions.html"
)

// Handler is an HTTP handler for the API
type Handler struct {
	logger port.Logger
	repo   port.Repository
}

// NewHandler creates a new instance of Handler
func NewHandler(repo port.Repository, logger port.Logger) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
	}
}

// run runs the API server on the specified address
func (h *Handler) Run(addr string) error {
	// Start the server and handle graceful shutdown
	template, err := h.getHtmlTemplate()
	if err != nil {
		return fmt.Errorf("error getting HTML template: %v", err)
	}
	mainMux, err := NewRoutes(h.repo, h.logger, template)
	if err != nil {
		return fmt.Errorf("error creating routes: %v", err)
	}
	h.logger.IPrintf(0, "starting server on %s", addr)
	if err := h.exec(addr, mainMux); err != nil {
		return fmt.Errorf("error running server: %v", err)
	}
	h.logger.IPrintf(0, "stopped server, shutdown gracefully")
	return nil
}

// getHtmlTemplate returns the HTML template path
func (h *Handler) getHtmlTemplate() ([]byte, error) {
	template, err := os.ReadFile(htmlTemplatePath)
	if err != nil {
		h.logger.IPrintf(0, "error reading HTML template: %v", err)
		return nil, err
	}
	return template, nil
}

// exec executes the server and handles graceful shutdown
func (h *Handler) exec(addr string, mainMux *http.ServeMux) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	server := &http.Server{
		Addr:    addr,
		Handler: mainMux,
	}
	// Start the server in a separate goroutine
	var err error
	go func() {
		if err2 := server.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("API server failed: %v", err2)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	if err != nil {
		return err
	}
	// Attempt graceful shutdown with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("API server shutdown failed: %v", err)
	}
	return nil
}
