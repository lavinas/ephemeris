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
	repo          port.Repository
	logger        port.Logger
	listService   *service.StatementListService
	syncService   *service.StatementSyncService
	actionService *service.AccountActionService
	tmpl          *template.Template
}

func NewHandlerHtml(
	repo port.Repository,
	logger port.Logger,
	listService *service.StatementListService,
	syncService *service.StatementSyncService,
	actionService *service.AccountActionService,
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
		"fieldName":    domain.FieldName,
		"operatorName": domain.OperatorName,
		"actionName":   domain.ActionName,
	}

	tmpl, err := template.New("statements").Funcs(funcMap).Parse(string(templateData))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &HandlerHtml{
		repo:          repo,
		logger:        logger,
		listService:   listService,
		syncService:   syncService,
		actionService: actionService,
		tmpl:          tmpl,
	}, nil
}

type StatementsViewData struct {
	Accounts        []domain.Account
	SelectedAccount *domain.Account
	Statements      []dto.StatementItemResponse
	Total           int64
	Page            int
	PageSize        int
	TotalPages      int
	StartDate       string
	EndDate         string
	Type            string // all, credit, debit
	SyncMessage     string
}

// Index renders the main statements page
func (h *HandlerHtml) Index(w http.ResponseWriter, r *http.Request) {
	h.logger.IPrintf(0, "HandlerHtml.Index called: method=%s, path=%s, HX-Request=%s", r.Method, r.URL.Path, r.Header.Get("HX-Request"))

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
	if accountID <= 0 {
		accountID, _ = strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	}
	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page, _ = strconv.Atoi(r.FormValue("page"))
	}
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
	if startDate == "" {
		startDate = r.FormValue("start_date")
	}
	endDate := q.Get("end_date")
	if endDate == "" {
		endDate = r.FormValue("end_date")
	}

	txType := q.Get("type")
	if txType == "" {
		txType = r.FormValue("type")
	}
	if txType == "" {
		txType = "all"
	}

	resp := h.listService.Run(&dto.StatementListRequest{
		AccountID: accountID,
		StartDate: startDate,
		EndDate:   endDate,
		Type:      txType,
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
		Type:            txType,
	}

	// If HTMX partial request for table only
	if r.Header.Get("HX-Request") == "true" && (q.Get("partial") == "table" || r.URL.Path == "/statements/tabela" || r.URL.Path == "/extratos/tabela") {
		h.render(w, "tabela", viewData)
		return
	}

	// If HTMX partial request from vendor shell (#conteudo-principal)
	if r.Header.Get("HX-Request") == "true" {
		h.render(w, "fragment", viewData)
		return
	}

	h.render(w, "statements.html", viewData)
}

// Table renders only the statements table
func (h *HandlerHtml) Table(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.IPrintf(1, "Warning: failed to parse form in Table: %v", err)
	}
	h.Index(w, r)
}

// TableReset resets filters and renders default table
func (h *HandlerHtml) TableReset(w http.ResponseWriter, r *http.Request) {
	r.Form = nil
	if r.URL != nil {
		newQ := r.URL.Query()
		newQ.Del("start_date")
		newQ.Del("end_date")
		newQ.Del("type")
		newQ.Del("page")
		r.URL.RawQuery = newQ.Encode()
	}
	h.Index(w, r)
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
	delVal := strings.ToLower(strings.TrimSpace(r.FormValue("delete_existing")))
	deleteExisting := delVal == "true" || delVal == "1" || delVal == "on"

	resp := h.syncService.Run(r.Context(), &dto.SyncStatementsRequest{
		AccountID:      accountID,
		StartDate:      startDate,
		EndDate:        endDate,
		DeleteExisting: deleteExisting,
	})

	statusClass := "alert-success"
	if resp.GetStatusCode() >= 400 {
		statusClass = "alert-danger"
	}

	msg := fmt.Sprintf("Sincronização concluída: %d buscadas, %d importadas, %d ignoradas.",
		resp.TotalFetched, resp.TotalImported, resp.TotalIgnored)
	if resp.TotalDeleted > 0 {
		msg = fmt.Sprintf("Sincronização concluída: %d excluídas no período, %d buscadas, %d importadas, %d ignoradas.",
			resp.TotalDeleted, resp.TotalFetched, resp.TotalImported, resp.TotalIgnored)
	}
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

type ActionModalViewData struct {
	AccountID       int64
	SelectedAccount *domain.Account
	Actions         []dto.AccountActionItemDTO
	Action          *dto.AccountActionItemDTO
	ErrorMessage    string
	SuccessMessage  string
}

// ActionsModal renders the actions modal wrapper for an account
func (h *HandlerHtml) ActionsModal(w http.ResponseWriter, r *http.Request) {
	accountID, _ := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 64)
	if accountID <= 0 {
		accountID, _ = strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	}
	acc, _ := h.repo.GetAccountByID(accountID)
	resp := h.actionService.ListByAccount(accountID)

	h.render(w, "modal_actions", ActionModalViewData{
		AccountID:       accountID,
		SelectedAccount: acc,
		Actions:         resp.Data,
	})
}

