package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"planner/internal/dto"
	"planner/internal/service"
)

// InvoicingHTML handles GET /invoicing and renders the main invoicing page or fragment
func (h *HandlerHtml) InvoicingHTML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:     1,
		PageSize: itensPorPag,
	}
	respdto := svc.Run(req)
	if respdto.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve invoices", http.StatusInternalServerError)
		return
	}
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getInvoicePageData(response)
	templateName := "index"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "fragment"
	}
	err := h.tmplInvoicing.ExecuteTemplate(w, templateName, page)
	if err != nil {
		h.logger.IPrintf(2, "Failed to render invoicing template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// InvoicingTable handles POST /invoicing/tabela and returns the updated table fragment
func (h *HandlerHtml) InvoicingTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina := 0
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		pagina, _ = strconv.Atoi(pStr)
	}
	if pagina <= 0 {
		pagina, _ = strconv.Atoi(r.FormValue("page"))
	}
	if pagina <= 0 {
		pagina = 1
	}
	tipoFiltro, _ := strconv.Atoi(r.FormValue("plan_type"))
	if tipoFiltro <= 0 {
		tipoFiltro, _ = strconv.Atoi(r.FormValue("tipo_filtro"))
	}
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("service_id"), 10, 64)
	if servicoFiltro <= 0 {
		servicoFiltro, _ = strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	}
	faturaID, _ := strconv.ParseInt(r.FormValue("invoice_id"), 10, 64)
	if faturaID <= 0 {
		faturaID, _ = strconv.ParseInt(r.FormValue("fatura_id"), 10, 64)
	}
	planoID, _ := strconv.ParseInt(r.FormValue("plan_id"), 10, 64)
	if planoID <= 0 {
		planoID, _ = strconv.ParseInt(r.FormValue("plano_id"), 10, 64)
	}

	startDate := r.FormValue("start_date")
	if startDate == "" {
		startDate = r.FormValue("data_inicio")
	}
	endDate := r.FormValue("end_date")
	if endDate == "" {
		endDate = r.FormValue("data_fim")
	}

	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		InvoiceID: faturaID,
		PlanID:    planoID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	respdto := svc.Run(req)
	if respdto.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve invoices", http.StatusInternalServerError)
		return
	}
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getInvoicePageData(response)
	h.tmplInvoicing.ExecuteTemplate(w, "tabela", page)
}

// InvoicingTableReset handles POST /invoicing/tabela/reset
func (h *HandlerHtml) InvoicingTableReset(w http.ResponseWriter, r *http.Request) {
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:     1,
		PageSize: itensPorPag,
	}
	respdto := svc.Run(req)
	if respdto.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve invoices", http.StatusInternalServerError)
		return
	}
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getInvoicePageData(response)
	h.tmplInvoicing.ExecuteTemplate(w, "tabela", page)
}

// InvoicingDeleteWarning handles GET /invoicing/deletar-aviso
func (h *HandlerHtml) InvoicingDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:      1,
		PageSize:  1,
		InvoiceID: id,
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.InvoiceListResponse)
	if !ok || len(response.Invoices) == 0 {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}
	invoiceSel := h.getInvoiceData(response)[0]
	h.tmplInvoicing.ExecuteTemplate(w, "linha_invoice_deletar_aviso", invoiceSel)
}

// InvoicingCancelDelete handles GET /invoicing/cancelar-delete
func (h *HandlerHtml) InvoicingCancelDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:      1,
		PageSize:  1,
		InvoiceID: id,
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.InvoiceListResponse)
	if !ok || len(response.Invoices) == 0 {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}
	invoiceSel := h.getInvoiceData(response)[0]
	h.tmplInvoicing.ExecuteTemplate(w, "linha_invoice", invoiceSel)
}

// InvoicingDelete handles POST /invoicing/deletar
func (h *HandlerHtml) InvoicingDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	if idStr == "" {
		idStr = r.FormValue("invoice_id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	svc := service.NewInvoiceDelete(h.repo, h.logger)
	req := &dto.InvoiceDeleteRequest{
		InvoiceID: id,
	}
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		http.Error(w, resp.GetMessage(), resp.GetStatusCode())
		return
	}

	// Remove single-invoice filters so InvoicingTable re-renders the complete filtered list
	r.Form.Del("invoice_id")
	r.Form.Del("fatura_id")
	r.Form.Del("id")

	h.InvoicingTable(w, r)
}

