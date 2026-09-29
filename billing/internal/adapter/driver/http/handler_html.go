package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"billing/internal/domain"
	"billing/internal/dto"
	"billing/internal/port"
	"billing/internal/service"
)

const (
	vendorID    = int64(1)
	itensPorPag = 10
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger port.Logger
	repo   port.Repository
	taxer  port.Taxer
	pixer  port.Pixer
	issuer port.Issuer
	tmpl   *template.Template
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer, tdata []byte) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
		"formatarMoeda": func(valor float64) string {
			return formatMoney(valor)
		},
	}

	tmpl, err := template.New("index").Funcs(funcMap).Parse(string(tdata))
	if err != nil {
		return nil, err
	}

	return &HandlerHtml{
		repo:   repo,
		logger: logger,
		taxer:  taxer,
		pixer:  pixer,
		issuer: issuer,
		tmpl:   tmpl,
	}, nil
}

// Index handles requests to the root "/" endpoint, serving the static dashboard page.
func (h *HandlerHtml) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "web/static/index.html")
}

// Ping handler for the /ping endpoint
func (h *HandlerHtml) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewPingService(h.logger)
	response := svc.Run(nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.GetStatusCode())
	w.Write([]byte("pong"))
}

// Invoices handles the main /html/invoices endpoint
func (h *HandlerHtml) Invoices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	page := h.loadPageData(1, "", "", "", "", "", "", "", "")

	templateName := "index"
	if r.Header.Get("HX-Request") == "true" {
		h.logger.IPrintf(2, "HTMX request detected for invoices, using fragment template")
		templateName = "fragment"
	} else {
		h.logger.IPrintf(2, "Standard request detected for invoices, using complete template")
	}

	err := h.tmpl.ExecuteTemplate(w, templateName, page)
	if err != nil {
		h.logger.IPrintf(1, "Failed to render invoices template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// InvoicesCreate renders the add invoice form fragment
func (h *HandlerHtml) InvoicesCreate(w http.ResponseWriter, r *http.Request) {
	hoje := time.Now().Format("2006-01-02")
	dueDate := time.Now().AddDate(0, 0, 15).Format("2006-01-02")
	dadosPadrao := map[string]interface{}{
		"InvoiceDate": hoje,
		"DueDate":     dueDate,
		"Nicknames":   h.getCustomersNicknames(),
		"Items": []map[string]interface{}{
			{
				"Description": "",
				"Quantity":    1,
				"Price":       "",
			},
		},
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
}

// InvoicesSave handles creating a new invoice with multiple items
func (h *HandlerHtml) InvoicesSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	customer := strings.TrimSpace(r.FormValue("add_customer"))
	invoiceDate := strings.TrimSpace(r.FormValue("add_invoicing"))
	dueDate := strings.TrimSpace(r.FormValue("add_due"))
	paymentDate := strings.TrimSpace(r.FormValue("add_payment"))
	emailSentDate := strings.TrimSpace(r.FormValue("add_email_sent"))
	emailReceiptDate := strings.TrimSpace(r.FormValue("add_email_receipt"))
	notes := strings.TrimSpace(r.FormValue("add_notes"))

	// Extract multiple items
	itemDescriptions := r.Form["item_descricao[]"]
	if len(itemDescriptions) == 0 {
		itemDescriptions = r.Form["item_descricao"]
	}
	itemQuantities := r.Form["item_quantidade[]"]
	if len(itemQuantities) == 0 {
		itemQuantities = r.Form["item_quantidade"]
	}
	itemPrices := r.Form["item_preco[]"]
	if len(itemPrices) == 0 {
		itemPrices = r.Form["item_preco"]
	}

	createItems := make([]*dto.InvoiceCreateItem, 0, len(itemDescriptions))
	formItems := make([]map[string]interface{}, 0, len(itemDescriptions))

	for i := range itemDescriptions {
		desc := strings.TrimSpace(itemDescriptions[i])
		qty := 1
		if i < len(itemQuantities) {
			qty, _ = strconv.Atoi(itemQuantities[i])
			if qty <= 0 {
				qty = 1
			}
		}
		var price float64
		if i < len(itemPrices) {
			priceStr := strings.ReplaceAll(itemPrices[i], ",", ".")
			price, _ = strconv.ParseFloat(priceStr, 64)
		}
		if desc != "" || price > 0 {
			createItems = append(createItems, &dto.InvoiceCreateItem{
				Description: desc,
				Quantity:    qty,
				Price:       price,
			})
			formItems = append(formItems, map[string]interface{}{
				"Description": desc,
				"Quantity":    qty,
				"Price":       price,
			})
		}
	}

	var pDate *string
	if paymentDate != "" {
		pDate = &paymentDate
	}
	var notesPtr *string
	if notes != "" {
		notesPtr = &notes
	}

	vendorNick := h.getVendorNickname()
	invoiceCreate := &dto.InvoiceCreate{
		Vendor:       vendorNick,
		Customer:     customer,
		InvoiceDate:  invoiceDate,
		DueDate:      dueDate,
		PaymentDate:  pDate,
		Notes:        notesPtr,
		InvoiceItems: createItems,
	}

	createReq := &dto.InvoiceCreateRequest{
		Items: []*dto.InvoiceCreate{invoiceCreate},
	}

	svc := service.NewInvoiceCreate(h.repo, h.logger)
	respOut := svc.Run(createReq)

	// Fetch current table based on filters
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		// Display error message and keep form inputs for user correction
		errMsg := "Erro ao criar invoice"
		if respBase, ok := respOut.(dto.InvoiceCreateResponse); ok && respBase.Message != "" {
			errMsg = respBase.Message
		}
		pageData["ErrorMessage"] = errMsg

		dadosForm := map[string]interface{}{
			"Customer":         customer,
			"InvoiceDate":      invoiceDate,
			"DueDate":          dueDate,
			"PaymentDate":      paymentDate,
			"EmailSentDate":    emailSentDate,
			"EmailReceiptDate": emailReceiptDate,
			"Notes":            notes,
			"Nicknames":        h.getCustomersNicknames(),
			"Items":            formItems,
		}

		var formBuf strings.Builder
		_ = h.tmpl.ExecuteTemplate(&formBuf, "formulario_cadastro", dadosForm)
		w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML">` + formBuf.String() + `</div>`))
		h.tmpl.ExecuteTemplate(w, "tabela", pageData)
		return
	}

	// Success: clear form container and re-render table
	w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML"></div>`))
	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesTable handles filter searches and pagination updates
func (h *HandlerHtml) InvoicesTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}

	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesTableReset resets the filter form and returns page 1
func (h *HandlerHtml) InvoicesTableReset(w http.ResponseWriter, r *http.Request) {
	pageData := h.loadPageData(1, "", "", "", "", "", "", "", "")
	w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesBlockClear returns an empty response to clear forms
func (h *HandlerHtml) InvoicesBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// InvoicesEdit loads the inline edit form for a single invoice row
func (h *HandlerHtml) InvoicesEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	invoice, err := h.repo.GetInvoice(id)
	if err != nil || invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	data := h.mapSingleInvoiceData(*invoice)
	h.tmpl.ExecuteTemplate(w, "linha_invoice_edit", data)
}

// InvoicesUpdate processes updating invoice date fields
func (h *HandlerHtml) InvoicesUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)

	vendorNick := h.getVendorNickname()
	req := &dto.InvoiceUpdateRequest{
		Vendor: vendorNick,
		ID:     id,
	}

	invoicing := strings.TrimSpace(r.FormValue("edit_invoicing"))
	if invoicing != "" {
		req.InvoiceDate = &invoicing
	}
	due := strings.TrimSpace(r.FormValue("edit_due"))
	if due != "" {
		req.DueDate = &due
	}
	payment := strings.TrimSpace(r.FormValue("edit_payment"))
	if payment != "" {
		req.PaymentDate = &payment
	}
	emailSent := strings.TrimSpace(r.FormValue("edit_email_sent"))
	if emailSent != "" {
		req.EmailSentDate = &emailSent
	}
	emailReceipt := strings.TrimSpace(r.FormValue("edit_email_receipt"))
	if emailReceipt != "" {
		req.EmailReceiptDate = &emailReceipt
	}

	svc := service.NewInvoiceUpdate(h.repo, h.logger)
	respOut := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		errMsg := "Falha ao atualizar invoice"
		if respBase, ok := respOut.(dto.InvoiceUpdateResponse); ok && respBase.Message != "" {
			errMsg = respBase.Message
		}
		pageData["ErrorMessage"] = errMsg
	}

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesCancelEdit cancels editing and re-renders the original row
func (h *HandlerHtml) InvoicesCancelEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	invoice, err := h.repo.GetInvoice(id)
	if err != nil || invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	data := h.mapSingleInvoiceData(*invoice)
	h.tmpl.ExecuteTemplate(w, "linha_invoice", data)
}

