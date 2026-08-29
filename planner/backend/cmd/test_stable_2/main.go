package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
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
	Servico       string
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
		{1, "AnaGamer", "2026-06-01", "01/06/2026", 60, "realizada", "#10b981", "aula/canto", "Boa partida.\nEvoluiu bastante no posicionamento tático e rotações."},
		{2, "CarlosPro", "2026-06-02", "02/06/2026", 30, "cancelada/cobrar", "#f59e0b", "aula/piano", "Desistiu em cima da hora do compromisso agendado."},
		{3, "BiaPlayer", "2026-06-15", "15/06/2026", 45, "cancelada/não cobrar", "#ef4444", "aula/canto", "Teve queda generalizada de energia na região onde mora."},
		{4, "JohnDoe", "2026-06-16", "16/06/2026", 90, "realizada", "#10b981", "aula/piano", "Focado nos objetivos estabelecidos no treinamento."},
		{5, "PlayerX", "2026-06-17", "17/06/2026", 60, "realizada", "#10b981", "aula/canto", "Treino de mira eficiente com evolução constante."},
		{6, "GamerPro99", "2026-06-18", "18/06/2026", 120, "realizada", "#10b981", "aula/piano", "Análise de replay detalhada em grupo."},
		{7, "LucasT", "2026-06-19", "19/06/2026", 45, "cancelada/cobrar", "#f59e0b", "aula/canto", "Esqueceu do compromisso e não respondeu avisos."},
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
	// Inicialização montada aqui no início para blindar o escopo das variáveis
	tmpl = template.Must(template.New("master").Funcs(funcMap).Parse(htmlTemplates))
}