// ActionsList renders only the list of actions
func (h *HandlerHtml) ActionsList(w http.ResponseWriter, r *http.Request) {
	accountID, _ := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 64)
	if accountID <= 0 {
		accountID, _ = strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	}
	acc, _ := h.repo.GetAccountByID(accountID)
	resp := h.actionService.ListByAccount(accountID)

	h.render(w, "modal_actions_list", ActionModalViewData{
		AccountID:       accountID,
		SelectedAccount: acc,
		Actions:         resp.Data,
	})
}

// ActionsNew renders the creation form for a new action
func (h *HandlerHtml) ActionsNew(w http.ResponseWriter, r *http.Request) {
	accountID, _ := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 64)
	if accountID <= 0 {
		accountID, _ = strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	}
	acc, _ := h.repo.GetAccountByID(accountID)

	actionItem := &dto.AccountActionItemDTO{
		AccountID: accountID,
		Action:    domain.ActionIgnoreTransaction,
		Rules: []dto.ActionRuleDTO{
			{Field: domain.RuleFieldTransaction, Operator: domain.RuleOpContains, Value: ""},
		},
	}

	h.render(w, "modal_action_form", ActionModalViewData{
		AccountID:       accountID,
		SelectedAccount: acc,
		Action:          actionItem,
	})
}

// ActionsEdit renders the edit form for an existing action
func (h *HandlerHtml) ActionsEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	resp := h.actionService.GetByID(id)
	if resp.Data == nil {
		http.Error(w, "Ação não encontrada", http.StatusNotFound)
		return
	}

	acc, _ := h.repo.GetAccountByID(resp.Data.AccountID)

	h.render(w, "modal_action_form", ActionModalViewData{
		AccountID:       resp.Data.AccountID,
		SelectedAccount: acc,
		Action:          resp.Data,
	})
}

// ActionsSave handles saving (creating or updating) an action
func (h *HandlerHtml) ActionsSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.IPrintf(1, "Warning: failed to parse form in ActionsSave: %v", err)
	}

	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	accountID, _ := strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	actionType, _ := strconv.Atoi(r.FormValue("action"))
	if actionType <= 0 {
		actionType = domain.ActionIgnoreTransaction
	}

	fields := r.Form["rule_field"]
	ops := r.Form["rule_operator"]
	vals := r.Form["rule_value"]
	caseSensitives := r.Form["rule_case_sensitive"]

	var rules []dto.ActionRuleDTO
	for i := 0; i < len(fields); i++ {
		f, _ := strconv.Atoi(fields[i])
		op, _ := strconv.Atoi(ops[i])
		v := strings.TrimSpace(vals[i])
		if v == "" {
			continue
		}
		cs := false
		if i < len(caseSensitives) && (caseSensitives[i] == "true" || caseSensitives[i] == "1") {
			cs = true
		}
		rules = append(rules, dto.ActionRuleDTO{
			Field:         f,
			Operator:      op,
			Value:         v,
			CaseSensitive: cs,
		})
	}

	if len(rules) == 0 {
		acc, _ := h.repo.GetAccountByID(accountID)
		h.render(w, "modal_action_form", ActionModalViewData{
			AccountID:       accountID,
			SelectedAccount: acc,
			ErrorMessage:    "Informe ao menos uma regra com valor preenchido.",
			Action: &dto.AccountActionItemDTO{
				ID:        id,
				AccountID: accountID,
				Action:    actionType,
				Rules: []dto.ActionRuleDTO{
					{Field: domain.RuleFieldTransaction, Operator: domain.RuleOpContains, Value: ""},
				},
			},
		})
		return
	}

	if id > 0 {
		h.actionService.Update(&dto.AccountActionUpdateRequest{
			ID:        id,
			AccountID: accountID,
			Action:    actionType,
			Rules:     rules,
		})
	} else {
		h.actionService.Create(&dto.AccountActionCreateRequest{
			AccountID: accountID,
			Action:    actionType,
			Rules:     rules,
		})
	}

	// Re-render the actions list
	acc, _ := h.repo.GetAccountByID(accountID)
	resp := h.actionService.ListByAccount(accountID)
	h.render(w, "modal_actions_list", ActionModalViewData{
		AccountID:       accountID,
		SelectedAccount: acc,
		Actions:         resp.Data,
		SuccessMessage:  "Ação salva com sucesso!",
	})
}

// ActionsDelete handles deleting an action
func (h *HandlerHtml) ActionsDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if id <= 0 {
		id, _ = strconv.ParseInt(r.FormValue("id"), 10, 64)
	}
	accountID, _ := strconv.ParseInt(r.URL.Query().Get("account_id"), 10, 64)
	if accountID <= 0 {
		accountID, _ = strconv.ParseInt(r.FormValue("account_id"), 10, 64)
	}

	h.actionService.Delete(id)

	acc, _ := h.repo.GetAccountByID(accountID)
	resp := h.actionService.ListByAccount(accountID)
	h.render(w, "modal_actions_list", ActionModalViewData{
		AccountID:       accountID,
		SelectedAccount: acc,
		Actions:         resp.Data,
		SuccessMessage:  "Ação removida com sucesso!",
	})
}
