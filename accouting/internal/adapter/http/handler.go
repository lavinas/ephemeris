package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"accounting/internal/port"
	"accounting/internal/service"
)

const (
	ServerShutdownTimeout = 10 * time.Second
	containerTemplatePath = "/app/web/templates/statements.html"
	localTemplatePath     = "./web/templates/statements.html"
)

// Handler manages HTTP server lifecycle
type Handler struct {
	repo        port.Repository
	logger      port.Logger
	listService *service.StatementListService
	syncService *service.StatementSyncService
}

func NewHandler(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
) *Handler {
	return &Handler{
		repo:        repo,
		logger:      logger,
		listService: listService,
		syncService: syncService,
	}
}

func (h *Handler) Run(addr string) error {
	tmplData, err := h.getTemplate()
	if err != nil {
		h.logger.IPrintf(1, "Warning: failed to load statements template (%v). Serving with fallback.", err)
		tmplData = []byte(fallbackTemplate)
	}

	mainMux, err := NewRoutes(h.repo, h.logger, h.listService, h.syncService, tmplData)
	if err != nil {
		return fmt.Errorf("error creating routes: %w", err)
	}

	h.logger.IPrintf(0, "Starting accounting server on %s", addr)
	return h.exec(addr, mainMux)
}

func (h *Handler) getTemplate() ([]byte, error) {
	if data, err := os.ReadFile(containerTemplatePath); err == nil {
		return data, nil
	}
	return os.ReadFile(localTemplatePath)
}

func (h *Handler) exec(addr string, mux *http.ServeMux) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	var err error
	go func() {
		if err2 := server.ListenAndServe(); err2 != nil && err2 != http.ErrServerClosed {
			err = fmt.Errorf("HTTP server failed: %w", err2)
		}
	}()

	<-quit
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()
	return server.Shutdown(ctx)
}

const fallbackTemplate = `{{define "statements.html"}}<!DOCTYPE html><html><body><h1>Extrato Bancário</h1></body></html>{{end}}`

