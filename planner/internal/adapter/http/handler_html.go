package http

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"planner/internal/domain"
	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/service"
)

var (
	mu           sync.Mutex
	statusColors = map[string]string{
		"realizada":            "#10b981", // verde
		"cancelada_cobrar":     "#f59e0b", // laranja
		"cancelada_nao_cobrar": "#ef4444", // vermelho
	}
	itensPorPag = 10
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger    port.Logger
	repo      port.Repository
	tmpl      *template.Template
	tmplPlans *template.Template
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, tdata []byte, plansData ...[]byte) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
		"adicionar": func(a, b int) int {
			return a + b
		},
		"subtrair": func(a, b int) int {
			return a - b
		},
	}
	logger.IPrintf(0, "Passou")
	tmpl, err := template.New("index").Funcs(funcMap).Parse(string(tdata))
	if err != nil {
		return nil, err
	}

	var pData []byte
	if len(plansData) > 0 && len(plansData[0]) > 0 {
		pData = plansData[0]
	} else {
		if data, err := os.ReadFile("/app/web/templates/plans.html"); err == nil {
			pData = data
		} else if data, err := os.ReadFile("web/templates/plans.html"); err == nil {
			pData = data
		} else if data, err := os.ReadFile("../../../web/templates/plans.html"); err == nil {
			pData = data
		} else {
			pData = []byte(`{{define "index"}}<html><body>Plans</body></html>{{end}}{{define "fragment"}}<div>Plans</div>{{end}}{{define "tabela"}}<div>Tabela</div>{{end}}`)
		}
	}

	tmplPlans, err := template.New("index").Funcs(funcMap).Parse(string(pData))
	if err != nil {
		return nil, err
	}

	return &HandlerHtml{
		repo:      repo,
		logger:    logger,
		tmpl:      tmpl,
		tmplPlans: tmplPlans,
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
	service := service.NewPingService(h.logger)
	response := service.Run(nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.GetStatusCode())
	w.Write([]byte("pong"))
}

// Sessions handler for the /sessions endpoint
func (h *HandlerHtml) Sessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:     1,
		PageSize: itensPorPag,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)
	h.logger.IPrintf(2, "Rendering for %v", page)
	// Se a requisição vem do HTMX (header HX-Request), retorna apenas o fragmento
	// interno (sem html/head/body) para ser injetado no div do frame pai.
	// Se for acesso direto ao navegador, retorna a página completa.
	template := "index"
	if r.Header.Get("HX-Request") == "true" {
		h.logger.IPrintf(2, "HTMX request detected, using fragment template")
		template = "fragment"
	} else {
		h.logger.IPrintf(2, "Standard request detected, using complete template")
	}
	err := h.tmpl.ExecuteTemplate(w, template, page)
	if err != nil {
		h.logger.IPrintf(2, "Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// SessionsCreate handler for the /sessions/novo endpoint
func (h *HandlerHtml) SessionsCreate(w http.ResponseWriter, r *http.Request) {
	hoje := time.Now().Format("2006-01-02")
	services := h.getServices()
	duracaoPadrao := 60
	var defaultServiceID int64
	if len(services) > 0 {
		defaultServiceID = services[0].ID
		if services[0].SessionMinutes != nil && *services[0].SessionMinutes > 0 {
			duracaoPadrao = *services[0].SessionMinutes
		}
	}
	dadosPadrao := map[string]interface{}{
		"DataFormatada": hoje,
		"DataPadrao":    hoje,
		"DuracaoPadrao": duracaoPadrao,
		"Nicknames":     h.getCustomersNicknames(),
		"Services":      services,
		"ServiceID":     defaultServiceID,
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
}

// SessionsSave handler for the /sessions/salvar endpoint
func (h *HandlerHtml) SessionsSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	minutes, _ := strconv.Atoi(r.FormValue("add_duracao"))
	nickname := r.FormValue("add_nickname")
	status := r.FormValue("add_status")
	serviceID, _ := strconv.ParseInt(r.FormValue("add_servico_id"), 10, 64)
	if serviceID == 0 {
		serviceID, _ = strconv.ParseInt(r.FormValue("add_servico"), 10, 64)
	}
	comment := r.FormValue("add_comentario")
	svc := service.NewSessionCreate(h.repo, h.logger)
	req := &dto.SessionCreateRequest{
		Nickname:  nickname,
		Date:      r.FormValue("add_data"),
		Minutes:   minutes,
		ServiceID: serviceID,
		Status:    status,
		Comments:  comment,
	}
	respdro := svc.Run(req)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	dur, _ := strconv.Atoi(r.FormValue("duracao_filtro"))
	serviceIDFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	svc2 := service.NewSessionList(h.repo, h.logger)
	req2 := &dto.SessionListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Minutes:   dur,
		Status:    r.FormValue("status"),
		ServiceID: serviceIDFiltro,
		Comments:  r.FormValue("comentario"),
	}
	respdro2 := svc2.Run(req2)
	if respdro2.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro2.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)

	if respdro.GetStatusCode() != 200 {
		page["ErrorMessage"] = respdro.GetMessage()
		dadosForm := map[string]interface{}{
			"DataFormatada": r.FormValue("add_data"),
			"DataPadrao":    r.FormValue("add_data"),
			"DuracaoPadrao": r.FormValue("add_duracao"),
			"Nicknames":     h.getCustomersNicknames(),
			"Nickname":      nickname,
			"Status":        status,
			"Services":      h.getServices(),
			"ServiceID":     serviceID,
			"Comentario":    comment,
		}
		var formBuf strings.Builder
		_ = h.tmpl.ExecuteTemplate(&formBuf, "formulario_cadastro", dadosForm)
		w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML">` + formBuf.String() + `</div>`))
		h.tmpl.ExecuteTemplate(w, "tabela", page)
		return
	}

	h.logger.IPrintf(2, "Rendering 2 for %v", page)
	w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML"></div>`))
	h.tmpl.ExecuteTemplate(w, "tabela", page)
}

