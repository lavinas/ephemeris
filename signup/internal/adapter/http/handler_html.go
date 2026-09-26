package http

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"signup/internal/domain"
	"signup/internal/dto"
	"signup/internal/port"
	"signup/internal/service"
)

var (
	mu          sync.Mutex
	itensPorPag = 10
)

// HandlerHtml is an HTTP handler for the HTML pages
type HandlerHtml struct {
	logger    port.Logger
	repo      port.Repository
	publisher port.CustomerEventPublisher
	tmpl      *template.Template
}

// NewHandlerHtml creates a new instance of HandlerHtml
func NewHandlerHtml(repo port.Repository, logger port.Logger, publisher port.CustomerEventPublisher, tdata []byte) (*HandlerHtml, error) {
	funcMap := template.FuncMap{
		"pularLinhas": func(texto string) template.HTML {
			safeStr := template.HTMLEscapeString(texto)
			comQuebras := strings.ReplaceAll(safeStr, "\n", "<br>")
			return template.HTML(comQuebras)
		},
	}
	tmpl, err := template.New("index").Funcs(funcMap).Parse(string(tdata))
	if err != nil {
		return nil, err
	}
	return &HandlerHtml{
		repo:      repo,
		logger:    logger,
		publisher: publisher,
		tmpl:      tmpl,
	}, nil
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

// Ping handler for the /html/ping endpoint
func (h *HandlerHtml) Ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	svc := service.NewPingService(h.logger)
	response := svc.Run(nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.GetStatusCode())
	w.Write([]byte("pong"))
}

// ==========================================
// CUSTOMERS HANDLERS
// ==========================================

// Customers handler for the /html/customers endpoint
func (h *HandlerHtml) Customers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)

	svc := service.NewList(h.repo, h.logger)
	req := dto.NewCustomerListRequest(h.repo)
	req.Page = 1
	req.PageSize = itensPorPag
	req.Vendor = vendorSel

	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve customers", http.StatusInternalServerError)
		return
	}
	var response *dto.CustomerListResponse
	if rPtr, ok := respdro.(*dto.CustomerListResponse); ok {
		response = rPtr
	} else if rVal, ok := respdro.(dto.CustomerListResponse); ok {
		response = &rVal
	} else {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageDataCustomer(response, vendors, vendorSel, 1, "", 0)

	tmplName := "customers_index"
	if r.Header.Get("HX-Request") == "true" {
		tmplName = "customers_fragment"
	}
	if err := h.tmpl.ExecuteTemplate(w, tmplName, page); err != nil {
		h.logger.IPrintf(0, "Failed to render customer template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// CustomersCreate handler for opening new customer form
func (h *HandlerHtml) CustomersCreate(w http.ResponseWriter, r *http.Request) {
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)
	data := map[string]interface{}{
		"Vendors":        vendors,
		"VendorSelected": vendorSel,
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro_customer", data)
}

// CustomersSave handler for creating a customer
func (h *HandlerHtml) CustomersSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	vendor := r.FormValue("add_vendor")
	if vendor == "" {
		vendor = r.FormValue("vendor")
	}
	if vendor == "" {
		vendor = h.getDefaultVendor(h.getVendors())
	}

	createReq := dto.NewCustomerCreateRequest(h.repo)
	createReq.Vendor = vendor
	createReq.Name = strings.TrimSpace(r.FormValue("add_name"))
	createReq.Nickname = strings.TrimSpace(r.FormValue("add_nickname"))
	createReq.Document = strings.TrimSpace(r.FormValue("add_document"))
	createReq.Email = strings.TrimSpace(r.FormValue("add_email"))
	createReq.Whatsapp = strings.TrimSpace(r.FormValue("add_whatsapp"))

	saveSvc := service.NewSave(h.repo, h.logger, h.publisher)
	respOut := saveSvc.Run(createReq)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}

	if respOut.GetStatusCode() != 200 {
		msg := getResponseMessage(respOut)
		formData := map[string]interface{}{
			"Vendors":        h.getVendors(),
			"VendorSelected": vendor,
			"AddVendor":      vendor,
			"AddNickname":    createReq.Nickname,
			"AddName":        createReq.Name,
			"AddDocument":    createReq.Document,
			"AddEmail":       createReq.Email,
			"AddWhatsapp":    createReq.Whatsapp,
		}
		var formBuf strings.Builder
		_ = h.tmpl.ExecuteTemplate(&formBuf, "formulario_cadastro_customer", formData)
		w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML">` + formBuf.String() + `</div>`))
		h.renderCustomerTableWithParamsAndError(w, r, vendor, pagina, "Erro ao salvar cliente: "+msg, 0, nil)
		return
	}

	// Success: clear the form container via OOB swap
	w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML"></div>`))
	h.renderCustomerTable(w, r, vendor, pagina)
}

// CustomersTableReset handler to reset filters
func (h *HandlerHtml) CustomersTableReset(w http.ResponseWriter, r *http.Request) {
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)
	w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
	h.renderCustomerTableWithParams(w, vendorSel, "", "", "", "", "", -1, 1)
}

