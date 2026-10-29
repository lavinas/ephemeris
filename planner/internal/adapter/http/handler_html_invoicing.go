package http

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"planner/internal/domain"
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
		Page:        1,
		PageSize:    itensPorPag,
		PendingOnly: true,
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

	pendingOnly := true
	if r.FormValue("form_filtro_enviado") == "1" {
		pendingOnly = r.FormValue("pendente") == "true" || r.FormValue("pendente") == "on" || r.FormValue("pendente") == "1"
	} else if r.URL.Query().Get("pendente") != "" || r.FormValue("pendente") != "" {
		pVal := r.FormValue("pendente")
		if pVal == "" {
			pVal = r.URL.Query().Get("pendente")
		}
		pendingOnly = pVal == "true" || pVal == "on" || pVal == "1"
	}

	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:        pagina,
		PageSize:    itensPorPag,
		Nickname:    r.FormValue("nickname"),
		PlanType:    tipoFiltro,
		ServiceID:   servicoFiltro,
		InvoiceID:   faturaID,
		PlanID:      planoID,
		StartDate:   startDate,
		EndDate:     endDate,
		PendingOnly: pendingOnly,
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
		Page:        1,
		PageSize:    itensPorPag,
		PendingOnly: true,
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

// InvoicingEdit handles GET /invoicing/editar and renders the editable row linha_invoice_edit
func (h *HandlerHtml) InvoicingEdit(w http.ResponseWriter, r *http.Request) {
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
	respdto := svc.Run(req)
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok || len(response.Invoices) == 0 {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}
	invoiceSel := h.getInvoiceData(response)[0]
	h.tmplInvoicing.ExecuteTemplate(w, "linha_invoice_edit", invoiceSel)
}

// InvoicingCancelEdit handles GET /invoicing/cancelar-edicao and restores the read-only row linha_invoice
func (h *HandlerHtml) InvoicingCancelEdit(w http.ResponseWriter, r *http.Request) {
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
	respdto := svc.Run(req)
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok || len(response.Invoices) == 0 {
		http.Error(w, "Invoice not found", http.StatusNotFound)
		return
	}
	invoiceSel := h.getInvoiceData(response)[0]
	h.tmplInvoicing.ExecuteTemplate(w, "linha_invoice", invoiceSel)
}

// InvoicingUpdate handles POST /invoicing/atualizar and updates due date and items' invoiced quantities
func (h *HandlerHtml) InvoicingUpdate(w http.ResponseWriter, r *http.Request) {
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

	dueDate := r.FormValue("due_date")
	if dueDate == "" {
		dueDate = r.FormValue("vencimento")
	}

	var items []dto.InvoiceItemUpdateRequest
	for k, v := range r.Form {
		if strings.HasPrefix(k, "item_qty_") {
			itemIDStr := strings.TrimPrefix(k, "item_qty_")
			itemID, _ := strconv.ParseInt(itemIDStr, 10, 64)
			if itemID > 0 && len(v) > 0 {
				qty, _ := strconv.Atoi(v[0])
				items = append(items, dto.InvoiceItemUpdateRequest{
					ID:               itemID,
					InvoicedQuantity: qty,
				})
			}
		}
	}

	svc := service.NewInvoiceUpdate(h.repo, h.logger, h.getBillingClient())
	req := &dto.InvoiceUpdateRequest{
		InvoiceID: id,
		DueDate:   dueDate,
		Items:     items,
	}
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		http.Error(w, resp.GetMessage(), resp.GetStatusCode())
		return
	}

	// Clean single-invoice filter so table reloads properly with existing filters
	r.Form.Del("invoice_id")
	r.Form.Del("fatura_id")
	r.Form.Del("id")

	h.InvoicingTable(w, r)
}

