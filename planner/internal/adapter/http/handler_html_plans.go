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

// Plans handles GET /planos and renders the main plans page or fragment
func (h *HandlerHtml) Plans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewPlanList(h.repo, h.logger)
	active := true
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: itensPorPag,
		Active:   &active,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve plans", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPlanPageData(response)
	templateName := "index"
	if r.Header.Get("HX-Request") == "true" {
		templateName = "fragment"
	}
	err := h.tmplPlans.ExecuteTemplate(w, templateName, page)
	if err != nil {
		h.logger.IPrintf(2, "Failed to render plans template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// PlansCreate handles GET /planos/novo and returns the new plan form fragment
func (h *HandlerHtml) PlansCreate(w http.ResponseWriter, r *http.Request) {
	hoje := time.Now().Format("2006-01-02")
	services := h.getServices()
	var defaultServiceID int64
	if len(services) > 0 {
		defaultServiceID = services[0].ID
	}
	defaultItems := []map[string]interface{}{
		{
			"ServiceID":      defaultServiceID,
			"OrderIndex":     1,
			"PriceFormatado": "",
		},
	}
	dadosPadrao := map[string]interface{}{
		"PlanStartPadrao":  hoje,
		"PlanEndPadrao":    "",
		"Nicknames":        h.getCustomersNicknames(),
		"Services":         services,
		"DefaultServiceID": defaultServiceID,
		"PlanType":         1,
		"Items":            defaultItems,
	}
	h.tmplPlans.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
}

// PlansSave handles POST /planos/salvar
func (h *HandlerHtml) PlansSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	nickname := r.FormValue("add_nickname")
	planStart := r.FormValue("add_plan_start")
	planEndStr := r.FormValue("add_plan_end")
	var planEnd *string
	if planEndStr != "" {
		planEnd = &planEndStr
	}
	planType, _ := strconv.Atoi(r.FormValue("add_plan_type"))
	var price *float64
	if pStr := r.FormValue("add_price"); pStr != "" {
		if p, err := strconv.ParseFloat(pStr, 64); err == nil {
			price = &p
		}
	}

	services := r.Form["add_servico_id"]
	orders := r.Form["add_order_index"]
	prices := r.Form["add_item_price"]

	var items []dto.PlanItemRequest
	if planType != int(domain.PlanTypeManual) {
		for i, sIDStr := range services {
			if sID, err := strconv.ParseInt(sIDStr, 10, 64); err == nil && sID > 0 {
				orderIdx := i + 1
				if i < len(orders) && orders[i] != "" {
					if o, err := strconv.Atoi(orders[i]); err == nil && o > 0 {
						orderIdx = o
					}
				}
				var itemPrice *float64
				if price == nil && i < len(prices) && prices[i] != "" {
					if ip, err := strconv.ParseFloat(prices[i], 64); err == nil {
						itemPrice = &ip
					}
				}
				items = append(items, dto.PlanItemRequest{
					ServiceID:  sID,
					OrderIndex: orderIdx,
					Price:      itemPrice,
				})
			}
		}
	}

	var agenda *dto.PlanAgendaRequest
	var pkg *dto.PlanPackageRequest
	var notebook *dto.PlanNotebookRequest
	var payperview *dto.PlanPayPerViewRequest

	switch planType {
	case 1:
		rec, _ := strconv.Atoi(r.FormValue("add_agenda_recorrencia"))
		wDay, _ := strconv.Atoi(r.FormValue("add_agenda_dia_semana"))
		term, _ := strconv.Atoi(r.FormValue("add_agenda_vigencia"))
		payDay, _ := strconv.Atoi(r.FormValue("add_agenda_dia_pagamento"))
		var limit *int
		if lStr := r.FormValue("add_agenda_limite"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil {
				limit = &l
			}
		}
		agenda = &dto.PlanAgendaRequest{
			Recurrence:          rec,
			WeekDay:             wDay,
			TermType:            term,
			PaymentDay:          payDay,
			MonthlyServiceLimit: limit,
		}
	case 2:
		qty, _ := strconv.Atoi(r.FormValue("add_pacote_quantidade"))
		pDate := r.FormValue("add_pacote_data_pagamento")
		pkg = &dto.PlanPackageRequest{
			Quantity:    qty,
			PaymentDate: pDate,
		}
	case 3:
		d, _ := strconv.Atoi(r.FormValue("add_caderneta_dia_pagamento"))
		notebook = &dto.PlanNotebookRequest{
			PaymentDay: d,
		}
	case 4:
		term, _ := strconv.Atoi(r.FormValue("add_payperview_prazo_pagamento"))
		payperview = &dto.PlanPayPerViewRequest{
			PaymentTerm: term,
		}
	}

	createSvc := service.NewPlanCreate(h.repo, h.logger)
	createReq := &dto.PlanCreateRequest{
		Nickname:   nickname,
		PlanStart:  planStart,
		PlanEnd:    planEnd,
		PlanType:   planType,
		Price:      price,
		Items:      items,
		Agenda:     agenda,
		Package:    pkg,
		Notebook:   notebook,
		PayPerView: payperview,
	}
	respdro := createSvc.Run(createReq)

	// Fetch current table page with filters
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	listSvc := service.NewPlanList(h.repo, h.logger)
	var activeFilter *bool
	if r.FormValue("ativo") == "true" || r.FormValue("ativo") == "1" || r.FormValue("ativo") == "on" {
		t := true
		activeFilter = &t
	}
	listReq := &dto.PlanListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Active:    activeFilter,
	}
	listResp := listSvc.Run(listReq)
	pageResponse, _ := listResp.(*dto.PlanListResponse)
	pageData := h.getPlanPageData(pageResponse)

	if respdro.GetStatusCode() != 200 {
		pageData["ErrorMessage"] = respdro.GetMessage()
		var formItemsData []map[string]interface{}
		for _, it := range items {
			pFormatted := ""
			if it.Price != nil {
				pFormatted = fmt.Sprintf("%.2f", *it.Price)
			}
			formItemsData = append(formItemsData, map[string]interface{}{
				"ServiceID":      it.ServiceID,
				"OrderIndex":     it.OrderIndex,
				"PriceFormatado": pFormatted,
			})
		}
		var defaultServiceID int64
		if len(items) > 0 {
			defaultServiceID = items[0].ServiceID
		}
		formDados := map[string]interface{}{
			"ErrorMessage":     respdro.GetMessage(),
			"Nickname":         nickname,
			"PlanStartPadrao":  planStart,
			"PlanEndPadrao":    planEndStr,
			"PlanType":         planType,
			"PricePadrao":      r.FormValue("add_price"),
			"Nicknames":        h.getCustomersNicknames(),
			"Services":         h.getServices(),
			"DefaultServiceID": defaultServiceID,
			"Items":            formItemsData,
		}
		var formBuf strings.Builder
		_ = h.tmplPlans.ExecuteTemplate(&formBuf, "formulario_cadastro", formDados)
		w.Write([]byte(`<div id="formulario-cadastro-container-planos" hx-swap-oob="innerHTML">` + formBuf.String() + `</div>`))
		h.tmplPlans.ExecuteTemplate(w, "tabela", pageData)
		return
	}

	w.Write([]byte(`<div id="formulario-cadastro-container-planos" hx-swap-oob="innerHTML"></div>`))
	h.tmplPlans.ExecuteTemplate(w, "tabela", pageData)
}

