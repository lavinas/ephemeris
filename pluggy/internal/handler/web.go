package handler

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pluggy/internal/config"
	"pluggy/internal/db"
	"pluggy/internal/pluggy"
	"pluggy/internal/service"
)

type WebHandler struct {
	cfg       *config.Config
	db        *db.DB
	client    *pluggy.Client
	syncer    *service.Syncer
	templates *template.Template
}

func NewWebHandler(cfg *config.Config, database *db.DB, client *pluggy.Client, syncer *service.Syncer, tmplDir string) (*WebHandler, error) {
	funcMap := template.FuncMap{
		"formatMoney": func(val any) string {
			if val == nil {
				return "-"
			}
			var f float64
			switch v := val.(type) {
			case float64:
				f = v
			case *float64:
				if v == nil {
					return "-"
				}
				f = *v
			case int:
				f = float64(v)
			default:
				return "-"
			}
			negative := f < 0
			if negative {
				f = -f
			}
			parts := strings.Split(fmt.Sprintf("%.2f", f), ".")
			intPart := parts[0]
			decPart := parts[1]

			var result []string
			for i := len(intPart); i > 0; i -= 3 {
				start := i - 3
				if start < 0 {
					start = 0
				}
				result = append([]string{intPart[start:i]}, result...)
			}
			formatted := strings.Join(result, ".") + "," + decPart
			if negative {
				return "-R$ " + formatted
			}
			return "R$ " + formatted
		},
		"formatMoneyPtr": func(val *float64) string {
			if val == nil {
				return "-"
			}
			f := *val
			negative := f < 0
			if negative {
				f = -f
			}
			parts := strings.Split(fmt.Sprintf("%.2f", f), ".")
			intPart := parts[0]
			decPart := parts[1]

			var result []string
			for i := len(intPart); i > 0; i -= 3 {
				start := i - 3
				if start < 0 {
					start = 0
				}
				result = append([]string{intPart[start:i]}, result...)
			}
			formatted := strings.Join(result, ".") + "," + decPart
			if negative {
				return "-R$ " + formatted
			}
			return "R$ " + formatted
		},
		"formatFloatPtr": func(f *float64, format string) string {
			if f == nil {
				return "-"
			}
			return fmt.Sprintf(format, *f)
		},
		"formatDate": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("02/01/2006")
		},
		"formatDateTime": func(t time.Time) string {
			if t.IsZero() {
				return "-"
			}
			return t.Format("02/01/2006 15:04")
		},
		"formatDatePtr": func(t *time.Time) string {
			if t == nil || t.IsZero() {
				return "-"
			}
			return t.Format("02/01/2006")
		},
		"formatDateTimePtr": func(t *time.Time) string {
			if t == nil || t.IsZero() {
				return "-"
			}
			return t.Format("02/01/2006 15:04")
		},
		"formatRate": func(r *float64, rateType string) string {
			if r == nil {
				return "-"
			}
			if rateType != "" {
				return fmt.Sprintf("%.2f%% %s", *r, rateType)
			}
			return fmt.Sprintf("%.2f%%", *r)
		},
		"statusBadge": func(status string) template.HTML {
			var bg, text string
			switch strings.ToUpper(status) {
			case "UPDATED", "SUCCESS", "ACTIVE", "POSTED":
				bg = "bg-emerald-100"
				text = "text-emerald-800"
			case "UPDATING", "PENDING":
				bg = "bg-amber-100"
				text = "text-amber-800"
			case "LOGIN_ERROR", "ERROR", "OUTDATED":
				bg = "bg-rose-100"
				text = "text-rose-800"
			default:
				bg = "bg-slate-100"
				text = "text-slate-800"
			}
			return template.HTML(fmt.Sprintf(`<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium %s %s">%s</span>`, bg, text, status))
		},
		"isDebit": func(t string, amount float64) bool {
			return strings.ToUpper(t) == "DEBIT" || amount < 0
		},
		"add": func(a, b int) int {
			return a + b
		},
		"subtract": func(a, b int) int {
			return a - b
		},
	}

	tmplPattern := tmplDir + "/*.html"
	tmpl, err := template.New("").Funcs(funcMap).ParseGlob(tmplPattern)
	if err != nil {
		return nil, fmt.Errorf("erro ao compilar templates em %s: %w", tmplPattern, err)
	}

	return &WebHandler{
		cfg:       cfg,
		db:        database,
		client:    client,
		syncer:    syncer,
		templates: tmpl,
	}, nil
}

// RenderDashboard exibe o painel principal
func (w *WebHandler) RenderDashboard(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	summary, err := w.db.GetPortfolioSummary(ctx)
	if err != nil {
		log.Printf("[Web] Erro ao buscar resumo: %v", err)
		summary = &db.PortfolioSummary{}
	}

	recentTxs, _, _ := w.db.ListTransactions(ctx, "", "", "", 8, 0)
	items, _ := w.db.ListItems(ctx)
	recentWebhooks, _ := w.db.ListWebhookEvents(ctx, 5)

	data := map[string]any{
		"Title":          "Dashboard",
		"ActiveTab":      "dashboard",
		"Summary":        summary,
		"RecentTxs":      recentTxs,
		"Items":          items,
		"RecentWebhooks": recentWebhooks,
		"Config":         w.cfg,
	}

	w.render(rw, r, "dashboard.html", data)
}