// InvoicesDeleteWarning displays confirmation before soft-deleting/canceling an invoice
func (h *HandlerHtml) InvoicesDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	invoice, err := h.repo.GetInvoice(id)
	if err != nil || invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	data := h.mapSingleInvoiceData(*invoice)
	h.tmpl.ExecuteTemplate(w, "linha_invoice_deletar_aviso", data)
}

// InvoicesDelete executes invoice cancellation by setting CancellationDate
func (h *HandlerHtml) InvoicesDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	vendorNick := h.getVendorNickname()
	nowStr := time.Now().Format("2006-01-02")

	req := &dto.InvoiceUpdateRequest{
		Vendor:           vendorNick,
		ID:               id,
		CancellationDate: &nowStr,
	}

	svc := service.NewInvoiceUpdate(h.repo, h.logger)
	respOut := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		errMsg := "Falha ao cancelar/deletar invoice"
		if respBase, ok := respOut.(dto.InvoiceUpdateResponse); ok && respBase.Message != "" {
			errMsg = respBase.Message
		}
		pageData["ErrorMessage"] = errMsg
	}

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesPaymentWarning displays confirmation before setting payment date
func (h *HandlerHtml) InvoicesPaymentWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	invoice, err := h.repo.GetInvoice(id)
	if err != nil || invoice == nil {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}

	data := h.mapSingleInvoiceData(*invoice)
	data["DataHoje"] = time.Now().Format("2006-01-02")
	h.tmpl.ExecuteTemplate(w, "linha_invoice_pagamento_aviso", data)
}

