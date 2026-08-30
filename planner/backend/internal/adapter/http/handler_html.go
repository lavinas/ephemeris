package http

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"planner/internal/dto"
	"planner/internal/port"
	"planner/internal/service"
)

// Sessao represents a session with its details for rendering in HTML templates.
type Sessao struct {
	ID            int64
	Nickname      string
	Data          string
	DataFormatada string
	Duracao       int
	Status        string
	StatusCor     string
	Servico       string
	Comentario    string
}

// PaginacaoData holds the paginated session data and pagination details for rendering in HTML templates.
type PaginacaoData struct {
	Sessoes      []Sessao
	PaginaAtual  int
	TotalPaginas int
	TemAnterior  bool
	TemProximo   bool
	PagAnterior  int
	PagProxima   int
}

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
	logger port.Logger
	repo   port.Repository
	tmpl   *template.Template
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, tdata []byte) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
	}
	tmpl, err := template.New("index").Funcs(funcMap).Parse(string(tdata))
	if err != nil {
		return nil, err
	}
	return &HandlerHtml{
		repo:   repo,
		logger: logger,
		tmpl:   tmpl,
	}, nil
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
	err := h.tmpl.ExecuteTemplate(w, "index", page)
	if err != nil {
		h.logger.IPrintf(2, "Failed to render template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// SessionsCreate handler for the /sessions/create endpoint
func (h *HandlerHtml) SessionsCreate(w http.ResponseWriter, r *http.Request) {
	hoje := time.Now().Format("2006-01-02")
	twoYearsAgo := time.Now().AddDate(-2, 0, 0).Format("2006-01-02")
	svc := service.NewSessionUsers(h.repo, h.logger)
	req := &dto.SessionUsersRequest{
		StartDate: twoYearsAgo,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve session users", http.StatusInternalServerError)
		return
	}
	response, ok := respdro.(*dto.SessionUsersResponse)
	if !ok {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	dadosPadrao := map[string]interface{}{
		"DataFormatada": hoje,
		"DataPadrao":    hoje,
		"DuracaoPadrao": "60",
		"Nicknames":     response.Nicknames,
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
}

// SessionsSave handler for the /sessions/save endpoint
func (h *HandlerHtml) SessionsSave(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	minutes, _ := strconv.Atoi(r.FormValue("add_duracao"))
	nickname := r.FormValue("add_nickname")
	status := r.FormValue("add_status")
	form_service := r.FormValue("add_servico")
	comment := r.FormValue("add_comentario")
	svc := service.NewSessionCreate(h.repo, h.logger)
	req := &dto.SessionCreateRequest{
		Nickname: nickname,
		Date:     r.FormValue("add_data"),
		Minutes:  minutes,
		Service:  form_service,
		Status:   status,
		Comments: comment,
	}
	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve sessions", http.StatusInternalServerError)
		return
	}
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	svc2 := service.NewSessionList(h.repo, h.logger)
	req2 := &dto.SessionListRequest{
		Page:     pagina,
		PageSize: itensPorPag,
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
	h.logger.IPrintf(2, "Rendering 2 for %v", page)
	w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela", page)

}

// SessionTableReset handler for the /sessions/table/reset endpoint
func (h *HandlerHtml) SessionsTableReset(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	svc := service.NewSessionList(h.repo, h.logger)
	req := &dto.SessionListRequest{
		Page:     pagina,
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

// SessionsBlockReset handler for the /sessions/block/reset endpoint
func (h *HandlerHtml) SessionsBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// SessionDeleteWarning handler for the /sessions/delete/warning endpoint
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

// SessionsDelete handler for the /sessions/delete endpoint
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
	svc2 := service.NewSessionList(h.repo, h.logger)
	req2 := &dto.SessionListRequest{
		Page:     pagina,
		PageSize: itensPorPag,
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

// SessionsCancelEdition handler for the /sessions/cancel/edition endpoint
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

// getSessionData retrieves the session data based on the provided filters and pagination parameters
func (h *HandlerHtml) getSessionData(response *dto.SessionListResponse) []Sessao {
	sessions := response.Sessions
	pageSessoes := []Sessao{}
	for _, session := range sessions {
		formatedDate, _ := time.Parse("2006-01-02", session.Date)
		ss := Sessao{
			ID:            session.SessionID,
			Nickname:      session.Nickname,
			Data:          session.Date,
			DataFormatada: formatedDate.Format("02/01/2006"),
			Duracao:       session.Minutes,
			Status:        session.Status,
			StatusCor:     statusColors[session.Status],
			Servico:       session.Service,
			Comentario:    session.Comments,
		}
		pageSessoes = append(pageSessoes, ss)
	}
	return pageSessoes
}

// getPageData paginates the filtered sessions based on the target page and items per page
func (h *HandlerHtml) getPageData(response *dto.SessionListResponse) PaginacaoData {
	page := response.Page
	totalPages := response.TotalPages
	pageSessoes := h.getSessionData(response)
	ret := PaginacaoData{
		Sessoes:      pageSessoes,
		PaginaAtual:  page,
		TotalPaginas: totalPages,
		TemAnterior:  page > 1,
		TemProximo:   page < totalPages,
		PagAnterior:  page - 1,
		PagProxima:   page + 1,
	}
	return ret
}