// PlansTableReset handles POST /planos/tabela/reset
func (h *HandlerHtml) PlansTableReset(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	svc := service.NewPlanList(h.repo, h.logger)
	active := true
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: itensPorPag,
		Active:   &active,
	}
	respdro := svc.Run(req)
	response, _ := respdro.(*dto.PlanListResponse)
	page := h.getPlanPageData(response)
	w.Write([]byte(`<script>if(typeof resetFiltrosPlanos==='function'){resetFiltrosPlanos();}else{document.getElementById("filtro-form-planos").reset();document.getElementById("input-pagina-form-planos").value="1";}</script>`))
	h.tmplPlans.ExecuteTemplate(w, "tabela", page)
}

// PlansBlockClear handles GET /planos/bloco/limpar
func (h *HandlerHtml) PlansBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// PlansDeleteWarning handles GET /planos/deletar-aviso
func (h *HandlerHtml) PlansDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanList(h.repo, h.logger)
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: 1,
		PlanID:   int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok || len(response.Plans) == 0 {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}
	planSel := h.getPlanData(response)[0]
	h.tmplPlans.ExecuteTemplate(w, "linha_plano_deletar_aviso", planSel)
}

// PlansDelete handles POST /planos/deletar
func (h *HandlerHtml) PlansDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanDelete(h.repo, h.logger)
	req := &dto.PlanDeleteRequest{
		PlanID: int64(id),
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to delete plan", http.StatusInternalServerError)
		return
	}

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	svc2 := service.NewPlanList(h.repo, h.logger)
	req2 := &dto.PlanListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
	}
	respdro2 := svc2.Run(req2)
	response, _ := respdro2.(*dto.PlanListResponse)
	page := h.getPlanPageData(response)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container-planos").innerHTML = "";</script>`))
	h.tmplPlans.ExecuteTemplate(w, "tabela", page)
}

