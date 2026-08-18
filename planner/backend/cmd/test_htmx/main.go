package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// --- DTOs ATUALIZADOS ---

type SessionListRequest struct {
	Page               int    `json:"page" validate:"required,gt=0"`
	PageSize           int    `json:"page_size" validate:"required,gt=0"`
	Nickname           string `json:"nickname,omitempty"`
	DateStart          string `json:"date_start,omitempty"`
	DateEnd            string `json:"date_end,omitempty"`
	Minutes            int    `json:"minutes,omitempty"`
	Realizada          bool   `json:"realizada,omitempty"`
	CanceladaCobrar    bool   `json:"cancelada_cobrar,omitempty"`
	CanceladaNaoCobrar bool   `json:"cancelada_nao_cobrar,omitempty"`
	Comments           string `json:"comments,omitempty"`
}

type SessionCreateRequest struct {
	Nickname           string `json:"nickname" validate:"required"`
	Date               string `json:"date" validate:"required"`
	Minutes            int    `json:"minutes" validate:"required"`
	Realizada          bool   `json:"realizada"`
	CanceladaCobrar    bool   `json:"cancelada_cobrar"`
	CanceladaNaoCobrar bool   `json:"cancelada_nao_cobrar"`
	Comments           string `json:"comments,omitempty"`
}

type SessionDeleteRequest struct {
	SessionID int64 `json:"id" validate:"required"`
}

// --- ENTIDADE DO MODELO ---

type Session struct {
	ID                 int64
	Nickname           string
	Date               string
	Minutes            int
	Realizada          bool
	CanceladaCobrar    bool
	CanceladaNaoCobrar bool
	Comments           string
}

var (
	sessions     = []Session{}
	nextID       int64 = 1
	sessionsMu   sync.Mutex
	templatesEnv *template.Template
)

func init() {
	var err error
	templatesEnv, err = template.ParseFiles("templates/index.html", "templates/partials.html")
	if err != nil {
		log.Fatal("Erro crítico ao compilar os arquivos de template HTML:", err)
	}

	// Carga inicial de testes mapeando os novos estados booleanos
	sessions = append(sessions, Session{ID: nextID, Nickname: "ArthurPendragon", Date: "2026-08-19", Minutes: 60, Realizada: true, Comments: "Mentoria de arquitetura concluída."})
	nextID++
	sessions = append(sessions, Session{ID: nextID, Nickname: "Beatrice_Dev", Date: "2026-08-15", Minutes: 90, CanceladaCobrar: true, Comments: "Cliente faltou em cima da hora."})
	nextID++
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/sessions", handleCreate)
	http.HandleFunc("/sessions/search", handleListAndFilter)
	http.HandleFunc("/sessions/delete", handleDelete)

	log.Println("🚀 Servidor com Status Booleanos rodando em http://localhost:8085")
	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal(err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sessionsMu.Lock()
	data := map[string]interface{}{"Sessions": sessions}
	sessionsMu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templatesEnv.ExecuteTemplate(w, "index.html", data)
}

// POST /sessions -> Criação traduzindo a String do Select em Booleanos estruturais
func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	minutesVal, _ := strconv.Atoi(r.FormValue("minutes"))
	statusForm := r.FormValue("status_selection")

	// Mapeia a seleção única do HTML para o DTO focado em múltiplos booleanos
	createDTO := SessionCreateRequest{
		Nickname:           r.FormValue("nickname"),
		Date:               r.FormValue("date"),
		Minutes:            minutesVal,
		Realizada:          statusForm == "realizada",
		CanceladaCobrar:    statusForm == "cancelada_cobrar",
		CanceladaNaoCobrar: statusForm == "cancelada_nao_cobrar",
		Comments:           r.FormValue("comments"),
	}

	sessionsMu.Lock()
	newSession := Session{
		ID:                 nextID,
		Nickname:           createDTO.Nickname,
		Date:               createDTO.Date,
		Minutes:            createDTO.Minutes,
		Realizada:          createDTO.Realizada,
		CanceladaCobrar:    createDTO.CanceladaCobrar,
		CanceladaNaoCobrar: createDTO.CanceladaNaoCobrar,
		Comments:           createDTO.Comments,
	}
	sessions = append(sessions, newSession)
	nextID++
	
	data := map[string]interface{}{"Sessions": sessions}
	sessionsMu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templatesEnv.ExecuteTemplate(w, "table-rows", data)
}

// GET /sessions/search -> Filtro aplicando os critérios booleanos do SessionListRequest
func handleListAndFilter(w http.ResponseWriter, r *http.Request) {
	minutesQuery, _ := strconv.Atoi(r.URL.Query().Get("minutes"))
	statusFilter := r.URL.Query().Get("status_selection")
	
	reqDTO := SessionListRequest{
		Page:               1,
		PageSize:           10,
		Nickname:           r.URL.Query().Get("nickname"),
		Minutes:            minutesQuery,
		DateStart:          r.URL.Query().Get("date_start"),
		DateEnd:            r.URL.Query().Get("date_end"),
		Realizada:          statusFilter == "realizada",
		CanceladaCobrar:    statusFilter == "cancelada_cobrar",
		CanceladaNaoCobrar: statusFilter == "cancelada_nao_cobrar",
	}

	sessionsMu.Lock()
	filteredSessions := []Session{}

	for _, s := range sessions {
		if reqDTO.Nickname != "" && !strings.Contains(strings.ToLower(s.Nickname), strings.ToLower(reqDTO.Nickname)) {
			continue
		}
		if reqDTO.Minutes > 0 && s.Minutes != reqDTO.Minutes {
			continue
		}
		if reqDTO.DateStart != "" && s.Date < reqDTO.DateStart {
			continue
		}
		if reqDTO.DateEnd != "" && s.Date > reqDTO.DateEnd {
			continue
		}
		
		// Se um filtro de status específico foi selecionado, valida o booleano correspondente
		if statusFilter != "" {
			if statusFilter == "realizada" && !s.Realizada {
				continue
			}
			if statusFilter == "cancelada_cobrar" && !s.CanceladaCobrar {
				continue
			}
			if statusFilter == "cancelada_nao_cobrar" && !s.CanceladaNaoCobrar {
				continue
			}
		}

		filteredSessions = append(filteredSessions, s)
	}

	data := map[string]interface{}{"Sessions": filteredSessions}
	sessionsMu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templatesEnv.ExecuteTemplate(w, "table-rows", data)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	idParsed, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	deleteDTO := SessionDeleteRequest{SessionID: idParsed}

	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	foundIndex := -1
	for idx, s := range sessions {
		if s.ID == deleteDTO.SessionID {
			foundIndex = idx
			break
		}
	}

	if foundIndex != -1 {
		sessions = append(sessions[:foundIndex], sessions[foundIndex+1:]...)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Error(w, "Sessão não encontrada", http.StatusNotFound)
}
