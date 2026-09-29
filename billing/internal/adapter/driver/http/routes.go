package http

import (
	"net/http"

	"billing/internal/port"
)

// NewRoutes creates the root http.ServeMux combining static, API, and HTML routes.
func NewRoutes(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer, htmlTemplate []byte) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	// Static assets
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	apiRoutes := NewAPIRoutes(repo, logger, taxer, pixer, issuer)
	mux.Handle("/api/", http.StripPrefix("/api", apiRoutes))

	// Direct endpoint mappings for API backward compatibility
	mux.Handle("/customer/api/", apiRoutes)
	mux.Handle("/invoice/api/", apiRoutes)
	mux.Handle("/tax/api/", apiRoutes)
	mux.HandleFunc("/ping", NewHandlerApi(repo, logger, taxer, pixer, issuer).Ping)

	// HTML routes
	htmlRoutes, htmlHandler, err := NewHTMLRoutes(repo, logger, taxer, pixer, issuer, htmlTemplate)
	if err != nil {
		return nil, err
	}
	mux.Handle("/html/", http.StripPrefix("/html", htmlRoutes))

	// Root dashboard
	mux.HandleFunc("/", htmlHandler.Index)

	return mux, nil
}

// NewAPIRoutes creates API routes mux
func NewAPIRoutes(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer) *http.ServeMux {
	mux := http.NewServeMux()
	handler := NewHandlerApi(repo, logger, taxer, pixer, issuer)

	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/customer/create", handler.CustomerCreate)
	mux.HandleFunc("/customer/api/create", handler.CustomerCreate)
	mux.HandleFunc("/customer/list", handler.CustomerList)
	mux.HandleFunc("/customer/api/list", handler.CustomerList)
	mux.HandleFunc("/customer/update", handler.CustomerUpdate)
	mux.HandleFunc("/customer/api/update", handler.CustomerUpdate)

	mux.HandleFunc("/invoice/create", handler.InvoiceCreate)
	mux.HandleFunc("/invoice/api/create", handler.InvoiceCreate)
	mux.HandleFunc("/invoice/list", handler.InvoiceList)
	mux.HandleFunc("/invoice/api/list", handler.InvoiceList)
	mux.HandleFunc("/invoice/update", handler.InvoiceUpdate)
	mux.HandleFunc("/invoice/api/update", handler.InvoiceUpdate)
	mux.HandleFunc("/invoice/payment", handler.InvoicePayment)
	mux.HandleFunc("/invoice/api/payment", handler.InvoicePayment)
	mux.HandleFunc("/invoice/bill", handler.InvoiceBill)
	mux.HandleFunc("/invoice/api/bill", handler.InvoiceBill)

	mux.HandleFunc("/tax/generate", handler.TaxGenerate)
	mux.HandleFunc("/tax/api/generate", handler.TaxGenerate)
	mux.HandleFunc("/tax/clear", handler.TaxClear)
	mux.HandleFunc("/tax/api/clear", handler.TaxClear)

	return mux
}

// NewHTMLRoutes creates HTML routes mux
func NewHTMLRoutes(repo port.Repository, logger port.Logger, taxer port.Taxer, pixer port.Pixer, issuer port.Issuer, htmlTemplate []byte) (*http.ServeMux, *HandlerHtml, error) {
	mux := http.NewServeMux()
	logger.IPrintf(0, "Initializing HTML routes")
	handler, err := NewHandlerHtml(repo, logger, taxer, pixer, issuer, htmlTemplate)
	if err != nil {
		logger.IPrintf(0, "Failed to initialize HTML handler: %v", err)
		return nil, nil, err
	}
	logger.IPrintf(0, "HTML handler initialized successfully")

	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/invoices", handler.Invoices)
	mux.HandleFunc("/invoices/", handler.Invoices)
	mux.HandleFunc("/invoices/novo", handler.InvoicesCreate)
	mux.HandleFunc("/invoices/salvar", handler.InvoicesSave)
	mux.HandleFunc("/invoices/tabela", handler.InvoicesTable)
	mux.HandleFunc("/invoices/tabela/reset", handler.InvoicesTableReset)
	mux.HandleFunc("/invoices/bloco/limpar", handler.InvoicesBlockClear)
	mux.HandleFunc("/invoices/editar", handler.InvoicesEdit)
	mux.HandleFunc("/invoices/atualizar", handler.InvoicesUpdate)
	mux.HandleFunc("/invoices/cancelar-edicao", handler.InvoicesCancelEdit)
	mux.HandleFunc("/invoices/deletar-aviso", handler.InvoicesDeleteWarning)
	mux.HandleFunc("/invoices/deletar", handler.InvoicesDelete)
	mux.HandleFunc("/invoices/pagamento-aviso", handler.InvoicesPaymentWarning)
	mux.HandleFunc("/invoices/pagar", handler.InvoicesPay)
	mux.HandleFunc("/invoices/enviar-invoice", handler.InvoicesSendInvoice)
	mux.HandleFunc("/invoices/enviar-recibo", handler.InvoicesSendReceipt)

	return mux, handler, nil
}