// PlansEndWarning handles GET /planos/encerrar-aviso
func (h *HandlerHtml) PlansEndWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanList(h.repo, h.logger)
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: 1,
		PlanID:   int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok || len(response.Plans) == 0 {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}
	planSel := h.getPlanData(response)[0]
	hoje := time.Now().Format("2006-01-02")
	dataPadrao := hoje
	if pStart, ok := planSel["PlanStart"].(string); ok && pStart > hoje {
		dataPadrao = pStart
	}
	planSel["DataHoje"] = dataPadrao
	h.tmplPlans.ExecuteTemplate(w, "linha_plano_encerrar_aviso", planSel)
}

// PlansEnd handles POST /planos/encerrar
func (h *HandlerHtml) PlansEnd(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	planEndStr := r.FormValue("encerrar_plan_end")
	if planEndStr == "" {
		planEndStr = r.FormValue("plan_end")
	}
	if planEndStr == "" {
		planEndStr = time.Now().Format("2006-01-02")
	}

	updateSvc := service.NewPlanUpdate(h.repo, h.logger)
	updateReq := &dto.PlanUpdateRequest{
		ID:      int64(id),
		PlanEnd: &planEndStr,
	}
	respdro := updateSvc.Run(updateReq)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, respdro.GetMessage(), http.StatusBadRequest)
		return
	}

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	listSvc := service.NewPlanList(h.repo, h.logger)
	var activeFilter *bool
	if r.FormValue("ativo") == "true" || r.FormValue("ativo") == "1" || r.FormValue("ativo") == "on" {
		t := true
		activeFilter = &t
	}
	listReq := &dto.PlanListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Active:    activeFilter,
	}
	listResp := listSvc.Run(listReq)
	response, _ := listResp.(*dto.PlanListResponse)
	pageData := h.getPlanPageData(response)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container-planos").innerHTML = "";</script>`))
	h.tmplPlans.ExecuteTemplate(w, "tabela", pageData)
}

// PlansEdit handles GET /planos/editar
func (h *HandlerHtml) PlansEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanList(h.repo, h.logger)
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: 1,
		PlanID:   int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok || len(response.Plans) == 0 {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}
	dadosPadrao := h.getPlanData(response)[0]
	dadosPadrao["Services"] = h.getServices()
	h.tmplPlans.ExecuteTemplate(w, "linha_plano_edit", dadosPadrao)
}

