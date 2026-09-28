package http

import (
	"encoding/json"
	"net/http"

	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"
)

// HandlerApi is an HTTP handler for the API endpoints
type HandlerApi struct {
	logger port.Logger
	repo   port.Repository
	taxer  port.Taxer
	pixer  port.Pixer
	issuer port.Issuer
}

// NewHandlerApi creates a new instance of HandlerApi
func NewHandlerApi(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer) *HandlerApi {
	return &HandlerApi{
		repo:   repo,
		logger: logger,
		taxer:  taxer,
		pixer:  pixer,
		issuer: issuer,
	}
}

// Ping handles the /ping endpoint
func (h *HandlerApi) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewPingService(h.logger)
	response := svc.Run(nil)
	h.writeResponse(w, response)
}

// CustomerCreate handles /customer/create or /customer/api/create
func (h *HandlerApi) CustomerCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.CustomerCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewCustomerCreate(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// CustomerList handles /customer/list or /customer/api/list
func (h *HandlerApi) CustomerList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.CustomerListRequest{}
	if r.Body != nil && r.Body != http.NoBody {
		_ = json.NewDecoder(r.Body).Decode(req)
	}
	svc := service.NewCustomerList(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// CustomerUpdate handles /customer/update or /customer/api/update
func (h *HandlerApi) CustomerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.CustomerUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewCustomerUpdate(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// InvoiceCreate handles /invoice/create or /invoice/api/create
func (h *HandlerApi) InvoiceCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.InvoiceCreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewInvoiceCreate(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// InvoiceList handles /invoice/list or /invoice/api/list
func (h *HandlerApi) InvoiceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.InvoiceListRequest{}
	if r.Body != nil && r.Body != http.NoBody {
		_ = json.NewDecoder(r.Body).Decode(req)
	}
	svc := service.NewInvoiceList(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// InvoiceUpdate handles /invoice/update or /invoice/api/update
func (h *HandlerApi) InvoiceUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.InvoiceUpdateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewInvoiceUpdate(h.repo, h.logger)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// InvoiceBill handles /invoice/bill or /invoice/api/bill
func (h *HandlerApi) InvoiceBill(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.BillRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewBill(h.repo, h.logger, h.issuer, h.pixer)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// TaxGenerate handles /tax/generate or /tax/api/generate
func (h *HandlerApi) TaxGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.TaxGenerateRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewTaxGenerate(h.repo, h.logger, h.taxer)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// TaxClear handles /tax/clear or /tax/api/clear
func (h *HandlerApi) TaxClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req := &dto.TaxClearRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.writeResponse(w, dto.NewResponseBase(http.StatusBadRequest, "error", "Invalid JSON format: "+err.Error()))
		return
	}
	svc := service.NewTaxClear(h.repo, h.logger, h.taxer)
	resp := svc.Run(req)
	h.writeResponse(w, resp)
}

// writeResponse writes the JSON response to the client
func (h *HandlerApi) writeResponse(w http.ResponseWriter, response port.OutDTO) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		h.logger.IPrintf(1, "Failed to marshal response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.GetStatusCode())
	w.Write(responseJSON)
}
