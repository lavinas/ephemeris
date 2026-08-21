package http

import (
	"encoding/json"
	"net/http"

	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/service"
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

// SessionCreate handler for the /session/create endpoint
func (h *HandlerApi) SessionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionCreateRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSessionCreate(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// SessionList handler for the /session/list endpoint
func (h *HandlerApi) SessionList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionListRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSessionList(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// SessionDelete handler for the /session/delete endpoint
func (h *HandlerApi) SessionDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionDeleteRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSessionDelete(h.repo, h.logger)
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