// PlansUpdate handles POST /planos/atualizar
func (h *HandlerHtml) PlansUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	planStart := r.FormValue("edit_plan_start")
	planEndStr := r.FormValue("edit_plan_end")
	clearPlanEnd := planEndStr == ""
	var planEnd *string
	if planEndStr != "" {
		planEnd = &planEndStr
	}

	var price *float64
	clearPrice := false
	pStr := r.FormValue("edit_price")
	if pStr != "" {
		if p, err := strconv.ParseFloat(pStr, 64); err == nil {
			price = &p
		}
	}

	services := r.Form["edit_servico_id"]
	orders := r.Form["edit_order_index"]
	prices := r.Form["edit_item_price"]

	hasAnyItemPrice := false
	for _, ipStr := range prices {
		if ipStr != "" {
			hasAnyItemPrice = true
			break
		}
	}
	if price == nil && hasAnyItemPrice {
		clearPrice = true
	}

	var items []dto.PlanItemRequest
	for i, sIDStr := range services {
		if sID, err := strconv.ParseInt(sIDStr, 10, 64); err == nil && sID > 0 {
			orderIdx := i + 1
			if i < len(orders) && orders[i] != "" {
				if o, err := strconv.Atoi(orders[i]); err == nil && o > 0 {
					orderIdx = o
				}
			}
			var itemPrice *float64
			if price == nil && i < len(prices) && prices[i] != "" {
				if ip, err := strconv.ParseFloat(prices[i], 64); err == nil {
					itemPrice = &ip
				}
			}
			items = append(items, dto.PlanItemRequest{
				ServiceID:  sID,
				OrderIndex: orderIdx,
				Price:      itemPrice,
			})
		}
	}

	var agenda *dto.PlanAgendaRequest
	if r.FormValue("edit_agenda_recorrencia") != "" {
		rec, _ := strconv.Atoi(r.FormValue("edit_agenda_recorrencia"))
		wDay, _ := strconv.Atoi(r.FormValue("edit_agenda_dia_semana"))
		term, _ := strconv.Atoi(r.FormValue("edit_agenda_vigencia"))
		payDay, _ := strconv.Atoi(r.FormValue("edit_agenda_dia_pagamento"))
		var limit *int
		if lStr := r.FormValue("edit_agenda_limite"); lStr != "" {
			if l, err := strconv.Atoi(lStr); err == nil {
				limit = &l
			}
		}
		agenda = &dto.PlanAgendaRequest{
			Recurrence:          rec,
			WeekDay:             wDay,
			TermType:            term,
			PaymentDay:          payDay,
			MonthlyServiceLimit: limit,
		}
	}

	var pkg *dto.PlanPackageRequest
	if r.FormValue("edit_pacote_quantidade") != "" {
		qty, _ := strconv.Atoi(r.FormValue("edit_pacote_quantidade"))
		pDate := r.FormValue("edit_pacote_data_pagamento")
		pkg = &dto.PlanPackageRequest{
			Quantity:    qty,
			PaymentDate: pDate,
		}
	}

	var notebook *dto.PlanNotebookRequest
	if dStr := r.FormValue("edit_caderneta_dia_pagamento"); dStr != "" {
		d, _ := strconv.Atoi(dStr)
		notebook = &dto.PlanNotebookRequest{
			PaymentDay: d,
		}
	}

	var payperview *dto.PlanPayPerViewRequest
	if tStr := r.FormValue("edit_payperview_prazo_pagamento"); tStr != "" {
		term, _ := strconv.Atoi(tStr)
		payperview = &dto.PlanPayPerViewRequest{
			PaymentTerm: term,
		}
	}

	updateSvc := service.NewPlanUpdate(h.repo, h.logger)
	updateReq := &dto.PlanUpdateRequest{
		ID:           int64(id),
		PlanStart:    planStart,
		PlanEnd:      planEnd,
		ClearPlanEnd: clearPlanEnd,
		Price:        price,
		ClearPrice:   clearPrice,
		Items:        items,
		Agenda:       agenda,
		Package:      pkg,
		Notebook:     notebook,
		PayPerView:   payperview,
	}
	respdro := updateSvc.Run(updateReq)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, respdro.GetMessage(), http.StatusBadRequest)
		return
	}

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	listSvc := service.NewPlanList(h.repo, h.logger)
	var activeFilter *bool
	if r.FormValue("ativo") == "true" || r.FormValue("ativo") == "1" || r.FormValue("ativo") == "on" {
		t := true
		activeFilter = &t
	}
	listReq := &dto.PlanListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Active:    activeFilter,
	}
	listResp := listSvc.Run(listReq)
	response, _ := listResp.(*dto.PlanListResponse)
	pageData := h.getPlanPageData(response)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container-planos").innerHTML = "";</script>`))
	h.tmplPlans.ExecuteTemplate(w, "tabela", pageData)
}