// InvoicingDeleteBatchWarning handles POST /invoicing/deletar-lote-aviso and renders the confirmation modal
func (h *HandlerHtml) InvoicingDeleteBatchWarning(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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
		Page:        1,
		PageSize:    1,
		Nickname:    r.FormValue("nickname"),
		PlanType:    tipoFiltro,
		ServiceID:   servicoFiltro,
		InvoiceID:   faturaID,
		PlanID:      planoID,
		StartDate:   startDate,
		EndDate:     endDate,
		PendingOnly: true,
	}
	respdto := svc.Run(req)
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok {
		http.Error(w, "Failed to retrieve invoice count", http.StatusInternalServerError)
		return
	}

	var filtrosResumo []string
	if nick := strings.TrimSpace(r.FormValue("nickname")); nick != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Cliente: %s", nick))
	}
	if startDate != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("A partir de: %s", startDate))
	}
	if endDate != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Até: %s", endDate))
	}
	if tipoFiltro > 0 {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Tipo de Plano: %s", domain.PlanType(tipoFiltro).String()))
	}
	if servicoFiltro > 0 {
		for _, s := range h.getServices() {
			if s.ID == servicoFiltro {
				filtrosResumo = append(filtrosResumo, fmt.Sprintf("Serviço: %s", s.Name))
				break
			}
		}
	}

	data := map[string]interface{}{
		"TotalCount":     response.TotalCount,
		"TotalAmount":    response.TotalAmount,
		"TotalFormatado": formatMoney(response.TotalAmount),
		"Filtros":        filtrosResumo,
		"TemFiltros":     len(filtrosResumo) > 0,
		"Nickname":       r.FormValue("nickname"),
		"StartDate":      startDate,
		"EndDate":        endDate,
		"PlanType":       tipoFiltro,
		"ServiceID":      servicoFiltro,
		"InvoiceID":      faturaID,
		"PlanID":         planoID,
	}

	h.tmplInvoicing.ExecuteTemplate(w, "modal_deletar_invoices_lote", data)
}

// InvoicingDeleteBatch handles POST /invoicing/deletar-lote and deletes all invoices matching current filters
func (h *HandlerHtml) InvoicingDeleteBatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
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

	svc := service.NewInvoiceBatchDelete(h.repo, h.logger)
	req := &dto.InvoiceBatchDeleteRequest{
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		InvoiceID: faturaID,
		PlanID:    planoID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		http.Error(w, resp.GetMessage(), resp.GetStatusCode())
		return
	}

	r.Form.Set("page", "1")
	h.InvoicingTable(w, r)
}

// InvoicingBill handles POST /invoicing/faturar for a single invoice
func (h *HandlerHtml) InvoicingBill(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "ID de pré-fatura inválido", http.StatusBadRequest)
		return
	}

	svc := service.NewInvoiceBill(h.repo, h.logger, h.getBillingClient())
	req := &dto.InvoiceBillRequest{
		InvoiceID: id,
	}
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		http.Error(w, resp.GetMessage(), resp.GetStatusCode())
		return
	}

	h.InvoicingTable(w, r)
}

// InvoicingBillBatchWarning handles POST /invoicing/faturar-lote-aviso and renders the confirmation modal
func (h *HandlerHtml) InvoicingBillBatchWarning(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
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

	// Fetch all matching invoices to check pending and already invoiced
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:      1,
		PageSize:  1000,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		InvoiceID: faturaID,
		PlanID:    planoID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	respdto := svc.Run(req)
	response, ok := respdto.(*dto.InvoiceListResponse)
	if !ok {
		http.Error(w, "Failed to retrieve invoice count", http.StatusInternalServerError)
		return
	}

	var pendingIDs []int64
	var pendingAmount float64
	alreadyInvoicedCount := 0

	if response != nil {
		for _, inv := range response.Invoices {
			if inv.Invoiced {
				alreadyInvoicedCount++
			} else {
				pendingIDs = append(pendingIDs, inv.ID)
				pendingAmount += inv.Amount
			}
		}
	}

	var filtrosResumo []string
	if nick := strings.TrimSpace(r.FormValue("nickname")); nick != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Cliente: %s", nick))
	}
	if startDate != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("A partir de: %s", startDate))
	}
	if endDate != "" {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Até: %s", endDate))
	}
	if tipoFiltro > 0 {
		filtrosResumo = append(filtrosResumo, fmt.Sprintf("Tipo de Plano: %s", domain.PlanType(tipoFiltro).String()))
	}
	if servicoFiltro > 0 {
		for _, s := range h.getServices() {
			if s.ID == servicoFiltro {
				filtrosResumo = append(filtrosResumo, fmt.Sprintf("Serviço: %s", s.Name))
				break
			}
		}
	}

	data := map[string]interface{}{
		"TotalCount":           response.TotalCount,
		"TotalAmount":          response.TotalAmount,
		"TotalFormatado":       formatMoney(response.TotalAmount),
		"PendingCount":         len(pendingIDs),
		"PendingAmount":        pendingAmount,
		"PendingFormatado":     formatMoney(pendingAmount),
		"AlreadyInvoicedCount": alreadyInvoicedCount,
		"Filtros":              filtrosResumo,
		"TemFiltros":           len(filtrosResumo) > 0,
		"Nickname":             r.FormValue("nickname"),
		"StartDate":            startDate,
		"EndDate":              endDate,
		"PlanType":             tipoFiltro,
		"ServiceID":            servicoFiltro,
		"InvoiceID":            faturaID,
		"PlanID":               planoID,
	}

	h.tmplInvoicing.ExecuteTemplate(w, "modal_faturar_invoices_lote", data)
}

