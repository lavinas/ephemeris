package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"
)

// mapping defines the structure for API endpoint mapping
type handleService struct {
	method  string
	dto     port.InDTO
	service port.Service
	issuer  port.Issuer
	biller  port.Biller
	pixer   port.Pixer
}

// newMapping creates a new instance of mapping with the provided endpoint, method, DTO, and service.
func newHandleService(method string, dto port.InDTO, service port.Service, issuer port.Issuer, biller port.Biller, pixer port.Pixer) *handleService {
	return &handleService{
		method:  method,
		dto:     dto,
		service: service,
		issuer:  issuer,
		biller:  biller,
		pixer:   pixer,
	}
}

// APIHandler is an HTTP handler for the API
type APIHandler struct {
	logger   port.Logger
	repo     port.Repository
	issuer   port.Issuer
	biller   port.Biller
	pixer    port.Pixer
	services map[string]handleService
}

// NewAPIHandler creates a new instance of APIHandler
func NewAPIHandler(addr string, logger port.Logger, repo port.Repository, issuer port.Issuer, biller port.Biller, pixer port.Pixer) *APIHandler {
	api := &APIHandler{
		logger: logger,
		repo:   repo,
		issuer: issuer,
	}
	api.mapServices()
	return api
}

// MapEndpoint maps an API endpoint to a service method
func (h *APIHandler) mapServices() {
	h.services = map[string]handleService{
		"/ping": *newHandleService(http.MethodGet, nil, service.NewPingService(h.logger), h.issuer, h.biller, h.pixer),
		"/customer/create": *newHandleService(http.MethodPost, &dto.CustomerCreateRequest{},
			service.NewCustomerCreate(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/customer/list": *newHandleService(http.MethodGet, &dto.CustomerListRequest{},
			service.NewCustomerList(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/customer/update": *newHandleService(http.MethodPatch, &dto.CustomerUpdateRequest{},
			service.NewCustomerUpdate(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/invoice/create": *newHandleService(http.MethodPost, &dto.InvoiceCreateRequest{},
			service.NewInvoiceCreate(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/invoice/list": *newHandleService(http.MethodGet, &dto.InvoiceListRequest{},
			service.NewInvoiceList(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/invoice/update": *newHandleService(http.MethodPatch, &dto.InvoiceUpdateRequest{},
			service.NewInvoiceUpdate(h.repo, h.logger), h.issuer, h.biller, h.pixer),
		"/invoice/bill/get": *newHandleService(http.MethodGet, &dto.BillGetRequest{},
			service.NewBillGet(h.repo, h.logger, h.biller, h.pixer), h.issuer, h.biller, h.pixer),
		"/tax/send": *newHandleService(http.MethodPost, &dto.TaxSendRequest{},
			service.NewTaxSend(h.repo, h.logger, h.issuer), h.issuer, h.biller, h.pixer),
		"/tax/receive": *newHandleService(http.MethodPost, &dto.TaxReceiveRequest{},
			service.NewTaxReceive(h.repo, h.logger, h.issuer), h.issuer, h.biller, h.pixer),
	}
}

// Run starts the API server
func (h *APIHandler) Run(addr string) {
	// Set up channel to listen for interrupt signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	server := &http.Server{
		Addr:    addr,
		Handler: h,
	}
	// Start the server in a separate goroutine
	go func() {
		h.logger.IPrintf(1, "Starting API server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.logger.IPrintf(1, "API server failed: %v", err)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	// Attempt graceful shutdown with a timeout context
	h.logger.IPrintf(1, "Shutting down API server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		h.logger.IPrintf(1, "API server shutdown failed: %v", err)
	}
	h.logger.IPrintf(1, "API server stopped")
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
	// Retrieve the service for the requested path
	hservice, errOut := h.getHandleService(r)
	if errOut != nil {
		h.writeResponse(w, errOut)
		return
	}
	// Call the service method and get the response
	response := hservice.service.Run(hservice.dto)
	h.writeResponse(w, response)
	h.logger.IPrintf(1, "Handled request for %s %s", r.Method, r.URL.Path)
}

// getService retrieves the service associated with the given path, if it exists.
func (h *APIHandler) getHandleService(r *http.Request) (*handleService, *dto.ResponseBase) {
	// Check if the requested path is registered in the services map
	service, exists := h.services[r.URL.Path]
	if !exists {
		h.logger.IPrintf(1, "Service not found for path: %s", r.URL.Path)
		out := dto.NewResponseBase(404, "error", fmt.Sprintf("service not found for path: %s",
			r.URL.Path))
		return nil, &out
	}
	// Check HTTP method
	if r.Method != service.method {
		h.logger.IPrintf(1, "Method not allowed: %s for path: %s", r.Method, r.URL.Path)
		out := dto.NewResponseBase(405, "error", fmt.Sprintf("method not allowed: %s for path: %s",
			r.Method, r.URL.Path))
		return nil, &out
	}
	// Decode request body into the appropriate DTO
	if service.dto != nil {
		service.dto.Reset()
		if err := json.NewDecoder(r.Body).Decode(service.dto); err != nil {
			h.logger.IPrintf(1, "Failed to decode request body: %v", err)
			out := dto.NewResponseBase(400, "error", fmt.Sprintf("invalid json format: %v", err))
			return nil, &out
		}
	}
	return &service, nil
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
	bodyString := string(bodyBytes)
	bodyString = strings.ReplaceAll(bodyString, "\t", " ")
	bodyString = strings.ReplaceAll(bodyString, "\r", " ")
	bodyString = strings.ReplaceAll(bodyString, "\n", " ")
	bodyString = strings.ReplaceAll(bodyString, " ", "")

	h.logger.IPrintf(1, "Request received method: %s, path: %s, body: %s",
		r.Method, r.URL.Path, bodyString)
}

// writeResponse writes the given response to the http.ResponseWriter with the appropriate status
func (h *APIHandler) writeResponse(w http.ResponseWriter, response port.OutDTO) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.IPrintf(1, "Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.GetStatusCode()))
	w.Write(responseJSON)
}
