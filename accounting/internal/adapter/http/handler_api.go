package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"accounting/internal/dto"
	"accounting/internal/port"
	"accounting/internal/service"
)

// HandlerApi handles JSON API endpoints
type HandlerApi struct {
	repo          port.Repository
	logger        port.Logger
	listService   *service.StatementListService
	syncService   *service.StatementSyncService
	actionService *service.AccountActionService
}

func NewHandlerApi(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
	actionService *service.AccountActionService,
) *HandlerApi {
	return &HandlerApi{
		repo:          repo,
		logger:        logger,
		listService:   listService,
		syncService:   syncService,
		actionService: actionService,
	}
}

// Ping returns health status
func (h *HandlerApi) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "accounting",
	})
}

// StatementList handles GET /api/statement/list
func (h *HandlerApi) StatementList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	accountID, _ := strconv.ParseInt(q.Get("account_id"), 10, 64)
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))

	req := &dto.StatementListRequest{
		AccountID: accountID,
		StartDate: q.Get("start_date"),
		EndDate:   q.Get("end_date"),
		Type:      q.Get("type"),
		Page:      page,
		PageSize:  pageSize,
	}

	resp := h.listService.Run(req)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// StatementSync handles POST /api/statement/sync
func (h *HandlerApi) StatementSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req dto.SyncStatementsRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	resp := h.syncService.Run(r.Context(), &req)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// AccountActionList handles GET /api/account/action?account_id=...
func (h *HandlerApi) AccountActionList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	accountID, _ := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 64)
	resp := h.actionService.ListByAccount(accountID)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// AccountActionDetail handles GET /api/account/action/{id}
func (h *HandlerApi) AccountActionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	resp := h.actionService.GetByID(id)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// AccountActionCreate handles POST /api/account/action
func (h *HandlerApi) AccountActionCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req dto.AccountActionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeResponse(w, http.StatusBadRequest, dto.NewResponseBase(400, "error", "JSON inválido"))
		return
	}
	resp := h.actionService.Create(&req)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// AccountActionUpdate handles PUT /api/account/action
func (h *HandlerApi) AccountActionUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req dto.AccountActionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeResponse(w, http.StatusBadRequest, dto.NewResponseBase(400, "error", "JSON inválido"))
		return
	}
	resp := h.actionService.Update(&req)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

// AccountActionDelete handles DELETE /api/account/action
func (h *HandlerApi) AccountActionDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	resp := h.actionService.Delete(id)
	h.writeResponse(w, resp.GetStatusCode(), resp)
}

func (h *HandlerApi) writeResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