// SessionTableReset handler for the /sessions/table/reset endpoint
func (h *HandlerHtml) SessionsTableReset(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:     1,
		PageSize: itensPorPag,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)
	h.logger.IPrintf(2, "Rendering 3 for %v", page)
	w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela", page)
}

// SessionsBlockReset handler for the /sessions/bloco/limpar endpoint
func (h *HandlerHtml) SessionsBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// SessionDeleteWarning handler for the /sessions/deletar-aviso endpoint
func (h *HandlerHtml) SessionsDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:      1,
		PageSize:  itensPorPag,
		SessionID: int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok || len(response.Sessions) == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	sSel := h.getSessionData(response)[0]
	h.tmpl.ExecuteTemplate(w, "linha_sessao_deletar_aviso", sSel)
}

// SessionsDelete handler for the /sessions/deletar endpoint
func (h *HandlerHtml) SessionsDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewSessionDelete(h.repo, h.logger)
	req := &dto.SessionDeleteRequest{
		SessionID: int64(id),
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		return
	}
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	dur, _ := strconv.Atoi(r.FormValue("duracao_filtro"))
	serviceIDFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	svc2 := service.NewSessionList(h.repo, h.logger)
	req2 := &dto.SessionListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Minutes:   dur,
		Status:    r.FormValue("status"),
		ServiceID: serviceIDFiltro,
		Comments:  r.FormValue("comentario"),
	}
	respdro2 := svc2.Run(req2)
	if respdro2.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro2.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)
	h.logger.IPrintf(2, "Rendering 4 for %v", page)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela", page)
}

// SessionsEdit handler for the /sessions/editar endpoint
func (h *HandlerHtml) SessionsEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:      1,
		PageSize:  itensPorPag,
		SessionID: int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok || len(response.Sessions) == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	dadosPadrao := h.getSessionData(response)[0]
	dadosPadrao["Services"] = h.getServices()
	h.tmpl.ExecuteTemplate(w, "linha_sessao_edit", dadosPadrao)
}

// SessionsUpdate handler for the /sessions/atualizar endpoint.
func (h *HandlerHtml) SessionsUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	minutes, _ := strconv.Atoi(r.FormValue("edit_duracao"))

	status := r.FormValue("edit_status")
	serviceID, _ := strconv.ParseInt(r.FormValue("edit_servico_id"), 10, 64)
	if serviceID == 0 {
		serviceID, _ = strconv.ParseInt(r.FormValue("edit_servico"), 10, 64)
	}
	comment := r.FormValue("edit_comentario")
	svc := service.NewSessionUpdate(h.repo, h.logger)
	req := &dto.SessionUpdateRequest{
		ID:        int64(id),
		Date:      r.FormValue("edit_data"),
		Minutes:   minutes,
		ServiceID: serviceID,
		Status:    status,
		Comments:  comment,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to update session", http.StatusInternalServerError)
		return
	}
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	dur, _ := strconv.Atoi(r.FormValue("duracao_filtro"))
	serviceIDFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	svc2 := service.NewSessionList(h.repo, h.logger)
	req2 := &dto.SessionListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Minutes:   dur,
		Status:    r.FormValue("status"),
		ServiceID: serviceIDFiltro,
		Comments:  r.FormValue("comentario"),
	}
	respdro2 := svc2.Run(req2)
	if respdro2.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro2.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)
	h.logger.IPrintf(2, "Rendering 5 for %v", page)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela", page)
}

// SessionsCancelEdit handler for the /sessions/cancelar-edicao endpoint
func (h *HandlerHtml) SessionsCancelEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:      1,
		PageSize:  itensPorPag,
		SessionID: int64(id),
	}
	respdro := svc.Run(req)
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok || len(response.Sessions) == 0 {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	sSel := h.getSessionData(response)[0]
	h.tmpl.ExecuteTemplate(w, "linha_sessao", sSel)
}

