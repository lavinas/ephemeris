package http

import (
	"net/http"
	"signup/internal/port"
)

// NewRoutes creates a new instance of Routes with the provided logger, repository, and publisher.
func NewRoutes(repo port.Repository, logger port.Logger, publisher port.CustomerEventPublisher, htmlTemplate []byte) (*http.ServeMux, error) {
	mux := http.NewServeMux()
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	apiRoutes := NewAPIRoutes(repo, logger, publisher)
	mux.Handle("/api/", http.StripPrefix("/api", apiRoutes))
	htmlRoutes, htmlHandler, err := NewHTMLRoutes(repo, logger, publisher, htmlTemplate)
	if err != nil {
		return nil, err
	}
	mux.Handle("/html/", http.StripPrefix("/html", htmlRoutes))
	mux.HandleFunc("/", htmlHandler.Index)
	return mux, nil
}

// NewAPIRoutes creates a new instance of API routes with the provided logger, repository, and publisher.
func NewAPIRoutes(repo port.Repository, logger port.Logger, publisher port.CustomerEventPublisher) *http.ServeMux {
	mux := http.NewServeMux()
	handler := NewHandlerApi(repo, logger, publisher)
	mux.HandleFunc("/ping", handler.Ping)
	mux.HandleFunc("/customer/create", handler.CustomerCreate)
	mux.HandleFunc("/customer/update", handler.CustomerUpdate)
	mux.HandleFunc("/customer/list", handler.CustomerList)
	mux.HandleFunc("/user/create", handler.UserCreate)
	mux.HandleFunc("/user/update", handler.UserUpdate)
	mux.HandleFunc("/user/list", handler.UserList)
	mux.HandleFunc("/vendor/create", handler.VendorCreate)
	mux.HandleFunc("/vendor/update", handler.VendorUpdate)
	mux.HandleFunc("/vendor/list", handler.VendorList)
	return mux
}

// NewHTMLRoutes creates a new instance of HTML routes with the provided repository, logger, publisher and template data.
func NewHTMLRoutes(repo port.Repository, logger port.Logger, publisher port.CustomerEventPublisher, htmlTemplate []byte) (*http.ServeMux, *HandlerHtml, error) {
	mux := http.NewServeMux()
	logger.IPrintf(0, "Initializing HTML routes")
	handler, err := NewHandlerHtml(repo, logger, publisher, htmlTemplate)
	if err != nil {
		logger.IPrintf(0, "Failed to initialize HTML handler: %v", err)
		return nil, nil, err
	}
	logger.IPrintf(0, "HTML handler initialized successfully")
	mux.HandleFunc("/ping", handler.Ping)

	// Customers routes
	mux.HandleFunc("/customers", handler.Customers)
	mux.HandleFunc("/customers/", handler.Customers)
	mux.HandleFunc("/clientes", handler.Customers)
	mux.HandleFunc("/clientes/", handler.Customers)
	mux.HandleFunc("/customers/novo", handler.CustomersCreate)
	mux.HandleFunc("/customers/salvar", handler.CustomersSave)
	mux.HandleFunc("/customers/tabela/reset", handler.CustomersTableReset)
	mux.HandleFunc("/customers/bloco/limpar", handler.CustomersBlockClear)
	mux.HandleFunc("/customers/deletar-aviso", handler.CustomersDeleteWarning)
	mux.HandleFunc("/customers/cancelar-edicao", handler.CustomersCancelEdit)
	mux.HandleFunc("/customers/deletar", handler.CustomersDelete)
	mux.HandleFunc("/customers/editar", handler.CustomersEdit)
	mux.HandleFunc("/customers/atualizar", handler.CustomersUpdate)
	mux.HandleFunc("/customers/tabela", handler.CustomersTable)

	// Users routes
	mux.HandleFunc("/users", handler.Users)
	mux.HandleFunc("/users/", handler.Users)
	mux.HandleFunc("/usuarios", handler.Users)
	mux.HandleFunc("/usuarios/", handler.Users)
	mux.HandleFunc("/users/novo", handler.UsersCreate)
	mux.HandleFunc("/users/salvar", handler.UsersSave)
	mux.HandleFunc("/users/tabela/reset", handler.UsersTableReset)
	mux.HandleFunc("/users/bloco/limpar", handler.UsersBlockClear)
	mux.HandleFunc("/users/deletar-aviso", handler.UsersDeleteWarning)
	mux.HandleFunc("/users/cancelar-edicao", handler.UsersCancelEdit)
	mux.HandleFunc("/users/deletar", handler.UsersDelete)
	mux.HandleFunc("/users/editar", handler.UsersEdit)
	mux.HandleFunc("/users/atualizar", handler.UsersUpdate)
	mux.HandleFunc("/users/tabela", handler.UsersTable)

	return mux, handler, nil
}