// InvoicingReconcile handles POST /invoicing/reconciliar
func (h *HandlerHtml) InvoicingReconcile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	if idStr == "" {
		idStr = r.FormValue("invoice_id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
		return
	}

	reconciler := h.getReconciler()
	if reconciler != nil {
		if err := reconciler.ReconcileInvoice(id); err != nil {
			h.logger.IPrintf(1, "Erro ao conciliar fatura %d: %v", id, err)
		}
	} else {
		h.logger.IPrintf(1, "ReconciliationService não disponível para fatura %d", id)
	}

	// Remove single-invoice filters so InvoicingTable re-renders the complete filtered list
	r.Form.Del("invoice_id")
	r.Form.Del("fatura_id")
	r.Form.Del("id")

	h.InvoicingTable(w, r)
}

// getInvoiceData prepares invoice rows and items for template rendering
func (h *HandlerHtml) getInvoiceData(response *dto.InvoiceListResponse) []map[string]interface{} {
	if response == nil || len(response.Invoices) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(response.Invoices))
	for _, inv := range response.Invoices {
		dataFaturaFormatada := inv.InvoiceDate
		if t, err := time.Parse("2006-01-02", inv.InvoiceDate); err == nil {
			dataFaturaFormatada = t.Format("02/01/2006")
		}
		vencimentoFormatado := inv.DueDate
		if t, err := time.Parse("2006-01-02", inv.DueDate); err == nil {
			vencimentoFormatado = t.Format("02/01/2006")
		}
		valorFormatado := fmt.Sprintf("%.2f", inv.Amount)

		itemsData := make([]map[string]interface{}, 0, len(inv.Items))
		for _, it := range inv.Items {
			subtotal := it.Price * float64(it.InvoicedQuantity)
			itemsData = append(itemsData, map[string]interface{}{
				"ID":                it.ID,
				"ServiceID":         it.ServiceID,
				"Description":       it.Description,
				"Price":             it.Price,
				"PrecoFormatado":    fmt.Sprintf("%.2f", it.Price),
				"InvoicedQuantity":  it.InvoicedQuantity,
				"RealizedQuantity":  it.RealizedQuantity,
				"SubtotalFormatado": fmt.Sprintf("%.2f", subtotal),
			})
		}

		periodStartFormatada := inv.PeriodStart
		if t, err := time.Parse("2006-01-02", inv.PeriodStart); err == nil {
			periodStartFormatada = t.Format("02/01/2006")
		}
		periodEndFormatada := inv.PeriodEnd
		if t, err := time.Parse("2006-01-02", inv.PeriodEnd); err == nil {
			periodEndFormatada = t.Format("02/01/2006")
		}

		result = append(result, map[string]interface{}{
			"ID":                   inv.ID,
			"InvoiceDate":          inv.InvoiceDate,
			"DataFaturaFormatada":  dataFaturaFormatada,
			"Nickname":             inv.Nickname,
			"PlanType":             inv.PlanType,
			"PlanTypeName":         inv.PlanTypeName,
			"DueDate":              inv.DueDate,
			"VencimentoFormatado":  vencimentoFormatado,
			"PeriodStart":          inv.PeriodStart,
			"PeriodStartFormatada": periodStartFormatada,
			"PeriodEnd":            inv.PeriodEnd,
			"PeriodEndFormatada":   periodEndFormatada,
			"Amount":               inv.Amount,
			"ValorFormatado":       valorFormatado,
			"Items":                itemsData,
		})
	}
	return result
}

// getInvoicePageData prepares pagination and options data for invoicing template
func (h *HandlerHtml) getInvoicePageData(response *dto.InvoiceListResponse) map[string]interface{} {
	page := 1
	totalPages := 1
	var totalRegistros int64 = 0
	var totalAmount float64 = 0
	if response != nil {
		page = response.Page
		totalPages = response.TotalPages
		if totalPages < 1 {
			totalPages = 1
		}
		totalRegistros = response.TotalCount
		totalAmount = response.TotalAmount
	}
	invoices := h.getInvoiceData(response)
	return map[string]interface{}{
		"Invoices":       invoices,
		"PaginaAtual":    page,
		"TotalPaginas":   totalPages,
		"TotalRegistros": totalRegistros,
		"TotalCount":     totalRegistros,
		"TotalAmount":    totalAmount,
		"TotalFormatado": formatMoney(totalAmount),
		"TemAnterior":    page > 1,
		"TemProximo":     page < totalPages,
		"PagAnterior":    page - 1,
		"PagProxima":     page + 1,
		"Nicknames":      h.getCustomersNicknames(),
		"Services":       h.getServices(),
	}
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