var htmlTemplates = `
	{{define "index"}}
	<!DOCTYPE html>
	<html lang="pt-BR">
	<head>
		<meta charset="UTF-8">
		<title>Dashboard de Sessões</title>
		<script src="https://tailwindcss.com"></script>
		<script src="/static/htmx.min.js"></script>
		<style>
			.grade-filtros {
				display: grid !important;
				grid-template-columns: 1.5fr 1fr 1fr 1fr 1.3fr 1.3fr 2fr !important;
				gap: 10px !important;
				width: 100% !important;
				box-sizing: border-box !important;
			}
			.grade-tabela {
				display: grid !important;
				grid-template-columns: 2fr 1.2fr 0.8fr 1.2fr 1.4fr 4.7fr 1.2fr !important;
				gap: 16px !important;
				align-items: center !important;
				width: 100% !important;
				box-sizing: border-box !important;
				padding: 16px 24px !important;
			}
			.topo-tabela {
				color: #9ca3af !important;
				font-size: 11px !important;
				font-weight: 700 !important;
				text-transform: uppercase !important;
				letter-spacing: 0.05em !important;
				border-bottom: 1px solid #374151 !important;
				padding-bottom: 12px !important;
				margin-bottom: 12px !important;
			}
			.linha-box {
				background-color: #f3f4f6 !important;
				border: 1px solid #e5e7eb !important;
				border-radius: 8px !important;
				margin-bottom: 8px !important;
				color: #1f2937 !important;
				transition: background-color 0.15s ease !important;
				box-shadow: 0 1px 3px rgba(0,0,0,0.1) !important;
			}
			.linha-box.formulario-add, .linha-box.formulario-edit, .linha-box.aviso-deletar {
				background-color: #ffffff !important;
				border: 1px solid #4f46e5 !important;
				color: #1f2937 !important;
			}
			.linha-box.aviso-deletar { border-color: #e11d48 !important; }
			.linha-box:hover:not(.formulario-add):not(.formulario-edit):not(.aviso-deletar) {
				background-color: #e5e7eb !important;
			}
			.campo-box {
				background-color: #1f2937 !important;
				border: 1px solid #374151 !important;
				padding: 12px !important;
				border-radius: 8px !important;
				display: flex !important;
				flex-direction: column !important;
			}
			.campo-box label {
				display: block !important;
				font-size: 11px !important;
				font-weight: 600 !important;
				color: #9ca3af !important;
				text-transform: uppercase !important;
				margin-bottom: 6px !important;
				letter-spacing: 0.05em !important;
			}
			.campo-box input, .campo-box select {
				width: 100% !important;
				padding: 8px 12px !important;
				background-color: #111827 !important;
				border: 1px solid #4b5563 !important;
				color: #f3f4f6 !important;
				border-radius: 6px !important;
				font-size: 14px !important;
				outline: none !important;
				box-sizing: border-box !important;
			}
			.input-inline-tabela {
				width: 100% !important;
				padding: 6px 10px !important;
				background-color: #f9fafb !important;
				border: 1px solid #d1d5db !important;
				color: #1f2937 !important;
				border-radius: 6px !important;
				font-size: 14px !important;
				outline: none !important;
				box-sizing: border-box !important;
			}
			.campo-box input:focus, .campo-box select:focus { border-color: #6366f1 !important; }
			.input-inline-tabela:focus { border-color: #4f46e5 !important; background-color: #ffffff !important; }
			.texto-truncado { white-space: nowrap !important; overflow: hidden !important; text-overflow: ellipsis !important; }
			.comentario-multilinha { white-space: normal !important; word-break: break-word !important; line-height: 1.4 !important; }
		</style>
	</head>
	<body class="bg-gray-950 text-gray-100 font-sans antialiased min-h-screen py-10 px-4">
		<div class="max-w-7xl mx-auto bg-gray-900 shadow-2xl rounded-2xl overflow-hidden border border-gray-800 p-8">
			
			<div style="width: 100% !important; text-align: center !important; margin-bottom: 35px !important; display: block !important;">
				<img src="/static/logo.png" alt="Amélia Cardoso Logo" style="height: 130px !important; max-height: 130px !important; margin: 0 auto !important; display: inline-block !important; object-fit: contain !important;" onError="this.style.display='none';">
			</div>

			<form id="filtro-form" hx-post="/tabela" hx-target="#tabela-container" hx-trigger="input delay:200ms, change" class="w-full">
				<input type="hidden" id="input-pagina-form" name="page" value="1">
				<div class="grade-filtros">
					<div class="campo-box">
						<label>Nickname</label>
						<input type="text" name="nickname" placeholder="Filtrar usuário...">
					</div>
					<div class="campo-box">
						<label>Data Início</label>
						<input type="date" name="data_inicio">
					</div>
					<div class="campo-box">
						<label>Data Fim</label>
						<input type="date" name="data_fim">
					</div>
					<div class="campo-box">
						<label>Duração</label>
						<input type="number" name="duracao_filtro" placeholder="Minutos..." min="0">
					</div>
					<div class="campo-box">
						<label>Status</label>
						<select name="status">
							<option value="">Todos</option>
							<option value="realizada">Realizada</option>
							<option value="cancelada/cobrar">Cancelada/Cobrar</option>
							<option value="cancelada/não cobrar">Cancelada/Não Cobrar</option>
						</select>
					</div>
					<div class="campo-box">
						<label>Serviço</label>
						<select name="servico_filtro">
							<option value="">Todos</option>
							<option value="aula/canto">Aula/Canto</option>
							<option value="aula/piano">Aula/Piano</option>
						</select>
					</div>
					<div class="campo-box">
						<label>Buscar nos Comentários</label>
						<input type="text" name="comentario" placeholder="Palavra-chave...">
					</div>
				</div>
			</form>
			<div style="height: 40px; width: 100%;"></div>
			<div class="flex justify-start gap-4">
				<button hx-get="/sessoes/novo" hx-target="#formulario-cadastro-container" hx-swap="innerHTML" class="bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-sm px-6 py-2.5 rounded-lg border border-indigo-500 transition duration-150 shadow-md">＋ Adicionar</button>
				<button hx-post="/tabela/reset" hx-target="#tabela-container" class="bg-gray-800 hover:bg-gray-700 text-gray-300 font-bold text-sm px-6 py-2.5 rounded-lg border border-gray-700 transition duration-150 shadow-md">🧹 Limpar Filtros</button>
			</div>
			<div id="formulario-cadastro-container" class="mt-4"></div>
			<div style="height: 50px; width: 100%;"></div>
			<div id="tabela-container" class="w-full">
				{{template "tabela" .}}
			</div>
		</div>
	</body>
	</html>
	{{end}}

	{{define "formulario_cadastro"}}
	<form hx-post="/sessoes/salvar" hx-target="#tabela-container" hx-include="#filtro-form" class="linha-box grade-tabela formulario-add text-sm shadow-lg">
		<div>
			<input type="text" name="add_nickname" placeholder="Nickname..." list="lista-nicknames" required class="input-inline-tabela">
			<datalist id="lista-nicknames">
				<option value="AnaGamer"><option value="CarlosPro"><option value="BiaPlayer"><option value="JohnDoe"><option value="PlayerX"><option value="GamerPro99"><option value="LucasT">
			</datalist>
		</div>
		<div><input type="date" name="add_data" value="{{.DataPadrao}}" required class="input-inline-tabela"></div>
		<div><input type="number" name="add_duracao" value="{{.DuracaoPadrao}}" min="1" required class="input-inline-tabela" style="text-align: center !important;"></div>
		<div>
			<select name="add_status" required class="input-inline-tabela">
				<option value="realizada">Realizada</option>
				<option value="cancelada/cobrar">Cancelada/Cobrar</option>
				<option value="cancelada/não cobrar">Cancelada/Não Cobrar</option>
			</select>
		</div>
		<div>
			<select name="add_servico" required class="input-inline-tabela">
				<option value="aula/canto">Aula/Canto</option>
				<option value="aula/piano">Aula/Piano</option>
			</select>
		</div>
		<div><textarea name="add_comentario" placeholder="Escreva um comentário..." rows="2" class="input-inline-tabela" style="resize: vertical;"></textarea></div>
		<div class="flex gap-2 justify-start w-full" style="justify-content: flex-start !important; align-items: center !important;">
			<button type="submit" class="bg-green-600 hover:bg-green-700 text-white font-bold text-xs px-3 py-2 rounded-md transition w-16 text-center">Salvar</button>
			<button type="button" hx-get="/limpar-bloco" hx-target="#formulario-cadastro-container" class="bg-gray-200 hover:bg-gray-300 text-gray-700 font-bold text-xs px-3 py-2 rounded-md border border-gray-300 transition w-16 text-center">Cancelar</button>
		</div>
	</form>
	{{end}}

	{{define "tabela"}}
	<div class="w-full">
		<div class="grade-tabela topo-tabela">
			<div>Usuário</div>
			<div>Data</div>
			<div style="text-align: center !important;">Duração</div>
			<div>Status</div>
			<div>Serviço</div>
			<div>Comentários</div>
			<div style="text-align: left !important; width: 100%;">Ações</div>
		</div>
		{{if .Sessoes}}
			{{range .Sessoes}}
				{{template "linha_sessao" .}}
			{{end}}
			<div style="height: 40px; width: 100%;"></div>
			<div class="mt-12 flex items-center justify-between bg-gray-800/40 p-4 rounded-xl border border-gray-800 text-sm">
				<div class="text-gray-400">Página <span class="text-gray-200 font-bold font-mono">{{.PaginaAtual}}</span> de <span class="text-gray-200 font-bold font-mono">{{.TotalPaginas}}</span></div>
				<div class="flex gap-2">
					<button {{if not .TemAnterior}}disabled style="opacity: 0.3; cursor: not-allowed;"{{end}} hx-post="/tabela?page={{.PagAnterior}}" hx-target="#tabela-container" hx-include="#filtro-form" onclick="document.getElementById('input-pagina-form').value='{{.PagAnterior}}'" class="bg-gray-800 hover:bg-gray-700 text-gray-200 font-bold px-4 py-2 rounded-lg border border-gray-700 transition">◀ Anterior</button>
					<button {{if not .TemProximo}}disabled style="opacity: 0.3; cursor: not-allowed;"{{end}} hx-post="/tabela?page={{.PagProxima}}" hx-target="#tabela-container" hx-include="#filtro-form" onclick="document.getElementById('input-pagina-form').value='{{.PagProxima}}'" class="bg-gray-800 hover:bg-gray-700 text-gray-200 font-bold px-4 py-2 rounded-lg border border-gray-700 transition">Próximo ▶</button>
				</div>
			</div>
		{{else}}
			<div class="bg-gray-900/50 border border-gray-800 rounded-xl py-12 text-center text-gray-500 italic">Nenhum registro corresponde aos filtros selecionados.</div>
		{{end}}
	</div>
	{{end}}

	{{define "linha_sessao"}}
	<div id="sessao-{{.ID}}" class="linha-box grade-tabela text-sm" style="align-items: start !important; padding-top: 20px; padding-bottom: 20px;">
		<div class="font-bold text-gray-900 texto-truncado">{{.Nickname}}</div>
		<div class="text-gray-600 font-mono font-medium">{{.DataFormatada}}</div>
		<div class="text-gray-700 font-bold font-mono text-base" style="text-align: center !important;">{{.Duracao}}</div>
		<div><span style="color: {{.StatusCor}} !important;" class="text-xs uppercase tracking-wider font-black">{{.Status}}</span></div>
		<div class="text-gray-700 font-medium italic font-mono">{{.Servico}}</div>
		<div class="text-gray-600 comentario-multilinha">{{pularLinhas .Comentario}}</div>
		<div class="flex gap-2 justify-start w-full" style="justify-content: flex-start !important; align-items: start !important;">
			<button hx-get="/sessoes/editar?id={{.ID}}" hx-target="#sessao-{{.ID}}" hx-swap="outerHTML" class="text-indigo-600 hover:text-white font-bold text-xs px-2.5 py-1.5 bg-indigo-600/10 hover:bg-indigo-600 rounded-md border border-indigo-600/20 transition shadow-sm w-16 text-center">Editar</button>
			<button hx-get="/sessoes/deletar-aviso?id={{.ID}}" hx-target="#sessao-{{.ID}}" hx-swap="outerHTML" class="text-rose-600 hover:text-white font-bold text-xs px-2.5 py-1.5 bg-rose-600/10 hover:bg-rose-600 rounded-md border border-rose-600/20 transition duration-150 shadow-sm w-16 text-center">Deletar</button>
		</div>
	</div>
	{{end}}

	{{define "linha_sessao_edit"}}
	<form id="sessao-{{.ID}}" hx-post="/sessoes/atualizar?id={{.ID}}" hx-target="#tabela-container" hx-include="#filtro-form" class="linha-box grade-tabela formulario-edit text-sm shadow-md" style="align-items: start !important; padding-top: 20px; padding-bottom: 20px;">
		<div><input type="text" name="edit_nickname" value="{{.Nickname}}" required class="input-inline-tabela"></div>
		<div><input type="date" name="edit_data" value="{{.Data}}" required class="input-inline-tabela"></div>
		<div><input type="number" name="edit_duracao" value="{{.Duracao}}" min="1" required class="input-inline-tabela" style="text-align: center !important;"></div>
		<div>
			<select name="edit_status" required class="input-inline-tabela">
				<option value="realizada" {{if eq .Status "realizada"}}selected{{end}}>Realizada</option>
				<option value="cancelada/cobrar" {{if eq .Status "cancelada/cobrar"}}selected{{end}}>Cancelada/Cobrar</option>
				<option value="cancelada/não cobrar" {{if eq .Status "cancelada/não cobrar"}}selected{{end}}>Cancelada/Não Cobrar</option>
			</select>
		</div>
		<div>
			<select name="edit_servico" required class="input-inline-tabela">
				<option value="aula/canto" {{if eq .Servico "aula/canto"}}selected{{end}}>Aula/Canto</option>
				<option value="aula/piano" {{if eq .Servico "aula/piano"}}selected{{end}}>Aula/Piano</option>
			</select>
		</div>
		<div><textarea name="edit_comentario" rows="2" class="input-inline-tabela" style="resize: vertical;">{{.Comentario}}</textarea></div>
		<div class="flex gap-2 justify-start w-full" style="justify-content: flex-start !important; align-items: start !important;">
			<button type="submit" class="bg-green-600 hover:bg-green-700 text-white font-bold text-xs px-2.5 py-1.5 rounded-md transition shadow-sm w-16 text-center">Salvar</button>
			<button type="button" hx-get="/sessoes/cancelar-edicao?id={{.ID}}" hx-target="#sessao-{{.ID}}" hx-swap="outerHTML" class="bg-gray-200 hover:bg-gray-300 text-gray-700 font-bold text-xs px-2.5 py-1.5 rounded-md border border-gray-300 transition shadow-sm w-16 text-center">Cancelar</button>
		</div>
	</form>
	{{end}}

	{{define "linha_sessao_deletar_aviso"}}
	<div id="sessao-{{.ID}}" class="linha-box grade-tabela aviso-deletar text-sm bg-rose-50/50 py-5" style="align-items: center !important;">
		<div class="col-span-6 text-rose-700 font-bold flex items-center gap-2" style="grid-column: span 6 / span 6 !important;">⚠️ Deseja realmente excluir permanentemente a sessão de <span class="underline">{{.Nickname}}</span>?</div>
		<div class="flex gap-2 justify-end w-full" style="justify-content: flex-end !important;">
			<button hx-post="/sessoes/deletar?id={{.ID}}" hx-target="#tabela-container" hx-include="#filtro-form" class="bg-rose-600 hover:bg-rose-700 text-white font-bold text-xs px-3 py-2 rounded-md transition shadow-md">Sim, deletar</button>
			<button hx-get="/sessoes/cancelar-edicao?id={{.ID}}" hx-target="#sessao-{{.ID}}" hx-swap="outerHTML" class="bg-gray-200 hover:bg-gray-300 text-gray-700 font-bold text-xs px-3 py-2 rounded-md border border-gray-300 transition shadow-sm">Não</button>
		</div>
	</div>
	{{end}}
	`

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

