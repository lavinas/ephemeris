package http

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	if req.Page == 0 && r.URL.Query().Get("page") != "" {
		if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
			req.Page = page
		}
	}
	if req.PageSize == 0 && r.URL.Query().Get("page_size") != "" {
		if pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil {
			req.PageSize = pageSize
		}
	}
	if req.Vendor == nil && r.URL.Query().Get("vendor") != "" {
		v := r.URL.Query().Get("vendor")
		req.Vendor = &v
	}
	if req.Customer == nil && r.URL.Query().Get("customer") != "" {
		c := r.URL.Query().Get("customer")
		req.Customer = &c
	}
	if req.InvoiceDate == nil && r.URL.Query().Get("invoicing") != "" {
		inv := r.URL.Query().Get("invoicing")
		req.InvoiceDate = &inv
	}
	if req.DueDate == nil && r.URL.Query().Get("due") != "" {
		d := r.URL.Query().Get("due")
		req.DueDate = &d
	}
	if req.Overdue == nil && r.URL.Query().Get("overdue") != "" {
		if ov, err := strconv.ParseBool(r.URL.Query().Get("overdue")); err == nil {
			req.Overdue = &ov
		}
	}
	if req.PaymentDate == nil && r.URL.Query().Get("payment") != "" {
		p := r.URL.Query().Get("payment")
		req.PaymentDate = &p
	}
	if req.EmailSentDate == nil && r.URL.Query().Get("email_sent") != "" {
		es := r.URL.Query().Get("email_sent")
		req.EmailSentDate = &es
	}
	if req.WhatsappSentDate == nil && r.URL.Query().Get("whatsapp_sent") != "" {
		ws := r.URL.Query().Get("whatsapp_sent")
		req.WhatsappSentDate = &ws
	}
	if req.EmailReceiptDate == nil && r.URL.Query().Get("email_receipt") != "" {
		er := r.URL.Query().Get("email_receipt")
		req.EmailReceiptDate = &er
	}
	if req.WhatsappReceiptDate == nil && r.URL.Query().Get("whatsapp_receipt") != "" {
		wr := r.URL.Query().Get("whatsapp_receipt")
		req.WhatsappReceiptDate = &wr
	}
	if req.TaxDate == nil && r.URL.Query().Get("tax") != "" {
		t := r.URL.Query().Get("tax")
		req.TaxDate = &t
	}
	if req.CancellationDate == nil && r.URL.Query().Get("cancellation") != "" {
		c := r.URL.Query().Get("cancellation")
		req.CancellationDate = &c
	}
	if req.Notes == nil && r.URL.Query().Get("notes") != "" {
		n := r.URL.Query().Get("notes")
		req.Notes = &n
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
