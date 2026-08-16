package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Aula representa o modelo de dados de uma aula de canto
type Aula struct {
	ID        int
	Aluno     string
	DataHora  string
	Estilo    string
	Descricao string
}

var db *sql.DB
var templates *template.Template

func main() {
	var err error
	// Inicializa o banco de dados SQLite (em memória para o exemplo)
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Criação da tabela
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS aulas (id INTEGER PRIMARY KEY AUTOINCREMENT, aluno TEXT, data_hora TEXT, estilo TEXT, descricao TEXT)")
	statement.Exec()

	// Insere alguns dados iniciais
	db.Exec("INSERT INTO aulas (aluno, data_hora, estilo, descricao) VALUES ('Maria Silva', '2026-08-20T14:00', 'Lírico', 'Treino de respiração e extensão vocal.')")
	db.Exec("INSERT INTO aulas (aluno, data_hora, estilo, descricao) VALUES ('João Souza', '2026-08-21T16:30', 'Pop/Rock', 'Afinação, drives básicos e presença de palco.')")

	// Carrega os templates HTML
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Rotas da aplicação
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/aulas", handleAulas)
	http.HandleFunc("/aulas/novo", handleNovaAulaForm)
	http.HandleFunc("/aulas/salvar", handleSalvarAula)
	http.HandleFunc("/aulas/editar", handleEditarAulaForm)
	http.HandleFunc("/aulas/deletar", handleDeletarAula)

	log.Println("Servidor iniciado em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Rota principal: Renderiza a casca da aplicação (SPA Feeling)
func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templates.ExecuteTemplate(w, "index.html", nil)
}

// Retorna a lista de aulas (Fragmento HTML ou Página completa dependendo do HX-Request)
func handleAulas(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, aluno, data_hora, estilo, descricao FROM aulas ORDER BY data_hora ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var aulas []Aula
	for rows.Next() {
		var a Aula
		rows.Scan(&a.ID, &a.Aluno, &a.DataHora, &a.Estilo, &a.Descricao)
		aulas = append(aulas, a)
	}

	templates.ExecuteTemplate(w, "lista.html", aulas)
}

// Retorna o formulário de cadastro limpo
func handleNovaAulaForm(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "formulario.html", Aula{})
}

// Cria ou Atualiza uma aula (POST)
func handleSalvarAula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.FormValue("id")
	aluno := r.FormValue("aluno")
	dataHora := r.FormValue("data_hora")
	estilo := r.FormValue("estilo")
	descricao := r.FormValue("descricao")

	if idStr == "" || idStr == "0" {
		// Inserir novo registro
		stmt, _ := db.Prepare("INSERT INTO aulas (aluno, data_hora, estilo, descricao) VALUES (?, ?, ?, ?)")
		_, err := stmt.Exec(aluno, dataHora, estilo, descricao)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Atualizar registro existente
		id, _ := strconv.Atoi(idStr)
		stmt, _ := db.Prepare("UPDATE aulas SET aluno=?, data_hora=?, estilo=?, descricao=? WHERE id=?")
		_, err := stmt.Exec(aluno, dataHora, estilo, descricao, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Força o HTMX a recarregar a lista de aulas atualizada
	w.Header().Set("HX-Trigger", "atualizarLista")
	w.Write([]byte("<div class='alert alert-success alert-sm mb-3'>Aula salva com sucesso!</div>"))
}

// Retorna o formulário preenchido para edição
func handleEditarAulaForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	var a Aula
	err := db.QueryRow("SELECT id, aluno, data_hora, estilo, descricao FROM aulas WHERE id = ?", id).Scan(&a.ID, &a.Aluno, &a.DataHora, &a.Estilo, &a.Descricao)
	if err != nil {
		http.Error(w, "Aula não encontrada", http.StatusNotFound)
		return
	}

	templates.ExecuteTemplate(w, "formulario.html", a)
}

// Deleta uma aula (DELETE)
func handleDeletarAula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	stmt, _ := db.Prepare("DELETE FROM aulas WHERE id = ?")
	_, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorna vazio ou mensagem rápida (o elemento HTML original será removido pelo hx-target correspondente se configurado,
	// ou podemos apenas disparar o evento de recarregar a lista)
	w.Header().Set("HX-Trigger", "atualizarLista")
	w.Write([]byte(""))
}