// CustomersBlockClear clears form container
func (h *HandlerHtml) CustomersBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// CustomersDeleteWarning renders the confirmation row
func (h *HandlerHtml) CustomersDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	customer := domain.StartCustomer(h.repo)
	ok, err := customer.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}
	data := map[string]interface{}{
		"ID":       customer.ID,
		"Nickname": customer.Nickname,
		"Name":     customer.Name,
	}
	h.tmpl.ExecuteTemplate(w, "linha_customer_deletar_aviso", data)
}

// CustomersCancelEdit cancels edit/delete view
func (h *HandlerHtml) CustomersCancelEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	customer := domain.StartCustomer(h.repo)
	ok, err := customer.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}
	data := h.customerToMap(customer)
	h.tmpl.ExecuteTemplate(w, "linha_customer", data)
}

// CustomersDelete sets customer status to 0 (inactivation)
func (h *HandlerHtml) CustomersDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	customer := domain.StartCustomer(h.repo)
	ok, err := customer.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	vendor := domain.StartVendor(h.repo)
	if okV, _ := vendor.GetByID(customer.VendorID); okV {
		zero := 0
		updateReq := dto.NewCustomerUpdateRequest(h.repo)
		updateReq.Vendor = vendor.Nickname
		updateReq.Nickname = customer.Nickname
		updateReq.Status = &zero
		saveSvc := service.NewSave(h.repo, h.logger, h.publisher)
		saveSvc.Run(updateReq)
	} else {
		zero := 0
		customer.Status = &zero
		_ = customer.Save()
	}

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	vendorSel := r.FormValue("vendor")
	if vendorSel == "" {
		vendorSel = h.getDefaultVendor(h.getVendors())
	}
	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.renderCustomerTable(w, r, vendorSel, pagina)
}

// CustomersEdit renders the inline edit row
func (h *HandlerHtml) CustomersEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	customer := domain.StartCustomer(h.repo)
	ok, err := customer.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}
	data := h.customerToMap(customer)
	h.tmpl.ExecuteTemplate(w, "linha_customer_edit", data)
}

// CustomersUpdate updates an existing customer
func (h *HandlerHtml) CustomersUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	customer := domain.StartCustomer(h.repo)
	ok, err := customer.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	vendor := domain.StartVendor(h.repo)
	vendorNick := ""
	if okV, _ := vendor.GetByID(customer.VendorID); okV {
		vendorNick = vendor.Nickname
	}

	name := r.FormValue("edit_name")
	doc := r.FormValue("edit_document")
	email := r.FormValue("edit_email")
	whatsapp := r.FormValue("edit_whatsapp")
	status, _ := strconv.Atoi(r.FormValue("edit_status"))

	updateReq := dto.NewCustomerUpdateRequest(h.repo)
	updateReq.Vendor = vendorNick
	updateReq.Nickname = customer.Nickname
	updateReq.Name = &name
	updateReq.Document = &doc
	updateReq.Email = &email
	updateReq.Whatsapp = &whatsapp
	updateReq.Status = &status

	saveSvc := service.NewSave(h.repo, h.logger, h.publisher)
	respOut := saveSvc.Run(updateReq)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}

	if respOut.GetStatusCode() != 200 {
		msg := getResponseMessage(respOut)
		editedOverride := &dto.CustomerDTO{
			ID:       int64(id),
			Nickname: customer.Nickname,
			Name:     name,
			Document: doc,
			Email:    email,
			Whatsapp: whatsapp,
			Status:   status,
		}
		h.renderCustomerTableWithParamsAndError(w, r, vendorNick, pagina, "Erro ao atualizar cliente: "+msg, int64(id), editedOverride)
		return
	}

	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.renderCustomerTable(w, r, vendorNick, pagina)
}

