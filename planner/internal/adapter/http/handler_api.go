package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"planner/internal/adapter/billing"
	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/service"
)

// HandlerApi is an HTTP handler for the API
type HandlerApi struct {
	logger        port.Logger
	repo          port.Repository
	billingClient port.BillingClient
}

// NewHandlerApi creates a new instance of HandlerApi
func NewHandlerApi(repo port.Repository, logger port.Logger, billingClients ...port.BillingClient) *HandlerApi {
	var client port.BillingClient
	if len(billingClients) > 0 && billingClients[0] != nil {
		client = billingClients[0]
	}
	return &HandlerApi{
		repo:          repo,
		logger:        logger,
		billingClient: client,
	}
}

func (h *HandlerApi) getBillingClient() port.BillingClient {
	if h.billingClient != nil {
		return h.billingClient
	}
	return billing.NewClient(os.Getenv("BILLING_URL"), h.logger)
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

// SessionServices handler for the /session/services and /service/list endpoints
func (h *HandlerApi) SessionServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.SessionServicesRequest{
		VendorID: 1,
	}
	if r.Body != nil && r.Body != http.NoBody && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	q := r.URL.Query()
	if requestDTO.Nickname == "" {
		if nick := q.Get("nickname"); nick != "" {
			requestDTO.Nickname = nick
		} else if nick := q.Get("add_nickname"); nick != "" {
			requestDTO.Nickname = nick
		}
	}
	if vStr := q.Get("vendor_id"); vStr != "" {
		if vid, err := strconv.ParseInt(vStr, 10, 64); err == nil {
			requestDTO.VendorID = vid
		}
	}
	svc := service.NewSessionServices(h.repo, h.logger)
	response := svc.Run(requestDTO)
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
	svc := service.NewInvoiceDelete(h.repo, h.logger, h.getBillingClient())
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

// InvoiceUpdate handles updating an invoice (due date and invoiced quantities) via JSON API
func (h *HandlerApi) InvoiceUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoiceUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if idStr := r.URL.Query().Get("id"); idStr != "" && requestDTO.InvoiceID == 0 {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			requestDTO.InvoiceID = id
		}
	}
	svc := service.NewInvoiceUpdate(h.repo, h.logger, h.getBillingClient())
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// InvoiceBatchDelete handles deleting invoices in batch via JSON API
func (h *HandlerApi) InvoiceBatchDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoiceBatchDeleteRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
	}
	q := r.URL.Query()
	if requestDTO.StartDate == "" && q.Get("start_date") != "" {
		requestDTO.StartDate = q.Get("start_date")
	}
	if requestDTO.EndDate == "" && q.Get("end_date") != "" {
		requestDTO.EndDate = q.Get("end_date")
	}
	if requestDTO.InvoiceID == 0 && q.Get("invoice_id") != "" {
		if id, err := strconv.ParseInt(q.Get("invoice_id"), 10, 64); err == nil {
			requestDTO.InvoiceID = id
		}
	}
	if requestDTO.PlanID == 0 && q.Get("plan_id") != "" {
		if id, err := strconv.ParseInt(q.Get("plan_id"), 10, 64); err == nil {
			requestDTO.PlanID = id
		}
	}
	if requestDTO.Nickname == "" && q.Get("nickname") != "" {
		requestDTO.Nickname = q.Get("nickname")
	}
	if requestDTO.PlanType == 0 && q.Get("plan_type") != "" {
		if pt, err := strconv.Atoi(q.Get("plan_type")); err == nil {
			requestDTO.PlanType = pt
		}
	}
	if requestDTO.ServiceID == 0 && q.Get("service_id") != "" {
		if sid, err := strconv.ParseInt(q.Get("service_id"), 10, 64); err == nil {
			requestDTO.ServiceID = sid
		}
	}
	svc := service.NewInvoiceBatchDelete(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// InvoiceBill handles /invoice/bill or /invoicing/bill endpoint
func (h *HandlerApi) InvoiceBill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	requestDTO := &dto.InvoiceBillRequest{}
	if r.Body != nil && r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
			h.writeResponse(w, dto.NewInvoiceBillResponse(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error(), nil))
			return
		}
	}
	q := r.URL.Query()
	if requestDTO.InvoiceID == 0 && q.Get("id") != "" {
		if id, err := strconv.ParseInt(q.Get("id"), 10, 64); err == nil {
			requestDTO.InvoiceID = id
		}
	}
	if requestDTO.InvoiceID == 0 && q.Get("invoice_id") != "" {
		if id, err := strconv.ParseInt(q.Get("invoice_id"), 10, 64); err == nil {
			requestDTO.InvoiceID = id
		}
	}

	svc := service.NewInvoiceBill(h.repo, h.logger, h.getBillingClient())
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// ComptimeDelete handler for deleting a comptime record via JSON API (/comptime/delete)
func (h *HandlerApi) ComptimeDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestDTO := &dto.ComptimeDeleteRequest{}
	if r.Body != nil && r.Body != http.NoBody {
		_ = json.NewDecoder(r.Body).Decode(requestDTO)
	}
	if requestDTO.ID == 0 && r.URL.Query().Get("id") != "" {
		if id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64); err == nil {
			requestDTO.ID = id
		}
	}

	if err := requestDTO.Validate(); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", err.Error()))
		return
	}

	cr, ok := h.repo.(port.ComptimeRepository)
	if !ok {
		h.writeResponse(w, dto.NewResponseBase(http.StatusInternalServerError, "error", "ComptimeRepository não disponível"))
		return
	}

	svc := service.NewComptime(cr, h.logger)
	if err := svc.DeleteComptime(r.Context(), requestDTO.ID); err != nil {
		h.logger.IPrintf(1, "Erro ao excluir comptime %d: %v", requestDTO.ID, err)
		h.writeResponse(w, dto.NewResponseBase(http.StatusInternalServerError, "error", err.Error()))
		return
	}

	h.writeResponse(w, dto.NewResponseBase(http.StatusOK, "success", "Registro de banco de horas excluído com sucesso"))
}