// InvoicingBillBatch handles POST /invoicing/faturar-lote and bills all pending invoices matching current filters
func (h *HandlerHtml) InvoicingBillBatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
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

	// First query invoices matching filters to bill only pending
	var pendingIDs []int64
	listSvc := service.NewInvoiceList(h.repo, h.logger)
	listReq := &dto.InvoiceListRequest{
		Page:      1,
		PageSize:  1000,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		InvoiceID: faturaID,
		PlanID:    planoID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	listResp := listSvc.Run(listReq)
	if invResp, ok := listResp.(*dto.InvoiceListResponse); ok && invResp != nil {
		for _, inv := range invResp.Invoices {
			if !inv.Invoiced {
				pendingIDs = append(pendingIDs, inv.ID)
			}
		}
	}

	if len(pendingIDs) == 0 {
		http.Error(w, "Nenhuma pré-fatura pendente de faturamento encontrada para os filtros selecionados", http.StatusBadRequest)
		return
	}

	svc := service.NewInvoiceBill(h.repo, h.logger, h.getBillingClient())
	req := &dto.InvoiceBillRequest{
		InvoiceIDs: pendingIDs,
	}
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		http.Error(w, resp.GetMessage(), resp.GetStatusCode())
		return
	}

	h.InvoicingTable(w, r)
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

	svc := service.NewInvoiceDelete(h.repo, h.logger, h.getBillingClient())
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

	if cs := h.getComptimeService(); cs != nil {
		if inv, err := h.repo.GetInvoiceByID(id); err == nil && inv != nil {
			for _, it := range inv.Items {
				_ = cs.ProcessInvoiceItemUpdated(r.Context(), it.ID)
			}
		}
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
			"CustomerName":         inv.CustomerName,
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
			"ReferenceInvoiceID":   inv.ReferenceInvoiceID,
			"Invoiced":             inv.Invoiced,
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

// InvoicingCreateModal handles GET /invoicing/novo and renders the manual invoice inclusion modal
func (h *HandlerHtml) InvoicingCreateModal(w http.ResponseWriter, r *http.Request) {
	data := h.getInvoiceModalCreateData(nil, "")
	h.tmplInvoicing.ExecuteTemplate(w, "modal_criar_invoice", data)
}

// InvoicingItemRow handles GET /invoicing/item-linha and renders a new empty row for invoice items
func (h *HandlerHtml) InvoicingItemRow(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Services": h.getServices(),
	}
	h.tmplInvoicing.ExecuteTemplate(w, "linha_item_invoice_nova", data)
}

// InvoicingSave handles POST /html/invoicing/salvar to create a new invoice manually
func (h *HandlerHtml) InvoicingSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dados de formulário inválidos", http.StatusBadRequest)
		return
	}

	nickname := strings.TrimSpace(r.FormValue("nickname"))
	invoiceDate := strings.TrimSpace(r.FormValue("invoice_date"))
	dueDate := strings.TrimSpace(r.FormValue("due_date"))
	periodStart := strings.TrimSpace(r.FormValue("period_start"))
	periodEnd := strings.TrimSpace(r.FormValue("period_end"))
	refStr := strings.TrimSpace(r.FormValue("reference_invoice_id"))
	var refID *int64
	if refStr != "" {
		if id, err := strconv.ParseInt(refStr, 10, 64); err == nil {
			refID = &id
		}
	}

	serviceIDs := r.Form["item_service_id"]
	prices := r.Form["item_price"]
	quantities := r.Form["item_quantity"]
	descriptions := r.Form["item_description"]

	var items []dto.InvoiceItemCreateRequest
	for i := 0; i < len(serviceIDs); i++ {
		sid, _ := strconv.ParseInt(serviceIDs[i], 10, 64)
		price := 0.0
		if i < len(prices) {
			cleanP := strings.ReplaceAll(strings.TrimSpace(prices[i]), ",", ".")
			price, _ = strconv.ParseFloat(cleanP, 64)
		}
		qty := 0
		if i < len(quantities) {
			qty, _ = strconv.Atoi(strings.TrimSpace(quantities[i]))
		}
		desc := ""
		if i < len(descriptions) {
			desc = strings.TrimSpace(descriptions[i])
		}

		if sid > 0 || price > 0 || qty > 0 || desc != "" {
			items = append(items, dto.InvoiceItemCreateRequest{
				ServiceID:   sid,
				Price:       price,
				Quantity:    qty,
				Description: desc,
			})
		}
	}

	req := &dto.InvoiceCreateRequest{
		Nickname:           nickname,
		InvoiceDate:        invoiceDate,
		DueDate:            dueDate,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		ReferenceInvoiceID: refID,
		Items:              items,
	}

	svc := service.NewInvoiceCreate(h.repo, h.logger)
	resp := svc.Run(req)
	if resp.GetStatusCode() != 201 {
		data := h.getInvoiceModalCreateData(req, resp.GetMessage())
		h.tmplInvoicing.ExecuteTemplate(w, "modal_criar_invoice", data)
		return
	}

	h.renderInvoicingTableWithMessages(w, r, "Fatura criada com sucesso!", "")
}