// CustomersTable handles filtering and pagination
func (h *HandlerHtml) CustomersTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	vendor := r.FormValue("vendor")
	if vendor == "" {
		vendor = h.getDefaultVendor(h.getVendors())
	}
	h.renderCustomerTable(w, r, vendor, pagina)
}

func (h *HandlerHtml) renderCustomerTable(w http.ResponseWriter, r *http.Request, vendor string, page int) {
	h.renderCustomerTableWithParamsAndError(w, r, vendor, page, "", 0, nil)
}

func (h *HandlerHtml) renderCustomerTableWithParams(w http.ResponseWriter, vendor, nickname, name, document, email, whatsapp string, statusVal, page int) {
	svc := service.NewList(h.repo, h.logger)
	req := dto.NewCustomerListRequest(h.repo)
	req.Vendor = vendor
	req.Page = page
	req.PageSize = itensPorPag
	if nickname != "" {
		req.Nickname = &nickname
	}
	if name != "" {
		req.Name = &name
	}
	if document != "" {
		req.Document = &document
	}
	if email != "" {
		req.Email = &email
	}
	if whatsapp != "" {
		req.Whatsapp = &whatsapp
	}
	if statusVal != -1 {
		req.Status = &statusVal
	}

	respdro := svc.Run(req)
	var response *dto.CustomerListResponse
	if respdro.GetStatusCode() == 200 {
		if rPtr, ok := respdro.(*dto.CustomerListResponse); ok {
			response = rPtr
		} else if rVal, ok := respdro.(dto.CustomerListResponse); ok {
			response = &rVal
		}
	}
	if response == nil {
		response = &dto.CustomerListResponse{Customers: []dto.CustomerDTO{}}
	}
	vendors := h.getVendors()
	pageData := h.getPageDataCustomer(response, vendors, vendor, page, "", 0)
	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela_customer", pageData)
}

func (h *HandlerHtml) renderCustomerTableWithParamsAndError(w http.ResponseWriter, r *http.Request, vendor string, page int, errMsg string, editingID int64, override *dto.CustomerDTO) {
	nickname := r.FormValue("nickname")
	name := r.FormValue("name")
	document := r.FormValue("document")
	email := r.FormValue("email")
	whatsapp := r.FormValue("whatsapp")
	statusVal := -1
	if s := r.FormValue("status"); s != "" {
		statusVal, _ = strconv.Atoi(s)
	}

	svc := service.NewList(h.repo, h.logger)
	req := dto.NewCustomerListRequest(h.repo)
	req.Vendor = vendor
	req.Page = page
	req.PageSize = itensPorPag
	if nickname != "" {
		req.Nickname = &nickname
	}
	if name != "" {
		req.Name = &name
	}
	if document != "" {
		req.Document = &document
	}
	if email != "" {
		req.Email = &email
	}
	if whatsapp != "" {
		req.Whatsapp = &whatsapp
	}
	if statusVal != -1 {
		req.Status = &statusVal
	}

	respdro := svc.Run(req)
	var response *dto.CustomerListResponse
	if respdro.GetStatusCode() == 200 {
		if rPtr, ok := respdro.(*dto.CustomerListResponse); ok {
			response = rPtr
		} else if rVal, ok := respdro.(dto.CustomerListResponse); ok {
			response = &rVal
		}
	}
	if response == nil {
		response = &dto.CustomerListResponse{Customers: []dto.CustomerDTO{}}
	}

	if override != nil {
		found := false
		for i := range response.Customers {
			if response.Customers[i].ID == override.ID {
				response.Customers[i] = *override
				found = true
				break
			}
		}
		if !found {
			response.Customers = append([]dto.CustomerDTO{*override}, response.Customers...)
		}
	}

	vendors := h.getVendors()
	pageData := h.getPageDataCustomer(response, vendors, vendor, page, errMsg, editingID)
	h.tmpl.ExecuteTemplate(w, "tabela_customer", pageData)
}