func filtrarSessoes(nickname, dataInicio, dataFim, duracaoFiltro, status, servicoFiltro, comentario string) []Sessao {
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
		if servicoFiltro != "" && s.Servico != servicoFiltro {
			continue
		}
		if comentario != "" && !strings.Contains(strings.ToLower(s.Comentario), strings.ToLower(comentario)) {
			continue
		}
		resultado = append(resultado, s)
	}
	return resultado
}

// FUNÇÃO CIRURGICAMENTE CORRIGIDA: Extrai os índices individuais para montar DD/MM/YYYY sem duplicações
func configurarRegistro(nickname, data, status, servico, comentario string, duracao int) Sessao {
	partes := strings.Split(data, "-")
	dForm := data
	if len(partes) == 3 {
		dForm = fmt.Sprintf("%s/%s/%s", partes[2], partes[1], partes[0])
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
		Duracao: duracao, Status: status, StatusCor: cor, Servico: servico, Comentario: comentario,
	}
}

func main() {
	fs := http.FileServer(http.Dir("web/static"))
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

		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("servico_filtro"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	http.HandleFunc("/tabela/reset", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		dados := sessoes
		mu.Unlock()
		w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(dados, 1))
	})

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
		regConfigurado := configurarRegistro(r.FormValue("edit_nickname"), r.FormValue("edit_data"), r.FormValue("edit_status"), r.FormValue("edit_servico"), r.FormValue("edit_comentario"), duracao)
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
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("servico_filtro"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	http.HandleFunc("/sessoes/salvar", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		duracao, _ := strconv.Atoi(r.FormValue("add_duracao"))
		novaSessao := configurarRegistro(r.FormValue("add_nickname"), r.FormValue("add_data"), r.FormValue("add_status"), r.FormValue("add_servico"), r.FormValue("add_comentario"), duracao)
		mu.Lock()
		novaSessao.ID = proximoID
		sessoes = append(sessoes, novaSessao)
		proximoID++
		mu.Unlock()
		pagina, _ := strconv.Atoi(r.FormValue("page"))
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("servico_filtro"), r.FormValue("comentario"))
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
		filtradas := filtrarSessoes(r.FormValue("nickname"), r.FormValue("data_inicio"), r.FormValue("data_fim"), r.FormValue("duracao_filtro"), r.FormValue("status"), r.FormValue("servico_filtro"), r.FormValue("comentario"))
		tmpl.ExecuteTemplate(w, "tabela", paginarSessoes(filtradas, pagina))
	})

	fmt.Println("Dashboard Escuro Atualizado rodando em http://localhost:8085")
	http.ListenAndServe(":8085", nil)
}

func NavFilter(nickname, dI, dF, status, coment string) []Sessao {
	return filtrarSessoes(nickname, dI, dF, "", status, "", coment)
}