// PlansCancelEdit handles GET /planos/cancelar-edicao
func (h *HandlerHtml) PlansCancelEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanList(h.repo, h.logger)
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: 1,
		PlanID:   int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok || len(response.Plans) == 0 {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}
	planSel := h.getPlanData(response)[0]
	h.tmplPlans.ExecuteTemplate(w, "linha_plano", planSel)
}

// PlansTable handles POST /planos/tabela
func (h *HandlerHtml) PlansTable(w http.ResponseWriter, r *http.Request) {
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
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	svc := service.NewPlanList(h.repo, h.logger)
	var activeFilter *bool
	if r.FormValue("ativo") == "true" || r.FormValue("ativo") == "1" || r.FormValue("ativo") == "on" {
		t := true
		activeFilter = &t
	}
	req := &dto.PlanListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		PlanType:  tipoFiltro,
		ServiceID: servicoFiltro,
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Active:    activeFilter,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve plans", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPlanPageData(response)
	h.tmplPlans.ExecuteTemplate(w, "tabela", page)
}

// getPlanData retrieves formatted plan items for template rendering
func (h *HandlerHtml) getPlanData(response *dto.PlanListResponse) []map[string]interface{} {
	if response == nil || len(response.Plans) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(response.Plans))
	for _, p := range response.Plans {
		startFormatted := ""
		if t, err := time.Parse("2006-01-02", p.PlanStart); err == nil {
			startFormatted = t.Format("02/01/2006")
		}
		endFormatted := ""
		if p.PlanEnd != nil && *p.PlanEnd != "" {
			if t, err := time.Parse("2006-01-02", *p.PlanEnd); err == nil {
				endFormatted = t.Format("02/01/2006")
			}
		}
		priceFormatted := ""
		if p.Price != nil {
			priceFormatted = fmt.Sprintf("%.2f", *p.Price)
		}

		var itemPrice *float64
		itemPriceFormatted := ""
		if len(p.Items) > 0 && p.Items[0].Price != nil {
			itemPrice = p.Items[0].Price
			itemPriceFormatted = fmt.Sprintf("%.2f", *p.Items[0].Price)
		}

		var serviceID int64
		if len(p.Items) > 0 {
			serviceID = p.Items[0].ServiceID
		}

		var packageData map[string]interface{}
		if p.Package != nil {
			pDateFormatted := p.Package.PaymentDate
			if t, err := time.Parse("2006-01-02", p.Package.PaymentDate); err == nil {
				pDateFormatted = t.Format("02/01/2006")
			}
			packageData = map[string]interface{}{
				"Quantity":             p.Package.Quantity,
				"PaymentDate":          p.Package.PaymentDate,
				"PaymentDateFormatada": pDateFormatted,
			}
		}

		var itemsData []map[string]interface{}
		for _, it := range p.Items {
			pFormatted := ""
			if it.Price != nil {
				pFormatted = fmt.Sprintf("%.2f", *it.Price)
			}
			itemsData = append(itemsData, map[string]interface{}{
				"ID":             it.ID,
				"ServiceID":      it.ServiceID,
				"ServiceName":    it.ServiceName,
				"OrderIndex":     it.OrderIndex,
				"Price":          it.Price,
				"PriceFormatado": pFormatted,
			})
		}
		if len(itemsData) == 0 && serviceID > 0 {
			itemsData = append(itemsData, map[string]interface{}{
				"ID":             0,
				"ServiceID":      serviceID,
				"ServiceName":    "",
				"OrderIndex":     1,
				"Price":          nil,
				"PriceFormatado": "",
			})
		}

		m := map[string]interface{}{
			"ID":                 p.ID,
			"CustomerID":         p.CustomerID,
			"Nickname":           p.Nickname,
			"CustomerName":       p.CustomerName,
			"PlanStart":          p.PlanStart,
			"PlanStartFormatada": startFormatted,
			"PlanEnd":            p.PlanEnd,
			"PlanEndFormatada":   endFormatted,
			"PlanType":           p.PlanType,
			"PlanTypeName":       p.PlanTypeName,
			"Price":              p.Price,
			"PriceFormatado":     priceFormatted,
			"ItemPrice":          itemPrice,
			"ItemPriceFormatado": itemPriceFormatted,
			"Items":              itemsData,
			"ServiceID":          serviceID,
			"Agenda":             p.Agenda,
			"Package":            packageData,
			"Notebook":           p.Notebook,
			"PayPerView":         p.PayPerView,
			"Active":             p.Active,
		}
		result = append(result, m)
	}
	return result
}