// ==========================================
// USERS HANDLERS
// ==========================================

// Users handler for the /html/users endpoint
func (h *HandlerHtml) Users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)

	svc := service.NewList(h.repo, h.logger)
	req := dto.NewUserListRequest(h.repo)
	req.Page = 1
	req.PageSize = itensPorPag
	req.Vendor = vendorSel

	respdro := svc.Run(req)
	if respdro.GetStatusCode() != 200 {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	var response *dto.UserListResponse
	if rPtr, ok := respdro.(*dto.UserListResponse); ok {
		response = rPtr
	} else if rVal, ok := respdro.(dto.UserListResponse); ok {
		response = &rVal
	} else {
		http.Error(w, "Invalid response type", http.StatusInternalServerError)
		return
	}
	page := h.getPageDataUser(response, vendors, vendorSel, 1, "", 0)

	tmplName := "users_index"
	if r.Header.Get("HX-Request") == "true" {
		tmplName = "users_fragment"
	}
	if err := h.tmpl.ExecuteTemplate(w, tmplName, page); err != nil {
		h.logger.IPrintf(0, "Failed to render user template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// UsersCreate handler for opening new user form
func (h *HandlerHtml) UsersCreate(w http.ResponseWriter, r *http.Request) {
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)
	data := map[string]interface{}{
		"Vendors":        vendors,
		"VendorSelected": vendorSel,
	}
	h.tmpl.ExecuteTemplate(w, "formulario_cadastro_user", data)
}

// UsersSave handler for creating a user
func (h *HandlerHtml) UsersSave(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	vendor := r.FormValue("add_vendor")
	if vendor == "" {
		vendor = r.FormValue("vendor")
	}
	if vendor == "" {
		vendor = h.getDefaultVendor(h.getVendors())
	}

	createReq := dto.NewUserCreateRequest(h.repo)
	createReq.Vendor = vendor
	createReq.Username = strings.TrimSpace(r.FormValue("add_username"))
	createReq.Name = strings.TrimSpace(r.FormValue("add_name"))
	createReq.Password = r.FormValue("add_password")
	createReq.Email = strings.TrimSpace(r.FormValue("add_email"))
	createReq.Whatsapp = strings.TrimSpace(r.FormValue("add_whatsapp"))

	saveSvc := service.NewSave(h.repo, h.logger, h.publisher)
	respOut := saveSvc.Run(createReq)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}

	if respOut.GetStatusCode() != 200 {
		msg := getResponseMessage(respOut)
		formData := map[string]interface{}{
			"Vendors":        h.getVendors(),
			"VendorSelected": vendor,
			"AddVendor":      vendor,
			"AddUsername":    createReq.Username,
			"AddName":        createReq.Name,
			"AddPassword":    createReq.Password,
			"AddEmail":       createReq.Email,
			"AddWhatsapp":    createReq.Whatsapp,
		}
		var formBuf strings.Builder
		_ = h.tmpl.ExecuteTemplate(&formBuf, "formulario_cadastro_user", formData)
		w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML">` + formBuf.String() + `</div>`))
		h.renderUserTableWithParamsAndError(w, r, vendor, pagina, "Erro ao salvar usuário: "+msg, 0, nil)
		return
	}

	// Success: clear the form container via OOB swap
	w.Write([]byte(`<div id="formulario-cadastro-container" hx-swap-oob="innerHTML"></div>`))
	h.renderUserTable(w, r, vendor, pagina)
}

// UsersTableReset handler to reset filters
func (h *HandlerHtml) UsersTableReset(w http.ResponseWriter, r *http.Request) {
	vendors := h.getVendors()
	vendorSel := h.getDefaultVendor(vendors)
	w.Write([]byte(`<script>document.getElementById("filtro-form").reset(); document.getElementById("input-pagina-form").value="1";</script>`))
	h.renderUserTableWithParams(w, vendorSel, "", "", "", "", -1, 1)
}

// UsersBlockClear clears form container
func (h *HandlerHtml) UsersBlockClear(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

// UsersDeleteWarning renders the confirmation row
func (h *HandlerHtml) UsersDeleteWarning(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user := domain.StartUser(h.repo)
	ok, err := user.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	data := map[string]interface{}{
		"ID":       user.ID,
		"Username": user.Username,
		"Name":     user.Name,
	}
	h.tmpl.ExecuteTemplate(w, "linha_user_deletar_aviso", data)
}

// UsersCancelEdit cancels edit/delete view
func (h *HandlerHtml) UsersCancelEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user := domain.StartUser(h.repo)
	ok, err := user.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	data := h.userToMap(user)
	h.tmpl.ExecuteTemplate(w, "linha_user", data)
}

// UsersDelete sets user status to 0 (inactivation)
func (h *HandlerHtml) UsersDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user := domain.StartUser(h.repo)
	ok, err := user.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	zero := 0
	user.Status = &zero
	_ = user.Save()

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	vendorSel := r.FormValue("vendor")
	if vendorSel == "" {
		vendorSel = h.getDefaultVendor(h.getVendors())
	}
	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.renderUserTable(w, r, vendorSel, pagina)
}