// InvoiceCreate handles creating a new manual invoice via JSON API (/invoice/create or /invoicing/create)
func (h *HandlerApi) InvoiceCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestDTO := &dto.InvoiceCreateRequest{}
	if r.Body != nil && r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
			h.writeResponse(w, dto.NewInvoiceCreateResponse(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error(), nil))
			return
		}
	}

	svc := service.NewInvoiceCreate(h.repo, h.logger)
	response := svc.Run(requestDTO)
	h.writeResponse(w, response)
}

// InvoiceImportCSV handles importing invoices via CSV upload in multipart/form-data (/invoice/import-csv or /invoicing/import-csv)
func (h *HandlerApi) InvoiceImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestDTO *dto.InvoiceImportCSVRequest
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			h.writeResponse(w, dto.NewInvoiceImportCSVResponse(http.StatusBadRequest, "error", "Erro ao processar formulário multipart: "+err.Error(), 0, 0, 0))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			file, _, err = r.FormFile("arquivo")
			if err != nil {
				file, _, err = r.FormFile("csv")
			}
		}
		if err != nil {
			h.writeResponse(w, dto.NewInvoiceImportCSVResponse(http.StatusBadRequest, "error", "Arquivo CSV não fornecido no formulário (campo 'file')", 0, 0, 0))
			return
		}
		defer file.Close()
		requestDTO = &dto.InvoiceImportCSVRequest{Reader: file}
	} else if strings.Contains(contentType, "text/csv") || strings.Contains(contentType, "text/plain") {
		requestDTO = &dto.InvoiceImportCSVRequest{Reader: r.Body}
	} else {
		requestDTO = &dto.InvoiceImportCSVRequest{}
		if r.Body != nil && r.Body != http.NoBody {
			if err := json.NewDecoder(r.Body).Decode(requestDTO); err != nil {
				h.writeResponse(w, dto.NewInvoiceImportCSVResponse(http.StatusBadRequest, "error", "Formato inválido: "+err.Error(), 0, 0, 0))
				return
			}
		}
	}

	svc := service.NewInvoiceImportCSV(h.repo, h.logger)
	response := svc.Run(requestDTO)
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