// getPlanPageData prepares pagination and options data for plans template
func (h *HandlerHtml) getPlanPageData(response *dto.PlanListResponse) map[string]interface{} {
	page := 1
	totalPages := 1
	if response != nil {
		page = response.Page
		totalPages = response.TotalPages
		if totalPages < 1 {
			totalPages = 1
		}
	}
	plans := h.getPlanData(response)
	return map[string]interface{}{
		"Planos":       plans,
		"PaginaAtual":  page,
		"TotalPaginas": totalPages,
		"TemAnterior":  page > 1,
		"TemProximo":   page < totalPages,
		"PagAnterior":  page - 1,
		"PagProxima":   page + 1,
		"Nicknames":    h.getCustomersNicknames(),
		"Services":     h.getServices(),
	}
}

// PlansGenerateInvoiceModal handles GET /planos/gerar-invoice-modal
func (h *HandlerHtml) PlansGenerateInvoiceModal(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewPlanList(h.repo, h.logger)
	req := &dto.PlanListRequest{
		Page:     1,
		PageSize: 1,
		PlanID:   int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.PlanListResponse)
	if !ok || len(response.Plans) == 0 {
		http.Error(w, "Plan not found", http.StatusNotFound)
		return
	}
	planSel := h.getPlanData(response)[0]

	now := time.Now()
	inicioMes := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	fimMes := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Format("2006-01-02")

	data := map[string]interface{}{
		"Plan":             planSel,
		"ID":               planSel["ID"],
		"Nickname":         planSel["Nickname"],
		"PlanType":         planSel["PlanType"],
		"DataInicioPadrao": inicioMes,
		"DataFimPadrao":    fimMes,
	}
	h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoice", data)
}

// PlansGenerateInvoice handles POST /planos/gerar-invoice
func (h *HandlerHtml) PlansGenerateInvoice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	if idStr == "" {
		idStr = r.FormValue("plan_id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		http.Error(w, "ID do plano inválido", http.StatusBadRequest)
		return
	}

	startDate := strings.TrimSpace(r.FormValue("start_date"))
	endDate := strings.TrimSpace(r.FormValue("end_date"))
	reprocess := r.FormValue("reprocess") == "true" || r.FormValue("reprocess") == "1" || r.FormValue("reprocess") == "on"

	invoicingSvc := service.NewInvoicing(h.repo, h.logger)
	invReq := &dto.InvoicingRequest{
		PlanID:    &id,
		StartDate: &startDate,
		EndDate:   &endDate,
		Reprocess: &reprocess,
	}

	resp := invoicingSvc.Run(invReq)
	isSuccess := resp.GetStatusCode() == 200

	stDate, enDate, errDates := invReq.GetParsedDates(time.Now())
	displayStart := startDate
	displayEnd := endDate
	if errDates == nil {
		displayStart = stDate.Format("2006-01-02")
		displayEnd = enDate.Format("2006-01-02")
	}

	data := map[string]interface{}{
		"Success":   isSuccess,
		"Message":   resp.GetMessage(),
		"PlanID":    id,
		"StartDate": displayStart,
		"EndDate":   displayEnd,
	}
	h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoice_resultado", data)
}

