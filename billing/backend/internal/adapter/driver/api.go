package driver

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"
)

// mapping defines the structure for API endpoint mapping
type handleService struct {
	method  string
	dto     port.InDTO
	service port.Service
}

// newMapping creates a new instance of mapping with the provided endpoint, method, DTO, and service.
func newHandleService(method string, dto port.InDTO, service port.Service) *handleService {
	return &handleService{
		method:  method,
		dto:     dto,
		service: service,
	}
}

// APIHandler is an HTTP handler for the API
type APIHandler struct {
	logger   port.Logger
	repo     port.Repository
	services map[string]handleService
}

// NewAPIHandler creates a new instance of APIHandler
func NewAPIHandler(addr string, logger port.Logger, repo port.Repository) *APIHandler {
	api := &APIHandler{
		logger: logger,
		repo:   repo,
	}
	api.mapServices()
	return api
}

// MapEndpoint maps an API endpoint to a service method
func (h *APIHandler) mapServices() {
	h.services = map[string]handleService{
		"/ping": *newHandleService(http.MethodGet, nil, service.NewPingService(h.logger)),
		"/customer/create": *newHandleService(http.MethodPost, &dto.CustomerCreateRequest{},
			service.NewCustomerCreate(h.repo, h.logger)),
		"/customer/list": *newHandleService(http.MethodGet, &dto.CustomerListRequest{},
			service.NewCustomerList(h.repo, h.logger)),
		"/invoice/create": *newHandleService(http.MethodPost, &dto.InvoiceCreateRequest{},
			service.NewInvoiceCreate(h.repo, h.logger)),
		"/invoice/list": *newHandleService(http.MethodGet, &dto.InvoiceListRequest{},
			service.NewInvoiceList(h.repo, h.logger)),
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

// Ping is a simple endpoint to check if the API is running
func (h *APIHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// ServeHTTP handles incoming HTTP requests and routes them to the appropriate service methods
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.IPrintf(1, "Received request: %s %s", r.Method, r.URL.Path)
	// Log the request body for debugging
	h.logRequestBody(r)
	// Check if the requested path is registered in the services map
	service, exists := h.services[r.URL.Path]
	if !exists {
		h.logger.IPrintf(1, "Service not found for path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}
	// Check HTTP method
	if r.Method != service.method {
		h.logger.IPrintf(1, "Method not allowed: %s for path: %s", r.Method, r.URL.Path)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Decode request body into the appropriate DTO
	if service.dto != nil {
		if err := json.NewDecoder(r.Body).Decode(service.dto); err != nil {
			h.logger.IPrintf(1, "Failed to decode request body: %v", err)
			http.Error(w, "Invalid json format", http.StatusBadRequest)
			return
		}
	}
	// Call the service method and get the response
	response := service.service.Run(service.dto)
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.IPrintf(1, "Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	// Write the response back to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.GetStatusCode()))
	w.Write(responseJSON)
	// Log the handled request
	h.logger.IPrintf(1, "Handled request for %s %s", r.Method, r.URL.Path)
}

// logRequestBody logs the body of the incoming HTTP request for debugging purposes.
func (h *APIHandler) logRequestBody(r *http.Request) {
	// verify if the request has a body
	if r.Body == nil || r.Body == http.NoBody {
		h.logger.IPrintf(1, "No body in request", "path", r.URL.Path)
		return
	}
	// read the original body bytes
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.IPrintf(1, "Error reading body for log", "error", err)
		return
	}
	// Return the bytes to the request so it doesn't lose the data
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	// Log the structured request (convert bytes to string)
	h.logger.IPrintf(1, "Request received method: %s, path: %s, body: %s",
		r.Method, r.URL.Path, string(bodyBytes))
}
