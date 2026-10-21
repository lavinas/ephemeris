package http

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"planner/internal/adapter/billing"
	"planner/internal/domain"
	"planner/internal/port"
	"planner/internal/service"
)

var (
	itensPorPag = 10
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger        port.Logger
	repo          port.Repository
	comptime      port.ComptimeService
	tmpl          *template.Template
	tmplPlans     *template.Template
	tmplInvoicing *template.Template
	tmplComptime  *template.Template
	billingClient port.BillingClient
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, tdata []byte, plansData ...[]byte) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
		"adicionar": func(a, b int) int {
			return a + b
		},
		"subtrair": func(a, b int) int {
			return a - b
		},
	}
	logger.IPrintf(0, "Passou")
	tmpl, err := template.New("index").Funcs(funcMap).Parse(string(tdata))
	if err != nil {
		return nil, err
	}

	var pData []byte
	if len(plansData) > 0 && len(plansData[0]) > 0 {
		pData = plansData[0]
	} else {
		if data, err := os.ReadFile("/app/web/templates/plans.html"); err == nil {
			pData = data
		} else if data, err := os.ReadFile("web/templates/plans.html"); err == nil {
			pData = data
		} else if data, err := os.ReadFile("../../../web/templates/plans.html"); err == nil {
			pData = data
		} else {
			pData = []byte(`{{define "index"}}<html><body>Plans</body></html>{{end}}{{define "fragment"}}<div>Plans</div>{{end}}{{define "tabela"}}<div>Tabela</div>{{end}}{{define "linha_plano_encerrar_aviso"}}<div>Encerrar</div>{{end}}{{define "linha_plano_deletar_aviso"}}<div>Deletar</div>{{end}}{{define "modal_gerar_invoice"}}<div>Modal</div>{{end}}{{define "modal_gerar_invoices"}}<div>Modal Invoices</div>{{end}}{{define "modal_gerar_invoice_resultado"}}<div>Resultado</div>{{end}}`)
		}
	}

	tmplPlans, err := template.New("index").Funcs(funcMap).Parse(string(pData))
	if err != nil {
		return nil, err
	}

	var invData []byte
	if len(plansData) > 1 && len(plansData[1]) > 0 {
		invData = plansData[1]
	} else {
		if data, err := os.ReadFile("/app/web/templates/invoicing.html"); err == nil {
			invData = data
		} else if data, err := os.ReadFile("web/templates/invoicing.html"); err == nil {
			invData = data
		} else if data, err := os.ReadFile("../../../web/templates/invoicing.html"); err == nil {
			invData = data
		} else {
			invData = []byte(`{{define "index"}}<html><body>Invoicing</body></html>{{end}}{{define "fragment"}}<div>Invoicing</div>{{end}}{{define "tabela"}}<div>Tabela</div>{{end}}{{define "linha_invoice"}}<div>Linha</div>{{end}}{{define "linha_invoice_deletar_aviso"}}<div>Aviso</div>{{end}}{{define "linha_invoice_edit"}}<div>Edit</div>{{end}}{{define "modal_deletar_invoices_lote"}}<div>Modal Lote</div>{{end}}`)
		}
	}

	tmplInvoicing, err := template.New("index").Funcs(funcMap).Parse(string(invData))
	if err != nil {
		return nil, err
	}

	var compData []byte
	if len(plansData) > 2 && len(plansData[2]) > 0 {
		compData = plansData[2]
	} else {
		if data, err := os.ReadFile("/app/web/templates/comptime.html"); err == nil {
			compData = data
		} else if data, err := os.ReadFile("web/templates/comptime.html"); err == nil {
			compData = data
		} else if data, err := os.ReadFile("../../../web/templates/comptime.html"); err == nil {
			compData = data
		} else {
			compData = []byte(`{{define "index"}}<html><body>Comptime</body></html>{{end}}{{define "fragment"}}<div>Comptime</div>{{end}}{{define "tabela"}}<div>Tabela</div>{{end}}{{define "modal_resync"}}<div>Resync</div>{{end}}{{define "modal_resync_resultado"}}<div>Resultado</div>{{end}}{{define "linha_comptime"}}<div>Linha</div>{{end}}{{define "linha_comptime_deletar_aviso"}}<div>Aviso</div>{{end}}`)
		}
	}

	tmplComptime, err := template.New("index").Funcs(funcMap).Parse(string(compData))
	if err != nil {
		return nil, err
	}

	var comptime port.ComptimeService
	if cr, ok := repo.(port.ComptimeRepository); ok {
		comptime = service.NewComptime(cr, logger)
	}

	return &HandlerHtml{
		repo:          repo,
		logger:        logger,
		comptime:      comptime,
		tmpl:          tmpl,
		tmplPlans:     tmplPlans,
		tmplInvoicing: tmplInvoicing,
		tmplComptime:  tmplComptime,
	}, nil
}

// SetComptimeService allows injecting a custom ComptimeService
func (h *HandlerHtml) SetComptimeService(cs port.ComptimeService) {
	h.comptime = cs
}

func (h *HandlerHtml) getComptimeService() port.ComptimeService {
	if h.comptime != nil {
		return h.comptime
	}
	if cr, ok := h.repo.(port.ComptimeRepository); ok {
		return service.NewComptime(cr, h.logger)
	}
	return nil
}

// Index handles requests to the root "/" endpoint, serving the static dashboard page.
func (h *HandlerHtml) Index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "web/static/index.html")
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

// getCustomersNicknames retrieves all customer nicknames for vendor 1
func (h *HandlerHtml) getCustomersNicknames() []string {
	customers, err := h.repo.FindCustomers(0, 0, 1, nil, nil, nil, nil, nil, nil)
	if err != nil {
		h.logger.IPrintf(2, "Failed to retrieve customers: %v", err)
		return nil
	}
	nicknames := make([]string, 0, len(customers))
	for _, c := range customers {
		nicknames = append(nicknames, c.Nickname)
	}
	return nicknames
}

// getServices retrieves all services for vendor 1
func (h *HandlerHtml) getServices() []domain.Service {
	services, err := h.repo.FindServices(1)
	if err != nil {
		h.logger.IPrintf(2, "Failed to retrieve services: %v", err)
		return nil
	}
	return services
}

// SetBillingClient sets the billing client for HandlerHtml
func (h *HandlerHtml) SetBillingClient(client port.BillingClient) {
	h.billingClient = client
}

func (h *HandlerHtml) getBillingClient() port.BillingClient {
	if h.billingClient != nil {
		return h.billingClient
	}
	return billing.NewClient(os.Getenv("BILLING_URL"), h.logger)
}