// InvoicesPay executes registering the payment date and sending the receipt email atomically
func (h *HandlerHtml) InvoicesPay(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	vendorNick := h.getVendorNickname()

	paymentDate := strings.TrimSpace(r.FormValue("payment_date"))
	if paymentDate == "" {
		paymentDate = time.Now().Format("2006-01-02")
	}

	req := &dto.InvoicePaymentRequest{
		Vendor:      vendorNick,
		ID:          id,
		PaymentDate: paymentDate,
	}

	svc := service.NewInvoicePayment(h.repo, h.logger, h.issuer, h.pixer)
	respOut := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		errMsg := "Falha ao registrar pagamento"
		if respBase, ok := respOut.(dto.InvoicePaymentResponse); ok && respBase.Message != "" {
			errMsg = respBase.Message
		}
		pageData["ErrorMessage"] = errMsg
	}

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesSendInvoice sends the invoice document via email (Doc = 0, Action = 0, Email = "")
func (h *HandlerHtml) InvoicesSendInvoice(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	vendorNick := h.getVendorNickname()

	req := &dto.BillRequest{
		Vendor:    vendorNick,
		InvoiceID: id,
		Doc:       0,
		Action:    0,
		Email:     "",
	}

	svc := service.NewBill(h.repo, h.logger, h.issuer, h.pixer)
	respOut := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		errMsg := "Falha ao enviar invoice por e-mail"
		if respBill, ok := respOut.(*dto.BillResponse); ok && respBill.Message != "" {
			errMsg = respBill.Message
		}
		pageData["ErrorMessage"] = errMsg
	}

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// InvoicesSendReceipt sends the receipt document via email (Doc = 1, Action = 0, Email = "")
func (h *HandlerHtml) InvoicesSendReceipt(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	vendorNick := h.getVendorNickname()

	req := &dto.BillRequest{
		Vendor:    vendorNick,
		InvoiceID: id,
		Doc:       1,
		Action:    0,
		Email:     "",
	}

	svc := service.NewBill(h.repo, h.logger, h.issuer, h.pixer)
	respOut := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	pageData := h.loadPageData(
		pagina,
		r.FormValue("customer"),
		r.FormValue("invoicing"),
		r.FormValue("due"),
		r.FormValue("payment"),
		r.FormValue("email_sent"),
		r.FormValue("email_receipt"),
		r.FormValue("item_desc"),
		r.FormValue("overdue"),
	)

	if respOut.GetStatusCode() != 200 {
		errMsg := "Falha ao enviar recibo por e-mail"
		if respBill, ok := respOut.(*dto.BillResponse); ok && respBill.Message != "" {
			errMsg = respBill.Message
		}
		pageData["ErrorMessage"] = errMsg
	}

	h.tmpl.ExecuteTemplate(w, "tabela", pageData)
}

