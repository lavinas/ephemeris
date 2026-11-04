package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"accounting/internal/domain"
	"accounting/internal/dto"
	"accounting/internal/port"
	"accounting/internal/service"
)

const (
	defaultPageSize = 20
)

// HandlerHtml handles HTML / HTMX rendering for banking statements
type HandlerHtml struct {
	repo        port.Repository
	logger      port.Logger
	listService *service.StatementListService
	syncService *service.StatementSyncService
	tmpl        *template.Template
}

func NewHandlerHtml(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
	templateData []byte,
) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"formatDate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("02/01/2006")
		},
		"formatMoney": func(v float64) string {
			str := fmt.Sprintf("%.2f", v)
			parts := strings.Split(str, ".")
			intPart := parts[0]
			decPart := parts[1]

			// Insert thousand separators
			var res []string
			isNeg := false
			if strings.HasPrefix(intPart, "-") {
				isNeg = true
				intPart = intPart[1:]
			}
			for len(intPart) > 3 {
				res = append([]string{intPart[len(intPart)-3:]}, res...)
				intPart = intPart[:len(intPart)-3]
			}
			if len(intPart) > 0 {
				res = append([]string{intPart}, res...)
			}
			formatted := strings.Join(res, ".")
			if isNeg {
				formatted = "-" + formatted
			}
			return fmt.Sprintf("R$ %s,%s", formatted, decPart)
		},
		"isNegative": func(v float64) bool {
			return v < 0
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}

	tmpl, err := template.New("statements").Funcs(funcMap).Parse(string(templateData))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &HandlerHtml{
		repo:        repo,
		logger:      logger,
		listService: listService,
		syncService: syncService,
		tmpl:        tmpl,
	}, nil
}

type StatementsViewData struct {
	Accounts       []domain.Account
	SelectedAccount *domain.Account
	Statements     []dto.StatementItemResponse
	Total          int64
	Page           int
	PageSize       int
	TotalPages     int
	StartDate      string
	EndDate        string
	SyncMessage    string
}

// Index renders the main statements page
func (h *HandlerHtml) Index(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repo.GetAccounts(true)
	if err != nil {
		h.logger.IPrintf(2, "Error fetching accounts: %v", err)
		http.Error(w, "Erro ao carregar contas bancárias", http.StatusInternalServerError)
		return
	}

	if len(accounts) == 0 {
		h.render(w, "statements.html", StatementsViewData{
			Accounts: accounts,
		})
		return
	}

	q := r.URL.Query()
	accountID, _ := strconv.ParseInt(q.Get("account_id"), 10, 64)
	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	// Default to first active account
	var selectedAccount *domain.Account
	if accountID <= 0 {
		selectedAccount = &accounts[0]
		accountID = selectedAccount.ID
	} else {
		for i := range accounts {
			if accounts[i].ID == accountID {
				selectedAccount = &accounts[i]
				break
			}
		}
		if selectedAccount == nil {
			selectedAccount = &accounts[0]
			accountID = selectedAccount.ID
		}
	}

	startDate := q.Get("start_date")
	endDate := q.Get("end_date")

	resp := h.listService.Run(&dto.StatementListRequest{
		AccountID: accountID,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		PageSize:  defaultPageSize,
	})

	viewData := StatementsViewData{
		Accounts:        accounts,
		SelectedAccount: selectedAccount,
		Statements:      resp.Data,
		Total:           resp.Total,
		Page:            resp.Page,
		PageSize:        resp.PageSize,
		TotalPages:      resp.TotalPages,
		StartDate:       startDate,
		EndDate:         endDate,
	}

	// If HTMX partial request for table only
	if r.Header.Get("HX-Request") == "true" && q.Get("partial") == "table" {
		h.render(w, "table_partial", viewData)
		return
	}

	h.render(w, "statements.html", viewData)
}

// Sync handles manual synchronization trigger via HTMX
func (h *HandlerHtml) Sync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	accountID, _ := strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	startDate := r.FormValue("sync_start_date")
	endDate := r.FormValue("sync_end_date")

	resp := h.syncService.Run(r.Context(), &dto.SyncStatementsRequest{
		AccountID: accountID,
		StartDate: startDate,
		EndDate:   endDate,
	})

	statusClass := "alert-success"
	if resp.GetStatusCode() >= 400 {
		statusClass = "alert-danger"
	}

	msg := fmt.Sprintf("Sincronização concluída: %d buscadas, %d importadas, %d ignoradas (duplicadas).",
		resp.TotalFetched, resp.TotalImported, resp.TotalIgnored)
	if resp.GetStatusCode() >= 400 {
		msg = fmt.Sprintf("Falha na sincronização: %s", resp.GetMessage())
	}

	w.Header().Set("HX-Trigger", `{"reloadStatements": true, "closeSyncModal": true}`)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(fmt.Appendf(nil, `<div class="alert %s alert-dismissible fade show" role="alert">
		%s
		<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
	</div>`, statusClass, msg))
}

func (h *HandlerHtml) render(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		h.logger.IPrintf(2, "Template execution error: %v", err)
		http.Error(w, fmt.Sprintf("Erro ao renderizar template: %v", err), http.StatusInternalServerError)
	}
}
