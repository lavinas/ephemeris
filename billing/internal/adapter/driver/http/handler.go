package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"billing/internal/port"
)

const (
	// ServerShutdownTimeout is the timeout duration for server shutdown
	ServerShutdownTimeout = 10 * time.Second

	htmlTemplatePath = "/app/web/templates/invoices.html"
)

// Handler is an HTTP handler for the API and HTML pages
type Handler struct {
	logger port.Logger
	repo   port.Repository
	taxer  port.Taxer
	pixer  port.Pixer
	issuer port.Issuer
}

// NewHandler creates a new instance of Handler
func NewHandler(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer) *Handler {
	return &Handler{
		repo:   repo,
		logger: logger,
		taxer:  taxer,
		pixer:  pixer,
		issuer: issuer,
	}
}

// Run runs the HTTP server on the specified address
func (h *Handler) Run(addr string) error {
	template, err := h.getHtmlTemplate()
	if err != nil {
		return fmt.Errorf("error getting HTML template: %v", err)
	}
	h.logger.IPrintf(0, "HTML template loaded successfully")

	mainMux, err := NewRoutes(h.repo, h.logger, h.taxer, h.pixer, h.issuer, template)
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

// getHtmlTemplate returns the HTML template content
func (h *Handler) getHtmlTemplate() ([]byte, error) {
	paths := []string{
		"web/templates/invoices.html",
		"./web/templates/invoices.html",
		htmlTemplatePath,
	}
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("error reading HTML template from searched paths (%v)", paths)
}

// exec executes the server and handles graceful shutdown
func (h *Handler) exec(addr string, mainMux *http.ServeMux) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	server := &http.Server{
		Addr:    addr,
		Handler: mainMux,
	}

	var err error
	go func() {
		if err2 := server.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("HTTP server failed: %v", err2)
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
		return fmt.Errorf("HTTP server shutdown failed: %v", err)
	}
	return nil
}