// parseOverdueFilter parses overdue filter string to *bool
func parseOverdueFilter(overdue string) *bool {
	if overdue == "" {
		return nil
	}
	s := strings.ToLower(strings.TrimSpace(overdue))
	if s == "true" || s == "on" || s == "1" {
		val := true
		return &val
	}
	if s == "false" || s == "0" {
		val := false
		return &val
	}
	return nil
}

// loadPageData retrieves invoices filtered and calculates pagination metadata
func (h *HandlerHtml) loadPageData(page int, customer, invoiceDate, dueDate, paymentDate, emailSentDate, emailReceiptDate, itemDesc, overdue string) map[string]interface{} {
	if page <= 0 {
		page = 1
	}

	vendorNick := h.getVendorNickname()
	cancellationNull := "null" // only non-cancelled invoices

	var invoiceDatePtr *string
	if invoiceDate != "" {
		invoiceDatePtr = &invoiceDate
	}
	var dueDatePtr *string
	if dueDate != "" {
		dueDatePtr = &dueDate
	}
	var paymentDatePtr *string
	if paymentDate != "" {
		paymentDatePtr = &paymentDate
	}
	var emailSentDatePtr *string
	if emailSentDate != "" {
		emailSentDatePtr = &emailSentDate
	}
	var emailReceiptDatePtr *string
	if emailReceiptDate != "" {
		emailReceiptDatePtr = &emailReceiptDate
	}

	// Resolve Customer ID if customer nickname is provided
	var customerID int64
	if customer != "" {
		cust, err := h.repo.GetCustomer(vendorID, customer)
		if err == nil && cust != nil {
			customerID = cust.ID
		}
	}

	// Fetch all matching un-paginated to compute total count accurately
	allInvoices, err := h.repo.FindInvoices(0, 0, customerID, invoiceDatePtr, dueDatePtr, paymentDatePtr, emailSentDatePtr, nil, emailReceiptDatePtr, nil, nil, &cancellationNull)
	if err != nil {
		h.logger.IPrintf(1, "Error searching invoices: %v", err)
	}

	// Filter by item description in memory if requested
	if itemDesc != "" {
		filtered := make([]domain.Invoice, 0)
		for _, inv := range allInvoices {
			match := false
			for _, itm := range inv.InvoiceItems {
				if strings.Contains(strings.ToLower(itm.Description), strings.ToLower(itemDesc)) {
					match = true
					break
				}
			}
			if match {
				filtered = append(filtered, inv)
			}
		}
		allInvoices = filtered
	}

	// Filter by overdue in memory if requested
	if ov := parseOverdueFilter(overdue); ov != nil {
		filtered := make([]domain.Invoice, 0)
		for _, inv := range allInvoices {
			if inv.IsOverdue() == *ov {
				filtered = append(filtered, inv)
			}
		}
		allInvoices = filtered
	}

	totalItems := len(allInvoices)
	var totalAmount float64
	for _, inv := range allInvoices {
		totalAmount += inv.Amount
	}

	totalPages := (totalItems + itensPorPag - 1) / itensPorPag
	if totalPages <= 0 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}

	// Paginate the slice
	start := (page - 1) * itensPorPag
	end := start + itensPorPag
	if start > totalItems {
		start = totalItems
	}
	if end > totalItems {
		end = totalItems
	}

	var pagedInvoices []domain.Invoice
	if start < totalItems {
		pagedInvoices = allInvoices[start:end]
	}

	// Map to view models
	mappedInvoices := make([]map[string]interface{}, len(pagedInvoices))
	for i, inv := range pagedInvoices {
		mappedInvoices[i] = h.mapSingleInvoiceData(inv)
	}

	return map[string]interface{}{
		"Invoices":       mappedInvoices,
		"PaginaAtual":    page,
		"TotalPaginas":   totalPages,
		"TotalCount":     totalItems,
		"TotalAmount":    totalAmount,
		"TotalFormatado": formatMoney(totalAmount),
		"TemAnterior":    page > 1,
		"TemProximo":     page < totalPages,
		"PagAnterior":    page - 1,
		"PagProxima":     page + 1,
		"Nicknames":      h.getCustomersNicknames(),
		"Vendor":         vendorNick,
	}
}