// PlansGenerateInvoicesModal handles GET and POST /planos/gerar-invoices-modal
func (h *HandlerHtml) PlansGenerateInvoicesModal(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()

	now := time.Now()
	inicioMes := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	fimMes := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Format("2006-01-02")

	nickname := strings.TrimSpace(r.FormValue("nickname"))
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	ativo := r.FormValue("ativo")

	var tipoPlanoNome string
	switch tipoFiltro {
	case 1:
		tipoPlanoNome = "Agenda (Recorrente)"
	case 2:
		tipoPlanoNome = "Pacote"
	case 3:
		tipoPlanoNome = "Caderneta"
	case 4:
		tipoPlanoNome = "Pay-per-view"
	}

	var servicoNome string
	if servicoFiltro > 0 {
		if s, err := h.repo.GetServiceByID(servicoFiltro); err == nil && s != nil {
			servicoNome = s.Name
		}
	}

	data := map[string]interface{}{
		"DataInicioPadrao": inicioMes,
		"DataFimPadrao":    fimMes,
		"Nickname":         nickname,
		"TipoFiltro":       tipoFiltro,
		"TipoPlanoNome":    tipoPlanoNome,
		"ServicoFiltro":    servicoFiltro,
		"ServicoNome":      servicoNome,
		"Ativo":            ativo,
		"TemFiltros":       nickname != "" || tipoPlanoNome != "" || servicoNome != "",
	}

	h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoices", data)
}

// PlansGenerateInvoices handles POST /planos/gerar-invoices (batch invoicing by active filters)
func (h *HandlerHtml) PlansGenerateInvoices(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	invReq := &dto.InvoicingRequest{}

	// 1. Nickname -> CustomerID
	nickname := strings.TrimSpace(r.FormValue("nickname"))
	if nickname != "" {
		customers, err := h.repo.FindCustomers(0, 0, 1, nil, &nickname, nil, nil, nil, nil)
		if err != nil {
			h.logger.IPrintf(0, "Error finding customer by nickname '%s': %v", nickname, err)
			data := map[string]interface{}{
				"Success": false,
				"Message": fmt.Sprintf("Erro ao buscar cliente: %v", err),
			}
			h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoice_resultado", data)
			return
		}
		if len(customers) == 0 {
			data := map[string]interface{}{
				"Success": true,
				"Message": fmt.Sprintf("Nenhum cliente encontrado com o nickname '%s'. Nenhum plano faturado.", nickname),
			}
			h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoice_resultado", data)
			return
		}
		cID := customers[0].ID
		invReq.CustomerID = &cID
	}

	// 2. Tipo de Plano
	tipoFiltro, _ := strconv.Atoi(r.FormValue("tipo_filtro"))
	if tipoFiltro > 0 {
		invReq.PlanType = &tipoFiltro
	}

	// 3. Serviço
	servicoFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	if servicoFiltro > 0 {
		invReq.ServiceID = &servicoFiltro
	}

	// 4. Datas
	dataInicio := strings.TrimSpace(r.FormValue("start_date"))
	if dataInicio == "" {
		dataInicio = strings.TrimSpace(r.FormValue("data_inicio"))
	}
	if dataInicio != "" {
		invReq.StartDate = &dataInicio
	}
	dataFim := strings.TrimSpace(r.FormValue("end_date"))
	if dataFim == "" {
		dataFim = strings.TrimSpace(r.FormValue("data_fim"))
	}
	if dataFim != "" {
		invReq.EndDate = &dataFim
	}

	// 5. Reprocess
	reprocess := r.FormValue("reprocess") == "true" || r.FormValue("reprocess") == "1" || r.FormValue("reprocess") == "on"
	invReq.Reprocess = &reprocess

	invoicingSvc := service.NewInvoicing(h.repo, h.logger)
	resp := invoicingSvc.Run(invReq)
	isSuccess := resp.GetStatusCode() == 200

	stDate, enDate, errDates := invReq.GetParsedDates(time.Now())
	displayStart := dataInicio
	displayEnd := dataFim
	if errDates == nil {
		displayStart = stDate.Format("2006-01-02")
		displayEnd = enDate.Format("2006-01-02")
	}

	data := map[string]interface{}{
		"Success":   isSuccess,
		"Message":   resp.GetMessage(),
		"StartDate": displayStart,
		"EndDate":   displayEnd,
	}
	h.tmplPlans.ExecuteTemplate(w, "modal_gerar_invoice_resultado", data)
}