// UsersEdit renders the inline edit row
func (h *HandlerHtml) UsersEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user := domain.StartUser(h.repo)
	ok, err := user.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	data := h.userToMap(user)
	h.tmpl.ExecuteTemplate(w, "linha_user_edit", data)
}

// UsersUpdate updates an existing user
func (h *HandlerHtml) UsersUpdate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	user := domain.StartUser(h.repo)
	ok, err := user.GetByID(int64(id))
	if err != nil || !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	vendor := domain.StartVendor(h.repo)
	vendorNick := ""
	if okV, _ := vendor.GetByID(user.VendorID); okV {
		vendorNick = vendor.Nickname
	}

	name := r.FormValue("edit_name")
	email := r.FormValue("edit_email")
	whatsapp := r.FormValue("edit_whatsapp")
	status, _ := strconv.Atoi(r.FormValue("edit_status"))

	updateReq := dto.NewUserUpdateRequest(h.repo)
	updateReq.Vendor = vendorNick
	updateReq.Username = user.Username
	updateReq.Name = &name
	updateReq.Email = &email
	updateReq.Whatsapp = &whatsapp
	updateReq.Status = &status

	saveSvc := service.NewSave(h.repo, h.logger, h.publisher)
	respOut := saveSvc.Run(updateReq)

	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}

	if respOut.GetStatusCode() != 200 {
		msg := getResponseMessage(respOut)
		editedOverride := &dto.UserListItem{
			ID:       int64(id),
			Username: user.Username,
			Name:     name,
			Email:    email,
			Whatsapp: whatsapp,
			Status:   status,
		}
		h.renderUserTableWithParamsAndError(w, r, vendorNick, pagina, "Erro ao atualizar usuário: "+msg, int64(id), editedOverride)
		return
	}

	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.renderUserTable(w, r, vendorNick, pagina)
}

// UsersTable handles filtering and pagination
func (h *HandlerHtml) UsersTable(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	pagina, _ := strconv.Atoi(r.FormValue("page"))
	if pagina <= 0 {
		pagina = 1
	}
	vendor := r.FormValue("vendor")
	if vendor == "" {
		vendor = h.getDefaultVendor(h.getVendors())
	}
	h.renderUserTable(w, r, vendor, pagina)
}

