package http

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	if r.Method != http.MethodDelete {
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

// SessionUpdate handler for the /session/update endpoint
func (h *HandlerApi) SessionUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSessionUpdate(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// SessionUsers handler for the /session/users endpoint
func (h *HandlerApi) SessionUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionUsersRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	service := service.NewSessionUsers(h.repo, h.logger)
	response := service.Run(requestDTO)
	h.writeResponse(w, response)
}

// PlanCreate handler for the /plan/create endpoint
func (h *HandlerApi) PlanCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.PlanCreateRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	svc := service.NewPlanCreate(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// PlanList handler for the /plan/list endpoint
func (h *HandlerApi) PlanList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.PlanListRequest{}
	if r.ContentLength > 0 {
		err := json.NewDecoder(r.Body).Decode(requestDTO)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	} else {
		requestDTO.Page = 1
		requestDTO.PageSize = 10
	}
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		if a, err := strconv.ParseBool(activeStr); err == nil {
			requestDTO.Active = &a
		}
	}
	svc := service.NewPlanList(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// PlanDelete handler for the /plan/delete endpoint
func (h *HandlerApi) PlanDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.PlanDeleteRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	svc := service.NewPlanDelete(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// PlanUpdate handler for the /plan/update endpoint
func (h *HandlerApi) PlanUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.PlanUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(requestDTO)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	svc := service.NewPlanUpdate(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// Invoicing handler for the /invoicing or /plan/invoicing endpoint
func (h *HandlerApi) Invoicing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoicingRequest{}
	if r.ContentLength > 0 {
		err := json.NewDecoder(r.Body).Decode(requestDTO)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	q := r.URL.Query()
	if requestDTO.PlanID == nil {
		if pStr := q.Get("plan_id"); pStr != "" {
			if id, err := strconv.ParseInt(pStr, 10, 64); err == nil && id > 0 {
				requestDTO.PlanID = &id
			}
		} else if pStr := q.Get("plano_id"); pStr != "" {
			if id, err := strconv.ParseInt(pStr, 10, 64); err == nil && id > 0 {
				requestDTO.PlanID = &id
			}
		}
	}
	if requestDTO.StartDate == nil && q.Get("start_date") != "" {
		s := q.Get("start_date")
		requestDTO.StartDate = &s
	}
	if requestDTO.EndDate == nil && q.Get("end_date") != "" {
		e := q.Get("end_date")
		requestDTO.EndDate = &e
	}
	svc := service.NewInvoicing(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// InvoiceDelete handler for deleting (soft-delete) an invoice (/invoice/delete or /invoicing/delete)
func (h *HandlerApi) InvoiceDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoiceDeleteRequest{}
	if r.ContentLength > 0 {
		err := json.NewDecoder(r.Body).Decode(requestDTO)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	if requestDTO.InvoiceID <= 0 {
		if idStr := r.URL.Query().Get("id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				requestDTO.InvoiceID = id
			}
		} else if idStr := r.URL.Query().Get("invoice_id"); idStr != "" {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil && id > 0 {
				requestDTO.InvoiceID = id
			}
		}
	}
	svc := service.NewInvoiceDelete(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// InvoiceList handler for querying invoices (/plan/invoicing/invoices and /plan/invoicing/list)
func (h *HandlerApi) InvoiceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoiceListRequest{
		Page:     1,
		PageSize: 10,
	}
	if r.ContentLength > 0 {
		err := json.NewDecoder(r.Body).Decode(requestDTO)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}

	q := r.URL.Query()
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			requestDTO.Page = p
		}
	}
	if pageSizeStr := q.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			requestDTO.PageSize = ps
		}
	}
	if sd := q.Get("start_date"); sd != "" {
		requestDTO.StartDate = sd
	}
	if ed := q.Get("end_date"); ed != "" {
		requestDTO.EndDate = ed
	}
	if invID := q.Get("invoice_id"); invID != "" {
		if id, err := strconv.ParseInt(invID, 10, 64); err == nil {
			requestDTO.InvoiceID = id
		}
	}
	if nick := q.Get("nickname"); nick != "" {
		requestDTO.Nickname = nick
	}
	if ptStr := q.Get("plan_type"); ptStr != "" {
		if pt, err := strconv.Atoi(ptStr); err == nil {
			requestDTO.PlanType = pt
		}
	}
	if svcID := q.Get("service_id"); svcID != "" {
		if sid, err := strconv.ParseInt(svcID, 10, 64); err == nil {
			requestDTO.ServiceID = sid
		}
	}
	if planID := q.Get("plan_id"); planID != "" {
		if pid, err := strconv.ParseInt(planID, 10, 64); err == nil {
			requestDTO.PlanID = pid
		}
	}

	svc := service.NewInvoiceList(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// Invoices is an alias for InvoiceList handler
func (h *HandlerApi) Invoices(w http.ResponseWriter, r *http.Request) {
	h.InvoiceList(w, r)
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
