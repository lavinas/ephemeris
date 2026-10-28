package http

import (
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"planner/internal/dto"
	"planner/internal/service"
)

var statusColors = map[string]string{
	"realizada":            "#10b981", // verde
	"cancelada_cobrar":     "#f59e0b", // laranja
	"cancelada_nao_cobrar": "#ef4444", // vermelho
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
	_ = r.ParseForm()
	nickname := strings.TrimSpace(r.URL.Query().Get("nickname"))
	if nickname == "" {
		nickname = strings.TrimSpace(r.FormValue("add_nickname"))
	}
	hoje := time.Now().Format("2006-01-02")
	services := h.getServicesForCustomer(nickname)
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
		"Nickname":      nickname,
		"Services":      services,
		"ServiceID":     defaultServiceID,
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
}

// SessionsServices handler for the /sessoes/servicos endpoint, returning the <option> list for a customer nickname.
func (h *HandlerHtml) SessionsServices(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	nickname := strings.TrimSpace(r.URL.Query().Get("nickname"))
	if nickname == "" {
		nickname = strings.TrimSpace(r.FormValue("add_nickname"))
	}
	if nickname == "" {
		nickname = strings.TrimSpace(r.FormValue("edit_nickname"))
	}
	if nickname == "" {
		nickname = strings.TrimSpace(r.FormValue("nickname"))
	}

	selectedIDStr := r.URL.Query().Get("service_id")
	if selectedIDStr == "" {
		selectedIDStr = r.FormValue("add_servico_id")
	}
	if selectedIDStr == "" {
		selectedIDStr = r.FormValue("edit_servico_id")
	}
	selectedServiceID, _ := strconv.ParseInt(selectedIDStr, 10, 64)

	services := h.getServicesForCustomer(nickname)
	foundSelected := false
	for _, s := range services {
		if s.ID == selectedServiceID {
			foundSelected = true
			break
		}
	}
	if !foundSelected {
		if len(services) > 0 {
			selectedServiceID = services[0].ID
		} else {
			selectedServiceID = 0
		}
	}

	data := map[string]interface{}{
		"Services":  services,
		"ServiceID": selectedServiceID,
	}
	if h.tmpl != nil && h.tmpl.Lookup("opcoes_servicos") != nil {
		_ = h.tmpl.ExecuteTemplate(w, "opcoes_servicos", data)
		return
	}

	var sb strings.Builder
	if selectedServiceID == 0 {
		sb.WriteString(`<option value="" disabled selected>Selecione o serviço...</option>`)
	} else {
		sb.WriteString(`<option value="" disabled>Selecione o serviço...</option>`)
	}
	for _, s := range services {
		mins := 60
		if s.SessionMinutes != nil && *s.SessionMinutes > 0 {
			mins = *s.SessionMinutes
		}
		selAttr := ""
		if s.ID == selectedServiceID {
			selAttr = " selected"
		}
		sb.WriteString(fmt.Sprintf(`<option value="%d" data-minutes="%d"%s>%s</option>`, s.ID, mins, selAttr, html.EscapeString(s.Name)))
	}
	w.Write([]byte(sb.String()))
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
			"Services":      h.getServicesForCustomer(nickname),
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

// SessionsTableReset handler for the /sessions/table/reset endpoint
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

// SessionsBlockClear handler for the /sessions/bloco/limpar endpoint
func (h *HandlerHtml) SessionsBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// SessionsDeleteWarning handler for the /sessions/deletar-aviso endpoint
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
	nickname, _ := dadosPadrao["Nickname"].(string)
	dadosPadrao["Services"] = h.getServicesForCustomer(nickname)
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
	if pagina <= 0 {
		if pStr := r.URL.Query().Get("page"); pStr != "" {
			pagina, _ = strconv.Atoi(pStr)
		}
	}
	if pagina <= 0 {
		pagina = 1
	}
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

// SessionsReconcile handles POST /sessoes/reconciliar
func (h *HandlerHtml) SessionsReconcile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.FormValue("id")
	}
	if idStr == "" {
		idStr = r.FormValue("session_id")
	}
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if id <= 0 {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	if cs := h.getComptimeService(); cs != nil {
		_ = cs.ProcessSessionUpdated(r.Context(), id)
	}

	// Remove single-session filters so SessionsTable re-renders the complete filtered list
	r.Form.Del("session_id")
	r.Form.Del("id")

	h.SessionsTable(w, r)
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
		status := strings.ToLower(strings.TrimSpace(session.Status))
		elegivel := status != "cancelada_nao_cobrar" && status != "cancelada/não cobrar"
		reconciled := false

		ss := map[string]interface{}{
			"ID":            session.SessionID,
			"Nickname":      session.Nickname,
			"CustomerName":  session.CustomerName,
			"Data":          session.Date,
			"DataFormatada": formatedDate.Format("02/01/2006"),
			"Duracao":       session.Minutes,
			"Status":        session.Status,
			"StatusCor":     statusColors[session.Status],
			"ServiceID":     session.ServiceID,
			"Servico":       session.Service,
			"Comentario":    session.Comments,
			"Reconciliada":  reconciled,
			"Elegivel":      elegivel,
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