func (h *HandlerHtml) renderUserTable(w http.ResponseWriter, r *http.Request, vendor string, page int) {
	h.renderUserTableWithParamsAndError(w, r, vendor, page, "", 0, nil)
}

func (h *HandlerHtml) renderUserTableWithParams(w http.ResponseWriter, vendor, username, name, email, whatsapp string, statusVal, page int) {
	svc := service.NewList(h.repo, h.logger)
	req := dto.NewUserListRequest(h.repo)
	req.Vendor = vendor
	req.Page = page
	req.PageSize = itensPorPag
	if username != "" {
		req.Username = &username
	}
	if name != "" {
		req.Name = &name
	}
	if email != "" {
		req.Email = &email
	}
	if whatsapp != "" {
		req.Whatsapp = &whatsapp
	}
	if statusVal != -1 {
		req.Status = &statusVal
	}

	respdro := svc.Run(req)
	var response *dto.UserListResponse
	if respdro.GetStatusCode() == 200 {
		if rPtr, ok := respdro.(*dto.UserListResponse); ok {
			response = rPtr
		} else if rVal, ok := respdro.(dto.UserListResponse); ok {
			response = &rVal
		}
	}
	if response == nil {
		response = &dto.UserListResponse{Users: []dto.UserListItem{}}
	}
	vendors := h.getVendors()
	pageData := h.getPageDataUser(response, vendors, vendor, page, "", 0)
	w.Write([]byte(`<script>if(document.getElementById("formulario-cadastro-container")) document.getElementById("formulario-cadastro-container").innerHTML = "";</script>`))
	h.tmpl.ExecuteTemplate(w, "tabela_user", pageData)
}

func (h *HandlerHtml) renderUserTableWithParamsAndError(w http.ResponseWriter, r *http.Request, vendor string, page int, errMsg string, editingID int64, override *dto.UserListItem) {
	username := r.FormValue("username")
	name := r.FormValue("name")
	email := r.FormValue("email")
	whatsapp := r.FormValue("whatsapp")
	statusVal := -1
	if s := r.FormValue("status"); s != "" {
		statusVal, _ = strconv.Atoi(s)
	}

	svc := service.NewList(h.repo, h.logger)
	req := dto.NewUserListRequest(h.repo)
	req.Vendor = vendor
	req.Page = page
	req.PageSize = itensPorPag
	if username != "" {
		req.Username = &username
	}
	if name != "" {
		req.Name = &name
	}
	if email != "" {
		req.Email = &email
	}
	if whatsapp != "" {
		req.Whatsapp = &whatsapp
	}
	if statusVal != -1 {
		req.Status = &statusVal
	}

	respdro := svc.Run(req)
	var response *dto.UserListResponse
	if respdro.GetStatusCode() == 200 {
		if rPtr, ok := respdro.(*dto.UserListResponse); ok {
			response = rPtr
		} else if rVal, ok := respdro.(dto.UserListResponse); ok {
			response = &rVal
		}
	}
	if response == nil {
		response = &dto.UserListResponse{Users: []dto.UserListItem{}}
	}

	if override != nil {
		found := false
		for i := range response.Users {
			if response.Users[i].ID == override.ID {
				response.Users[i] = *override
				found = true
				break
			}
		}
		if !found {
			response.Users = append([]dto.UserListItem{*override}, response.Users...)
		}
	}

	vendors := h.getVendors()
	pageData := h.getPageDataUser(response, vendors, vendor, page, errMsg, editingID)
	h.tmpl.ExecuteTemplate(w, "tabela_user", pageData)
}

// ==========================================
// HELPERS
// ==========================================

func (h *HandlerHtml) getVendors() []domain.Vendor {
	vDom := domain.StartVendor(h.repo)
	domList, err := vDom.Find(1, 100)
	if err != nil || len(domList) == 0 {
		return []domain.Vendor{}
	}
	res := make([]domain.Vendor, 0, len(domList))
	for _, d := range domList {
		if v, ok := d.(*domain.Vendor); ok {
			res = append(res, *v)
		}
	}
	return res
}

