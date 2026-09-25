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
