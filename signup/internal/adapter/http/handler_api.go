package http

import (
	"encoding/json"
	"net/http"
	"signup/internal/dto"
	"signup/internal/port"
	"signup/internal/service"
)

// HandlerApi is an HTTP handler for the API
type HandlerApi struct {
	logger port.Logger
	repo   port.Repository
}

// NewHandlerApi creates a new instance of HandlerApi
func NewHandlerApi(repo port.Repository, logger port.Logger) *HandlerApi {
	return &HandlerApi{
		repo:   repo,
		logger: logger,
	}
}

// Ping handler for the /ping endpoint
func (h *HandlerApi) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	service := service.NewPingService(h.logger)
	response := service.Run(nil)
	h.writeResponse(w, response)
}

// CustomerCreate is a handler for the /customer/create endpoint
func (h *HandlerApi) CustomerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewCustomerCreateRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSave(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// CustomerUpdate is a handler for the /customer/update endpoint
func (h *HandlerApi) CustomerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewCustomerUpdateRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSave(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// CustomerList is a handler for the /customer/list endpoint
func (h *HandlerApi) CustomerList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewCustomerListRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewList(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// UserCreate is a handler for the /user/create endpoint
func (h *HandlerApi) UserCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewUserCreateRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSave(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// UserList is a handler for the /user/list endpoint
func (h *HandlerApi) UserList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewUserListRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewList(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// UserUpdate is a handler for the /user/update endpoint
func (h *HandlerApi) UserUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := dto.NewUserUpdateRequest(h.repo)
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSave(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// writeResponse writes the given response to the http.ResponseWriter with the appropriate status
func (h *HandlerApi) writeResponse(w http.ResponseWriter, response port.OutDTO) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.GetStatusCode()))
	w.Write(responseJSON)
}