// SessionsTable handler for the /sessions/tabela endpoint
func (h *HandlerHtml) SessionsTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	svc := service.NewSessionList(h.repo, h.logger)
	dur, _ := strconv.Atoi(r.FormValue("duracao_filtro"))
	serviceIDFiltro, _ := strconv.ParseInt(r.FormValue("servico_filtro"), 10, 64)
	req := &dto.SessionListRequest{
		Page:      pagina,
		PageSize:  itensPorPag,
		Nickname:  r.FormValue("nickname"),
		DateStart: r.FormValue("data_inicio"),
		DateEnd:   r.FormValue("data_fim"),
		Minutes:   dur,
		Status:    r.FormValue("status"),
		ServiceID: serviceIDFiltro,
		Comments:  r.FormValue("comentario"),
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.SessionListResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageData(response)
	h.logger.IPrintf(2, "Rendering 6 for %v", page)
	h.tmpl.ExecuteTemplate(w, "tabela", page)
}

// getSessionData retrieves the session data formatted for template rendering
func (h *HandlerHtml) getSessionData(response *dto.SessionListResponse) []map[string]interface{} {
	sessions := response.Sessions
	if len(sessions) == 0 {
		return nil
	}
	result := []map[string]interface{}{}
	for _, session := range sessions {
		formatedDate, _ := time.Parse("2006-01-02", session.Date)
		ss := map[string]interface{}{
			"ID":            session.SessionID,
			"Nickname":      session.Nickname,
			"Data":          session.Date,
			"DataFormatada": formatedDate.Format("02/01/2006"),
			"Duracao":       session.Minutes,
			"Status":        session.Status,
			"StatusCor":     statusColors[session.Status],
			"ServiceID":     session.ServiceID,
			"Servico":       session.Service,
			"Comentario":    session.Comments,
		}
		result = append(result, ss)
	}
	return result
}

// getPageData paginates the filtered sessions based on the target page and items per page
func (h *HandlerHtml) getPageData(response *dto.SessionListResponse) map[string]interface{} {
	page := response.Page
	totalPages := response.TotalPages
	pageSessoes := h.getSessionData(response)
	ret := map[string]interface{}{
		"Sessoes":      pageSessoes,
		"PaginaAtual":  page,
		"TotalPaginas": totalPages,
		"TemAnterior":  page > 1,
		"TemProximo":   page < totalPages,
		"PagAnterior":  page - 1,
		"PagProxima":   page + 1,
		"Nicknames":    h.getCustomersNicknames(),
		"Services":     h.getServices(),
	}
	return ret
}

// getCustomersNicknames retrieves all customer nicknames for vendor 1
func (h *HandlerHtml) getCustomersNicknames() []string {
	customers, err := h.repo.FindCustomers(0, 0, 1, nil, nil, nil, nil, nil, nil)
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

// getServices retrieves all services for vendor 1
func (h *HandlerHtml) getServices() []domain.Service {
	services, err := h.repo.FindServices(1)
	if err != nil {
		h.logger.IPrintf(2, "Failed to retrieve services: %v", err)
		return nil
	}
	return services
}

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
	var pkg *dto.PlanPackageRequest
	var notebook *dto.PlanNotebookRequest

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
		modo := r.FormValue("caderneta_modo")
		var pDay *int
		var pTerm *int
		if modo == "dia" {
			d, _ := strconv.Atoi(r.FormValue("add_caderneta_dia_pagamento"))
			pDay = &d
		} else if modo == "prazo" {
			t, _ := strconv.Atoi(r.FormValue("add_caderneta_prazo_pagamento"))
			pTerm = &t
		} else {
			if dStr := r.FormValue("add_caderneta_dia_pagamento"); dStr != "" {
				d, _ := strconv.Atoi(dStr)
				pDay = &d
			} else if tStr := r.FormValue("add_caderneta_prazo_pagamento"); tStr != "" {
				t, _ := strconv.Atoi(tStr)
				pTerm = &t
			}
		}
		notebook = &dto.PlanNotebookRequest{
			PaymentDay:  pDay,
			PaymentTerm: pTerm,
		}
	}

	createSvc := service.NewPlanCreate(h.repo, h.logger)
	createReq := &dto.PlanCreateRequest{
		Nickname:  nickname,
		PlanStart: planStart,
		PlanEnd:   planEnd,
		PlanType:  planType,
		Price:     price,
		Items:     items,
		Agenda:    agenda,
		Package:   pkg,
		Notebook:  notebook,
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
	dStr := r.FormValue("edit_caderneta_dia_pagamento")
	tStr := r.FormValue("edit_caderneta_prazo_pagamento")
	if dStr != "" || tStr != "" {
		var pDay *int
		var pTerm *int
		if dStr != "" {
			d, _ := strconv.Atoi(dStr)
			pDay = &d
		} else if tStr != "" {
			t, _ := strconv.Atoi(tStr)
			pTerm = &t
		}
		notebook = &dto.PlanNotebookRequest{
			PaymentDay:  pDay,
			PaymentTerm: pTerm,
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