// RenderStatements exibe os extratos bancários (contas correntes e poupanças)
func (w *WebHandler) RenderStatements(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accountID := r.URL.Query().Get("accountId")
	search := r.URL.Query().Get("search")
	pageStr := r.URL.Query().Get("page")

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	limit := 30
	offset := (page - 1) * limit

	bankAccounts, _ := w.db.ListAccounts(ctx, "BANK")
	txs, total, err := w.db.ListTransactions(ctx, accountID, "BANK", search, limit, offset)
	if err != nil {
		log.Printf("[Web] Erro ao buscar transações bancárias: %v", err)
	}

	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	data := map[string]any{
		"Title":        "Extratos Bancários",
		"ActiveTab":    "statements",
		"Accounts":     bankAccounts,
		"Transactions": txs,
		"AccountID":    accountID,
		"Search":       search,
		"Page":         page,
		"TotalPages":   totalPages,
		"Total":        total,
	}

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "statements-table-container" {
		w.renderFragment(rw, "statements-table", data)
		return
	}

	w.render(rw, r, "statements.html", data)
}

// RenderCreditCards exibe os cartões de crédito, limites, faturas e despesas
func (w *WebHandler) RenderCreditCards(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cards, _ := w.db.ListAccounts(ctx, "CREDIT")

	cardID := r.URL.Query().Get("cardId")
	if cardID == "" && len(cards) > 0 {
		cardID = cards[0].ID
	}

	bills, _ := w.db.ListBills(ctx, cardID)
	txs, _, _ := w.db.ListTransactions(ctx, cardID, "CREDIT", "", 50, 0)

	data := map[string]any{
		"Title":        "Cartões de Crédito",
		"ActiveTab":    "credit_cards",
		"Cards":        cards,
		"SelectedCard": cardID,
		"Bills":        bills,
		"Transactions": txs,
	}

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "card-details-container" {
		w.renderFragment(rw, "card-details", data)
		return
	}

	w.render(rw, r, "credit_cards.html", data)
}

// RenderInvestments exibe o portfólio de investimentos e operações
func (w *WebHandler) RenderInvestments(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	itemID := r.URL.Query().Get("itemId")
	items, _ := w.db.ListItems(ctx)
	investments, _ := w.db.ListInvestments(ctx, itemID)

	var totalBalance, totalAmount float64
	for _, inv := range investments {
		totalBalance += inv.Balance
		totalAmount += inv.Amount
	}

	data := map[string]any{
		"Title":        "Aplicações e Investimentos",
		"ActiveTab":    "investments",
		"Items":        items,
		"SelectedItem": itemID,
		"Investments":  investments,
		"TotalBalance": totalBalance,
		"TotalAmount":  totalAmount,
	}

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "investments-table-container" {
		w.renderFragment(rw, "investments-table", data)
		return
	}

	w.render(rw, r, "investments.html", data)
}

// RenderInvestmentTransactions retorna o modal/fragmento de transações de um ativo
func (w *WebHandler) RenderInvestmentTransactions(rw http.ResponseWriter, r *http.Request) {
	invID := r.URL.Query().Get("id")
	if invID == "" {
		http.Error(rw, "id é obrigatório", http.StatusBadRequest)
		return
	}

	txs, _ := w.db.ListInvestmentTransactions(r.Context(), invID)
	data := map[string]any{
		"Transactions": txs,
		"InvestmentID": invID,
	}
	w.renderFragment(rw, "investment-transactions-modal", data)
}

// RenderConnections exibe as instituições conectadas (Items) e permite gerenciar conexões
func (w *WebHandler) RenderConnections(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, _ := w.db.ListItems(ctx)

	data := map[string]any{
		"Title":     "Conexões Pluggy",
		"ActiveTab": "connections",
		"Items":     items,
	}

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "connections-list-container" {
		w.renderFragment(rw, "connections-list", data)
		return
	}

	w.render(rw, r, "connections.html", data)
}

// RenderWebhooks exibe os logs de webhook e configuração de túnel público
func (w *WebHandler) RenderWebhooks(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	events, _ := w.db.ListWebhookEvents(ctx, 50)
	pluggyWebhooks, err := w.client.ListWebhooks(ctx)
	if err != nil {
		log.Printf("[Web] Aviso ao listar webhooks cadastrados na Pluggy: %v", err)
	}

	data := map[string]any{
		"Title":          "Webhooks & Integrações",
		"ActiveTab":      "webhooks",
		"Events":         events,
		"PluggyWebhooks": pluggyWebhooks,
		"Config":         w.cfg,
	}

	if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "webhooks-events-container" {
		w.renderFragment(rw, "webhooks-table", data)
		return
	}

	w.render(rw, r, "webhooks.html", data)
}

func (w *WebHandler) render(rw http.ResponseWriter, r *http.Request, tmplName string, data map[string]any) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := w.templates.ExecuteTemplate(rw, tmplName, data)
	if err != nil {
		log.Printf("[Web] Erro renderizando template %s: %v", tmplName, err)
		http.Error(rw, fmt.Sprintf("Template Error: %v", err), http.StatusInternalServerError)
	}
}

func (w *WebHandler) renderFragment(rw http.ResponseWriter, fragmentName string, data map[string]any) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := w.templates.ExecuteTemplate(rw, fragmentName, data)
	if err != nil {
		log.Printf("[Web] Erro renderizando fragmento %s: %v", fragmentName, err)
		http.Error(rw, fmt.Sprintf("Fragment Error: %v", err), http.StatusInternalServerError)
	}
}