// InvoicingImportCSV handles POST /html/invoicing/importar-csv to import invoices from CSV
func (h *HandlerHtml) InvoicingImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		h.renderInvoicingTableWithMessages(w, r, "", "Erro ao processar formulário multipart: "+err.Error())
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
		h.renderInvoicingTableWithMessages(w, r, "", "Arquivo CSV não selecionado (campo 'file')")
		return
	}
	defer file.Close()

	req := &dto.InvoiceImportCSVRequest{
		Reader: file,
	}

	svc := service.NewInvoiceImportCSV(h.repo, h.logger)
	resp := svc.Run(req)
	if resp.GetStatusCode() != 200 {
		h.renderInvoicingTableWithMessages(w, r, "", resp.GetMessage())
		return
	}

	h.renderInvoicingTableWithMessages(w, r, resp.GetMessage(), "")
}

func (h *HandlerHtml) getInvoiceModalCreateData(req *dto.InvoiceCreateRequest, erro string) map[string]interface{} {
	now := time.Now()
	invoiceDate := now.Format("2006-01-02")
	dueDate := now.AddDate(0, 0, 3).Format("2006-01-02")
	firstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastDay := firstDay.AddDate(0, 1, -1)
	periodStart := firstDay.Format("2006-01-02")
	periodEnd := lastDay.Format("2006-01-02")
	nickname := ""
	var refID *int64
	var items []dto.InvoiceItemCreateRequest

	if req != nil {
		if req.InvoiceDate != "" {
			invoiceDate = req.InvoiceDate
		}
		if req.DueDate != "" {
			dueDate = req.DueDate
		}
		if req.PeriodStart != "" {
			periodStart = req.PeriodStart
		}
		if req.PeriodEnd != "" {
			periodEnd = req.PeriodEnd
		}
		nickname = req.Nickname
		refID = req.ReferenceInvoiceID
		items = req.Items
	}

	if len(items) == 0 {
		items = []dto.InvoiceItemCreateRequest{
			{Quantity: 1},
		}
	}

	return map[string]interface{}{
		"Nicknames":          h.getCustomersNicknames(),
		"Services":           h.getServices(),
		"InvoiceDate":        invoiceDate,
		"DueDate":            dueDate,
		"PeriodStart":        periodStart,
		"PeriodEnd":          periodEnd,
		"Nickname":           nickname,
		"ReferenceInvoiceID": refID,
		"Items":              items,
		"Erro":               erro,
	}
}

func (h *HandlerHtml) renderInvoicingTableWithMessages(w http.ResponseWriter, _ *http.Request, msgSucesso, msgErro string) {
	svc := service.NewInvoiceList(h.repo, h.logger)
	req := &dto.InvoiceListRequest{
		Page:        1,
		PageSize:    itensPorPag,
		PendingOnly: true,
	}
	respdto := svc.Run(req)
	var response *dto.InvoiceListResponse
	if respdto.GetStatusCode() == 200 {
		response, _ = respdto.(*dto.InvoiceListResponse)
	}
	page := h.getInvoicePageData(response)
	if msgSucesso != "" {
		page["MensagemSucesso"] = msgSucesso
	}
	if msgErro != "" {
		page["MensagemErro"] = msgErro
	}
	h.tmplInvoicing.ExecuteTemplate(w, "tabela", page)
}