// mapSingleInvoiceData converts a domain.Invoice into a view-ready map
func (h *HandlerHtml) mapSingleInvoiceData(inv domain.Invoice) map[string]interface{} {
	itemsView := make([]map[string]interface{}, len(inv.InvoiceItems))
	for j, it := range inv.InvoiceItems {
		itemsView[j] = map[string]interface{}{
			"ID":             it.ID,
			"Description":    it.Description,
			"Quantity":       it.Quantity,
			"Price":          it.Price,
			"PrecoFormatado": formatMoney(it.Price),
			"Total":          float64(it.Quantity) * it.Price,
		}
	}

	paymentDateRaw := ""
	if inv.PaymentDate != nil {
		paymentDateRaw = inv.PaymentDate.Format("2006-01-02")
	}
	emailSentDateRaw := ""
	if inv.EmailSentDate != nil {
		emailSentDateRaw = inv.EmailSentDate.Format("2006-01-02")
	}
	emailReceiptDateRaw := ""
	if inv.EmailReceiptDate != nil {
		emailReceiptDateRaw = inv.EmailReceiptDate.Format("2006-01-02")
	}
	notesStr := ""
	if inv.Notes != nil {
		notesStr = *inv.Notes
	}

	return map[string]interface{}{
		"ID":                        inv.ID,
		"Customer":                  inv.Customer.Nickname,
		"DataInvoice":               inv.InvoiceDate.Format("2006-01-02"),
		"DataInvoiceFormatada":      inv.InvoiceDate.Format("02/01/2006"),
		"DueDate":                   inv.DueDate.Format("2006-01-02"),
		"DueDateFormatada":          inv.DueDate.Format("02/01/2006"),
		"Overdue":                   inv.IsOverdue(),
		"PaymentDateRaw":            paymentDateRaw,
		"PaymentDateFormatada":      formatDate(paymentDateRaw),
		"EmailSentDateRaw":          emailSentDateRaw,
		"EmailSentDateFormatada":    formatDate(emailSentDateRaw),
		"EmailReceiptDateRaw":       emailReceiptDateRaw,
		"EmailReceiptDateFormatada": formatDate(emailReceiptDateRaw),
		"ValorTotal":                inv.Amount,
		"ValorFormatado":            formatMoney(inv.Amount),
		"Notes":                     notesStr,
		"Items":                     itemsView,
	}
}

// getCustomersNicknames retrieves all customer nicknames for vendor 1
func (h *HandlerHtml) getCustomersNicknames() []string {
	customers, err := h.repo.FindCustomers(0, 0, vendorID, nil, nil, nil, nil, nil, nil)
	if err != nil {
		h.logger.IPrintf(2, "Failed to retrieve customers: %v", err)
		return nil
	}
	nicknames := make([]string, 0, len(customers))
	for _, c := range customers {
		nicknames = append(nicknames, c.Nickname)
	}
	return nicknames
}

// getVendorNickname retrieves the vendor nickname corresponding to vendorID
func (h *HandlerHtml) getVendorNickname() string {
	vendors, err := h.repo.FindVendors(0, 0, nil, nil, nil, nil, nil, nil)
	if err == nil {
		for _, v := range vendors {
			if v.ID == vendorID {
				return v.Nickname
			}
		}
		if len(vendors) > 0 {
			return vendors[0].Nickname
		}
	}
	return "estudio_amelia"
}

// Helper formatting functions
func formatDate(dateStr string) string {
	if dateStr == "" || dateStr == "-" {
		return "-"
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("02/01/2006")
}

func formatMoney(amount float64) string {
	s := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	decPart := parts[1]

	var result []string
	for i, c := range reverse(intPart) {
		if i > 0 && i%3 == 0 {
			result = append(result, ".")
		}
		result = append(result, string(c))
	}
	return "R$ " + reverse(strings.Join(result, "")) + "," + decPart
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
