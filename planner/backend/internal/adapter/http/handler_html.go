package http

import (
	"encoding/json"
	"net/http"

	"planner/internal/port"
	"planner/internal/service"
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger port.Logger
	repo     port.Repository
	template []byte
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, template []byte) *HandlerHtml {
	return &HandlerHtml{
		repo:   repo,
		logger: logger,
		template: template,
	}
}

// Ping handler for the /ping endpoint
func (h *HandlerHtml) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	service := service.NewPingService(h.logger)
	response := service.Run(nil)
	h.writeResponse(w, response)
}

// writeResponse writes the given response to the http.ResponseWriter with the appropriate status
func (h *HandlerHtml) writeResponse(w http.ResponseWriter, response port.OutDTO) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.GetStatusCode()))
	w.Write(responseJSON)
}