package http

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"math"

	"planner/internal/port"
	"planner/internal/service"
)

type Sessao struct {
	ID            int
	Nickname      string
	Data          string
	DataFormatada string
	Duracao       int
	Status        string
	StatusCor     string
	Comentario    string
}

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
	mu      sync.Mutex
	sessoes = []Sessao{
		{1, "AnaGamer", "2026-06-01", "01/06/2026", 60, "realizada", "#10b981", "Boa partida.\nEvoluiu bastante no posicionamento tático e rotações."},
		{2, "CarlosPro", "2026-06-02", "02/06/2026", 30, "cancelada/cobrar", "#f59e0b", "Desistiu em cima da hora do compromisso agendado."},
		{3, "BiaPlayer", "2026-06-15", "15/06/2026", 45, "cancelada/não cobrar", "#ef4444", "Teve queda generalizada de energia na região onde mora."},
		{4, "JohnDoe", "2026-06-16", "16/06/2026", 90, "realizada", "#10b981", "Focado nos objetivos estabelecidos no treinamento."},
		{5, "PlayerX", "2026-06-17", "17/06/2026", 60, "realizada", "#10b981", "Treino de mira eficiente com evolução constante."},
		{6, "GamerPro99", "2026-06-18", "18/06/2026", 120, "realizada", "#10b981", "Análise de replay detalhada em grupo."},
		{7, "LucasT", "2026-06-19", "19/06/2026", 45, "cancelada/cobrar", "#f59e0b", "Esqueceu do compromisso e não respondeu avisos."},
	}
	proximoID   = 8
	itensPorPag = 5
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger   port.Logger
	repo     port.Repository
	tmpl     *template.Template
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
	h.writeResponse(w, response)
}

// Sessions handler for the /sessions endpoint
func (h *HandlerHtml) Sessions(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	mu.Lock()
	dados := sessoes
	mu.Unlock()
	h.tmpl.ExecuteTemplate(w, "index", h.paginarSessoes(dados, 1))
}

// writeResponse writes the given response to the http.ResponseWriter with the appropriate status
func (h *HandlerHtml) writeResponse(w http.ResponseWriter, response port.OutDTO) {
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(response.GetStatusCode()))
	w.Write(responseJSON)
}

// paginarSessoes paginates the given sessions based on the target page number
// temporary
func (h *HandlerHtml) paginarSessoes(filtradas []Sessao, paginaAlvo int) PaginacaoData {
	totalItens := len(filtradas)
	if totalItens == 0 {
		return PaginacaoData{Sessoes: []Sessao{}, PaginaAtual: 1, TotalPaginas: 1}
	}
	totalPaginas := int(math.Ceil(float64(totalItens) / float64(itensPorPag)))
	if paginaAlvo < 1 {
		paginaAlvo = 1
	}
	if paginaAlvo > totalPaginas {
		paginaAlvo = totalPaginas
	}
	inicio := (paginaAlvo - 1) * itensPorPag
	fim := inicio + itensPorPag
	if fim > totalItens {
		fim = totalItens
	}
	return PaginacaoData{
		Sessoes: filtradas[inicio:fim], PaginaAtual: paginaAlvo, TotalPaginas: totalPaginas,
		TemAnterior: paginaAlvo > 1, TemProximo: paginaAlvo < totalPaginas,
		PagAnterior: paginaAlvo - 1, PagProxima: paginaAlvo + 1,
	}
}