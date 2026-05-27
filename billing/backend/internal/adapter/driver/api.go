package driver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"billing/internal/port"
	"billing/internal/dto"
	
)


// APIHandler is an HTTP handler for the API
type APIHandler struct {
	logger  port.Logger
	service port.Service
}

// NewAPIHandler creates a new instance of APIHandler
func NewAPIHandler(service port.Service, logger port.Logger) *APIHandler {
	return &APIHandler{
		service: service,
		logger:  logger,
	}
}

// Run starts the API server
func (h *APIHandler) Run(addr string) {
	http.Handle("/", h)
	h.logger.IPrintf(1, "Starting API server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		h.logger.IPrintf(1, "API server failed: %v", err)
	}
}

// ServeHTTP handles incoming HTTP requests and routes them to the appropriate service methods
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for routing logic.
	switch r.URL.Path {
	case "/create-customer":
		h.handleCreateCustomer(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleCreateCustomer is a place for the actual implementation of the create customer endpoint
func (h *APIHandler) handleCreateCustomer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Placeholder for request parsing and service invocation
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	dto := dto.NewCreateCustomerRequest(requestData)
	response := h.service.Run(&dto)
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
	
	fmt.Fprintf(w, "Customer creation endpoint hit with data: %v", requestData)
}
