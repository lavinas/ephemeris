package driver

import (
	"encoding/json"
	"net/http"
	"fmt"

	"billing/internal/dto"
	"billing/internal/port"
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
	h.route()
	h.logger.IPrintf(1, "Starting API server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		h.logger.IPrintf(1, "API server failed: %v", err)
	}
}

// Mapping of API endpoints to service methods would be implemented here. This is a placeholder for the actual routing logic.
func (h *APIHandler) route() {
	http.HandleFunc("/create-customer", h.handleCreateCustomer)
	http.HandleFunc("/ping", h.Ping)
}

// Ping is a simple endpoint to check if the API is running
func (h *APIHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
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
	var requestData dto.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		h.logger.IPrintf(1, "Failed to decode request body: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}
	response := h.service.Run(&requestData)
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.IPrintf(1, "Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

	h.logger.IPrintf(1, "Handled create customer request: %v", requestData)

}