func (h *HandlerHtml) getDefaultVendor(vendors []domain.Vendor) string {
	if len(vendors) > 0 {
		return vendors[0].Nickname
	}
	return ""
}

func (h *HandlerHtml) customerToMap(c *domain.Customer) map[string]interface{} {
	docStr := "-"
	if c.Document != nil {
		docStr = *c.Document
	}
	emailStr := "-"
	if c.Email != nil {
		emailStr = *c.Email
	}
	whatsappStr := "-"
	if c.Whatsapp != nil {
		whatsappStr = *c.Whatsapp
	}
	statusVal := 0
	if c.Status != nil {
		statusVal = *c.Status
	}
	return map[string]interface{}{
		"ID":       c.ID,
		"Nickname": c.Nickname,
		"Name":     c.Name,
		"Document": docStr,
		"Email":    emailStr,
		"Whatsapp": whatsappStr,
		"Status":   statusVal,
	}
}

func (h *HandlerHtml) userToMap(u *domain.User) map[string]interface{} {
	emailStr := "-"
	if u.Email != nil {
		emailStr = *u.Email
	}
	whatsappStr := "-"
	if u.Whatsapp != nil {
		whatsappStr = *u.Whatsapp
	}
	statusVal := 0
	if u.Status != nil {
		statusVal = *u.Status
	}
	return map[string]interface{}{
		"ID":       u.ID,
		"Username": u.Username,
		"Name":     u.Name,
		"Email":    emailStr,
		"Whatsapp": whatsappStr,
		"Status":   statusVal,
	}
}

func (h *HandlerHtml) getPageDataCustomer(response *dto.CustomerListResponse, vendors []domain.Vendor, vendorSelected string, page int, errMsg string, editingID int64) map[string]interface{} {
	customers := response.Customers
	totalPages := page
	if len(customers) >= itensPorPag {
		totalPages = page + 1
	}
	return map[string]interface{}{
		"Customers":      customers,
		"Vendors":        vendors,
		"VendorSelected": vendorSelected,
		"PaginaAtual":    page,
		"TotalPaginas":   totalPages,
		"TemAnterior":    page > 1,
		"TemProximo":     page < totalPages,
		"PagAnterior":    page - 1,
		"PagProxima":     page + 1,
		"ErrorMessage":   errMsg,
		"EditingID":      editingID,
	}
}

func (h *HandlerHtml) getPageDataUser(response *dto.UserListResponse, vendors []domain.Vendor, vendorSelected string, page int, errMsg string, editingID int64) map[string]interface{} {
	users := response.Users
	totalPages := page
	if len(users) >= itensPorPag {
		totalPages = page + 1
	}
	return map[string]interface{}{
		"Users":          users,
		"Vendors":        vendors,
		"VendorSelected": vendorSelected,
		"PaginaAtual":    page,
		"TotalPaginas":   totalPages,
		"TemAnterior":    page > 1,
		"TemProximo":     page < totalPages,
		"PagAnterior":    page - 1,
		"PagProxima":     page + 1,
		"ErrorMessage":   errMsg,
		"EditingID":      editingID,
	}
}

func getResponseMessage(out port.OutDTO) string {
	if m, ok := out.(interface{ GetMessage() string }); ok {
		return m.GetMessage()
	}
	return "Operação não concluída"
}

func (h *HandlerHtml) renderErrorMessage(w http.ResponseWriter, msg, tableEndpoint string) {
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf(`
	<div class="linha-box bg-rose-50 border border-rose-300 text-rose-800 p-4 rounded-xl shadow-md mb-4 flex items-center justify-between gap-4">
		<div class="flex items-center gap-3">
			<span class="text-xl">⚠️</span>
			<div class="font-bold text-sm text-rose-900">%s</div>
		</div>
		<button type="button" onclick="this.closest('.linha-box').remove()" class="text-rose-500 hover:text-rose-700 font-bold text-xs px-2 py-1 bg-rose-100 hover:bg-rose-200 rounded transition cursor-pointer">✕ Fechar</button>
	</div>`, template.HTMLEscapeString(msg))
	w.Write([]byte(html))
}
