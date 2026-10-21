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

	htmlTemplatePath           = "/app/web/templates/sessions.html"
	localHtmlTemplatePath      = "./web/templates/sessions.html"
	plansTemplatePath          = "/app/web/templates/plans.html"
	localPlansTemplatePath     = "./web/templates/plans.html"
	invoicingTemplatePath      = "/app/web/templates/invoicing.html"
	localInvoicingTemplatePath = "./web/templates/invoicing.html"
	comptimeTemplatePath       = "/app/web/templates/comptime.html"
	localComptimeTemplatePath  = "./web/templates/comptime.html"
)

// Handler is an HTTP handler for the API
type Handler struct {
	logger        port.Logger
	repo          port.Repository
	billingClient port.BillingClient
}

// NewHandler creates a new instance of Handler
func NewHandler(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *Handler {
	var client port.BillingClient
	if len(billingClients) > 0 && billingClients[0] != nil {
		client = billingClients[0]
	}
	return &Handler{
		repo:          repo,
		logger:        logger,
		billingClient: client,
	}
}

// run runs the API server on the specified address
func (h *Handler) Run(addr string) error {
	// Start the server and handle graceful shutdown
	sessionsTemplate, err := h.getHtmlTemplate()
	if err != nil {
		return fmt.Errorf("error getting HTML sessions template: %v", err)
	}
	plansTemplate, err := h.getPlansTemplate()
	if err != nil {
		h.logger.IPrintf(1, "Warning reading plans template (%v), fallback will be used", err)
	}
	invoicingTemplate, err := h.getInvoicingTemplate()
	if err != nil {
		h.logger.IPrintf(1, "Warning reading invoicing template (%v), fallback will be used", err)
	}
	comptimeTemplate, err := h.getComptimeTemplate()
	if err != nil {
		h.logger.IPrintf(1, "Warning reading comptime template (%v), fallback will be used", err)
	}
	h.logger.IPrintf(0, "HTML templates loaded successfully")
	mainMux, err := NewRoutesWithBilling(h.repo, h.logger, h.billingClient, sessionsTemplate, plansTemplate, invoicingTemplate, comptimeTemplate)
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

// getHtmlTemplate returns the HTML sessions template
func (h *Handler) getHtmlTemplate() ([]byte, error) {
	template, err := os.ReadFile(htmlTemplatePath)
	if err == nil {
		return template, nil
	}
	return os.ReadFile(localHtmlTemplatePath)
}

// getPlansTemplate returns the HTML plans template
func (h *Handler) getPlansTemplate() ([]byte, error) {
	template, err := os.ReadFile(plansTemplatePath)
	if err == nil {
		return template, nil
	}
	return os.ReadFile(localPlansTemplatePath)
}

// getInvoicingTemplate returns the HTML invoicing template
func (h *Handler) getInvoicingTemplate() ([]byte, error) {
	template, err := os.ReadFile(invoicingTemplatePath)
	if err == nil {
		return template, nil
	}
	return os.ReadFile(localInvoicingTemplatePath)
}

// getComptimeTemplate returns the HTML comptime template
func (h *Handler) getComptimeTemplate() ([]byte, error) {
	template, err := os.ReadFile(comptimeTemplatePath)
	if err == nil {
		return template, nil
	}
	return os.ReadFile(localComptimeTemplatePath)
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
