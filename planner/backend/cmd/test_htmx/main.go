package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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

var tmpl *template.Template

func init() {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
	}
	htmlTemplates, err := os.ReadFile("templates/sessions.html")
	if err != nil {
		panic(err)
	}
	tmpl = template.Must(template.New("master").Funcs(funcMap).Parse(string(htmlTemplates)))
}

func paginarSessoes(filtradas []Sessao, paginaAlvo int) PaginacaoData {
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

func filtrarSessoes(nickname, dataInicio, dataFim, duracaoFiltro, status, comentario string) []Sessao {
	mu.Lock()
	defer mu.Unlock()
	var resultado []Sessao
	for _, s := range sessoes {
		if nickname != "" && !strings.Contains(strings.ToLower(s.Nickname), strings.ToLower(nickname)) {
			continue
		}
		if dataInicio != "" && s.Data < dataInicio {
			continue
		}
		if dataFim != "" && s.Data > dataFim {
			continue
		}
		if duracaoFiltro != "" {
			dVal, err := strconv.Atoi(duracaoFiltro)
			if err == nil && s.Duracao != dVal {
				continue
			}
		}
		if status != "" && !strings.Contains(strings.ToLower(s.Status), strings.ToLower(status)) {
			continue
		}
		if comentario != "" && !strings.Contains(strings.ToLower(s.Comentario), strings.ToLower(comentario)) {
			continue
		}
		resultado = append(resultado, s)
	}
	return resultado
}

func configurarRegistro(nickname, data, status, comentario string, duracao int) Sessao {
	partes := strings.Split(data, "-")
	dForm := data
	if len(partes) == 3 {
		dForm = fmt.Sprintf("%s/%s/%s", partes[0], partes[1], partes[2])
	}
	cor := "#9ca3af"
	stLimpo := strings.ToLower(strings.TrimSpace(status))
	if stLimpo == "realizada" {
		cor = "#10b981"
	}
	if stLimpo == "cancelada/cobrar" {
		cor = "#f59e0b"
	}
	if stLimpo == "cancelada/não cobrar" {
		cor = "#ef4444"
	}

	return Sessao{
		Nickname: nickname, Data: data, DataFormatada: dForm,
		Duracao: duracao, Status: status, StatusCor: cor, Comentario: comentario,
	}
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		dados := sessoes
		mu.Unlock()
		tmpl.ExecuteTemplate(w, "index", paginarSessoes(dados, 1))
	})

	http.HandleFunc("/tabela", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		pagStr := r.URL.Query().Get("page")
		if pagStr == "" {
			pagStr = r.FormValue("page")
		}
		pagina, _ := strconv.Atoi(pagStr)
		if pagina < 1 {
			pagina = 1
		}

		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	http.HandleFunc("/tabela/reset", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		dados := sessoes
		mu.Unlock()
		w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(dados, 1))
	})

	// ATUALIZADO: Passa um mapa contendo a data e a duração padrão para o HTML
	http.HandleFunc("/sessoes/novo", func(w http.ResponseWriter, r *http.Request) {
		hoje := time.Now().Format("2006-01-02")
		dadosPadrao := map[string]string{
			"DataFormatada": hoje,
			"DataPadrao":    hoje,
			"DuracaoPadrao": "60",
		}
		tmpl.ExecuteTemplate(w, "formulario_cadastro", dadosPadrao)
	})

	http.HandleFunc("/limpar-bloco", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("")) })

	http.HandleFunc("/sessoes/editar", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		mu.Lock()
		var sSel Sessao
		for _, s := range sessoes {
			if s.ID == id {
				sSel = s
				break
			}
		}
		mu.Unlock()
		tmpl.ExecuteTemplate(w, "linha_sessao_edit", sSel)
	})

	http.HandleFunc("/sessoes/deletar-aviso", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		mu.Lock()
		var sSel Sessao
		for _, s := range sessoes {
			if s.ID == id {
				sSel = s
				break
			}
		}
		mu.Unlock()
		tmpl.ExecuteTemplate(w, "linha_sessao_deletar_aviso", sSel)
	})

	http.HandleFunc("/sessoes/cancelar-edicao", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		mu.Lock()
		var sSel Sessao
		for _, s := range sessoes {
			if s.ID == id {
				sSel = s
				break
			}
		}
		mu.Unlock()
		tmpl.ExecuteTemplate(w, "linha_sessao", sSel)
	})

	http.HandleFunc("/sessoes/atualizar", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		r.ParseForm()
		duracao, _ := strconv.Atoi(r.FormValue("edit_duracao"))
		regConfigurado := configurarRegistro(r.FormValue("edit_nickname"), r.FormValue("edit_data"), r.FormValue("edit_status"), r.FormValue("edit_comentario"), duracao)
		mu.Lock()
		for i, s := range sessoes {
			if s.ID == id {
				sessoes[i] = regConfigurado
				sessoes[i].ID = id
				break
			}
		}
		mu.Unlock()
		pagina, _ := strconv.Atoi(r.FormValue("page"))
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	http.HandleFunc("/sessoes/salvar", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		duracao, _ := strconv.Atoi(r.FormValue("add_duracao"))
		novaSessao := configurarRegistro(r.FormValue("add_nickname"), r.FormValue("add_data"), r.FormValue("add_status"), r.FormValue("add_comentario"), duracao)
		mu.Lock()
		novaSessao.ID = proximoID
		sessoes = append(sessoes, novaSessao)
		proximoID++
		mu.Unlock()
		pagina, _ := strconv.Atoi(r.FormValue("page"))
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("comentario"))
		w.Write([]byte(`<script>document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	http.HandleFunc("/sessoes/deletar", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(r.URL.Query().Get("id"))
		mu.Lock()
		idx := -1
		for i, s := range sessoes {
			if s.ID == id {
				idx = i
				break
			}
		}
		if idx != -1 {
			sessoes = append(sessoes[:idx], sessoes[idx+1:]...)
		}
		mu.Unlock()
		r.ParseForm()
		pagina, _ := strconv.Atoi(r.FormValue("page"))
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	fmt.Println("Dashboard Escuro Corrigido rodando em http://localhost:8085")
	http.ListenAndServe(":8085", nil)
}

func NavFilter(nickname, dI, dF, status, coment string) []Sessao {
	return filtrarSessoes(nickname, dI, dF, "", status, coment)
}
