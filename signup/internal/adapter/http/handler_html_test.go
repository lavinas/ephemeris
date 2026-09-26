package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"signup/internal/domain"
)

func loadTestTemplates(t *testing.T) []byte {
	t.Helper()
	var combined []byte
	paths := []string{
		"../../../web/templates/customers.html",
		"../../../web/templates/users.html",
		"../../web/templates/customers.html",
		"../../web/templates/users.html",
		"web/templates/customers.html",
		"web/templates/users.html",
		"/app/web/templates/customers.html",
		"/app/web/templates/users.html",
	}
	for _, p := range paths {
		if content, err := os.ReadFile(p); err == nil {
			combined = append(combined, content...)
			combined = append(combined, []byte("\n")...)
		}
	}
	if len(combined) == 0 {
		t.Skip("HTML templates not found on disk for unit test")
	}
	return combined
}

func setupHTMLTest(t *testing.T) (*HandlerHtml, *memoryRepo, *mockLogger, *mockPublisher) {
	t.Helper()
	repo := newMemoryRepo()
	logger := newMockLogger()
	publisher := &mockPublisher{}
	tmplData := loadTestTemplates(t)

	handler, err := NewHandlerHtml(repo, logger, publisher, tmplData)
	if err != nil {
		t.Fatalf("failed to create HandlerHtml: %v", err)
	}

	// Seed sample customer
	cust := domain.NewCustomer(
		repo,
		1,
		strPtr("Cliente Teste"),
		strPtr("cliente_teste"),
		strPtr("11144477735"),
		strPtr("cliente@teste.com"),
		strPtr("+5511999998888"),
	)
	cust.ID = 10
	repo.Customers[cust.ID] = cust

	// Seed sample user
	usr := domain.NewUser(
		repo,
		1,
		"Usuario Teste",
		"usuario_teste",
		"hashpass123",
		strPtr("usuario@teste.com"),
		strPtr("+5511999997777"),
	)
	usr.ID = 20
	repo.Users[usr.ID] = usr

	return handler, repo, logger, publisher
}

func strPtr(s string) *string {
	return &s
}

func TestHandlerHtml_Ping(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	req := httptest.NewRequest(http.MethodGet, "/html/ping", nil)
	rec := httptest.NewRecorder()

	handler.Ping(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Errorf("expected body 'pong', got '%s'", rec.Body.String())
	}
}

func TestHandlerHtml_Customers_List(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Standard request
	req := httptest.NewRequest(http.MethodGet, "/html/customers", nil)
	rec := httptest.NewRecorder()

	handler.Customers(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "cliente_teste") {
		t.Errorf("expected body to contain 'cliente_teste', got: %s", body)
	}
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("expected standard request to include full html doctype")
	}

	// HTMX request
	reqHTMX := httptest.NewRequest(http.MethodGet, "/html/customers", nil)
	reqHTMX.Header.Set("HX-Request", "true")
	recHTMX := httptest.NewRecorder()

	handler.Customers(recHTMX, reqHTMX)
	if recHTMX.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recHTMX.Code)
	}
	bodyHTMX := recHTMX.Body.String()
	if !strings.Contains(bodyHTMX, "cliente_teste") {
		t.Errorf("expected HTMX body to contain 'cliente_teste'")
	}
}

func TestHandlerHtml_Customers_CreateAndSave(t *testing.T) {
	handler, repo, _, _ := setupHTMLTest(t)

	// Open create form
	reqForm := httptest.NewRequest(http.MethodGet, "/html/customers/novo", nil)
	recForm := httptest.NewRecorder()
	handler.CustomersCreate(recForm, reqForm)
	if recForm.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recForm.Code)
	}
	if !strings.Contains(recForm.Body.String(), "formulario-add") {
		t.Errorf("expected creation form in body")
	}

	// Save new customer
	form := url.Values{}
	form.Set("add_vendor", "acme")
	form.Set("add_nickname", "novocliente")
	form.Set("add_name", "Novo Cliente da Silva")
	form.Set("add_document", "11144477735")
	form.Set("add_whatsapp", "+5511988887777")
	form.Set("add_email", "novo@cliente.com")
	form.Set("page", "1")

	reqSave := httptest.NewRequest(http.MethodPost, "/html/customers/salvar", strings.NewReader(form.Encode()))
	reqSave.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recSave := httptest.NewRecorder()

	handler.CustomersSave(recSave, reqSave)
	if recSave.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recSave.Code, recSave.Body.String())
	}
	body := recSave.Body.String()
	if !strings.Contains(body, "novocliente") {
		t.Errorf("expected table to contain 'novocliente', got: %s", body)
	}

	// Check customer in repo
	var found bool
	for _, c := range repo.Customers {
		if c.Nickname == "novocliente" {
			found = true
			if c.Status == nil || *c.Status != 1 {
				t.Errorf("expected new customer status to be 1, got %v", c.Status)
			}
		}
	}
	if !found {
		t.Errorf("new customer not found in repo")
	}
}

func TestHandlerHtml_Customers_EditUpdateDelete(t *testing.T) {
	handler, repo, _, _ := setupHTMLTest(t)

	// 1. Delete warning
	reqWarn := httptest.NewRequest(http.MethodGet, "/html/customers/deletar-aviso?id=10", nil)
	recWarn := httptest.NewRecorder()
	handler.CustomersDeleteWarning(recWarn, reqWarn)
	if recWarn.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recWarn.Code)
	}
	if !strings.Contains(recWarn.Body.String(), "aviso-deletar") {
		t.Errorf("expected warning in body")
	}

	// 2. Cancel edit
	reqCancel := httptest.NewRequest(http.MethodGet, "/html/customers/cancelar-edicao?id=10", nil)
	recCancel := httptest.NewRecorder()
	handler.CustomersCancelEdit(recCancel, reqCancel)
	if recCancel.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recCancel.Code)
	}
	if !strings.Contains(recCancel.Body.String(), "cliente_teste") {
		t.Errorf("expected row in body")
	}

	// 3. Edit row
	reqEdit := httptest.NewRequest(http.MethodGet, "/html/customers/editar?id=10", nil)
	recEdit := httptest.NewRecorder()
	handler.CustomersEdit(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recEdit.Code)
	}
	if !strings.Contains(recEdit.Body.String(), "formulario-edit") {
		t.Errorf("expected edit form in body")
	}

	// 4. Update
	form := url.Values{}
	form.Set("edit_name", "Cliente Teste Alterado")
	form.Set("edit_document", "11144477735")
	form.Set("edit_email", "cliente_alt@teste.com")
	form.Set("edit_whatsapp", "+5511999998888")
	form.Set("edit_status", "1")
	form.Set("page", "1")

	reqUpdate := httptest.NewRequest(http.MethodPost, "/html/customers/atualizar?id=10", strings.NewReader(form.Encode()))
	reqUpdate.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recUpdate := httptest.NewRecorder()
	handler.CustomersUpdate(recUpdate, reqUpdate)
	if recUpdate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recUpdate.Code, recUpdate.Body.String())
	}
	if !strings.Contains(recUpdate.Body.String(), "Cliente Teste Alterado") {
		t.Errorf("expected updated name in table")
	}

	// 5. Delete (inactivate with status = 0)
	formDel := url.Values{}
	formDel.Set("page", "1")
	formDel.Set("vendor", "acme")
	reqDel := httptest.NewRequest(http.MethodPost, "/html/customers/deletar?id=10", strings.NewReader(formDel.Encode()))
	reqDel.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recDel := httptest.NewRecorder()
	handler.CustomersDelete(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recDel.Code)
	}

	// Verify status in repo is 0
	cust := repo.Customers[10]
	if cust.Status == nil || *cust.Status != 0 {
		t.Errorf("expected customer status to be 0 after delete, got %v", cust.Status)
	}
}

func TestHandlerHtml_Users_List(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Standard request
	req := httptest.NewRequest(http.MethodGet, "/html/users", nil)
	rec := httptest.NewRecorder()

	handler.Users(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "usuario_teste") {
		t.Errorf("expected body to contain 'usuario_teste', got: %s", body)
	}
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Errorf("expected standard request to include full html doctype")
	}

	// HTMX request
	reqHTMX := httptest.NewRequest(http.MethodGet, "/html/users", nil)
	reqHTMX.Header.Set("HX-Request", "true")
	recHTMX := httptest.NewRecorder()

	handler.Users(recHTMX, reqHTMX)
	if recHTMX.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recHTMX.Code)
	}
	bodyHTMX := recHTMX.Body.String()
	if !strings.Contains(bodyHTMX, "usuario_teste") {
		t.Errorf("expected HTMX body to contain 'usuario_teste'")
	}
}

func TestHandlerHtml_Users_CreateAndSave(t *testing.T) {
	handler, repo, _, _ := setupHTMLTest(t)

	// Open create form
	reqForm := httptest.NewRequest(http.MethodGet, "/html/users/novo", nil)
	recForm := httptest.NewRecorder()
	handler.UsersCreate(recForm, reqForm)
	if recForm.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recForm.Code)
	}
	if !strings.Contains(recForm.Body.String(), "formulario-add") {
		t.Errorf("expected creation form in body")
	}

	// Save new user
	form := url.Values{}
	form.Set("add_vendor", "acme")
	form.Set("add_username", "novousuario")
	form.Set("add_name", "Novo Usuario Pereira")
	form.Set("add_password", "senhasegura123")
	form.Set("add_whatsapp", "+5511977776666")
	form.Set("add_email", "novo@usuario.com")
	form.Set("page", "1")

	reqSave := httptest.NewRequest(http.MethodPost, "/html/users/salvar", strings.NewReader(form.Encode()))
	reqSave.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recSave := httptest.NewRecorder()

	handler.UsersSave(recSave, reqSave)
	if recSave.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recSave.Code, recSave.Body.String())
	}
	body := recSave.Body.String()
	if !strings.Contains(body, "novousuario") {
		t.Errorf("expected table to contain 'novousuario', got: %s", body)
	}

	// Check user in repo
	var found bool
	for _, u := range repo.Users {
		if u.Username == "novousuario" {
			found = true
			if u.Status == nil || *u.Status != 1 {
				t.Errorf("expected new user status to be 1, got %v", u.Status)
			}
		}
	}
	if !found {
		t.Errorf("new user not found in repo")
	}
}

func TestHandlerHtml_Users_EditUpdateDelete(t *testing.T) {
	handler, repo, _, _ := setupHTMLTest(t)

	// 1. Delete warning
	reqWarn := httptest.NewRequest(http.MethodGet, "/html/users/deletar-aviso?id=20", nil)
	recWarn := httptest.NewRecorder()
	handler.UsersDeleteWarning(recWarn, reqWarn)
	if recWarn.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recWarn.Code)
	}
	if !strings.Contains(recWarn.Body.String(), "aviso-deletar") {
		t.Errorf("expected warning in body")
	}

	// 2. Cancel edit
	reqCancel := httptest.NewRequest(http.MethodGet, "/html/users/cancelar-edicao?id=20", nil)
	recCancel := httptest.NewRecorder()
	handler.UsersCancelEdit(recCancel, reqCancel)
	if recCancel.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recCancel.Code)
	}
	if !strings.Contains(recCancel.Body.String(), "usuario_teste") {
		t.Errorf("expected row in body")
	}

	// 3. Edit row
	reqEdit := httptest.NewRequest(http.MethodGet, "/html/users/editar?id=20", nil)
	recEdit := httptest.NewRecorder()
	handler.UsersEdit(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recEdit.Code)
	}
	if !strings.Contains(recEdit.Body.String(), "formulario-edit") {
		t.Errorf("expected edit form in body")
	}

	// 4. Update
	form := url.Values{}
	form.Set("edit_name", "Usuario Teste Alterado")
	form.Set("edit_email", "usuario_alt@teste.com")
	form.Set("edit_whatsapp", "+5511999997777")
	form.Set("edit_status", "1")
	form.Set("page", "1")

	reqUpdate := httptest.NewRequest(http.MethodPost, "/html/users/atualizar?id=20", strings.NewReader(form.Encode()))
	reqUpdate.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recUpdate := httptest.NewRecorder()
	handler.UsersUpdate(recUpdate, reqUpdate)
	if recUpdate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, recUpdate.Code, recUpdate.Body.String())
	}
	if !strings.Contains(recUpdate.Body.String(), "Usuario Teste Alterado") {
		t.Errorf("expected updated name in table")
	}

	// 5. Delete (inactivate with status = 0)
	formDel := url.Values{}
	formDel.Set("page", "1")
	formDel.Set("vendor", "acme")
	reqDel := httptest.NewRequest(http.MethodPost, "/html/users/deletar?id=20", strings.NewReader(formDel.Encode()))
	reqDel.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recDel := httptest.NewRecorder()
	handler.UsersDelete(recDel, reqDel)
	if recDel.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recDel.Code)
	}

	// Verify status in repo is 0
	usr := repo.Users[20]
	if usr.Status == nil || *usr.Status != 0 {
		t.Errorf("expected user status to be 0 after delete, got %v", usr.Status)
	}
}

func TestNewRoutes_WithHTML(t *testing.T) {
	repo := newMemoryRepo()
	logger := newMockLogger()
	publisher := &mockPublisher{}
	tmplData := loadTestTemplates(t)

	mux, err := NewRoutes(repo, logger, publisher, tmplData)
	if err != nil {
		t.Fatalf("NewRoutes returned error: %v", err)
	}
	if mux == nil {
		t.Fatal("expected non-nil mux")
	}

	// Test GET /html/customers route
	req := httptest.NewRequest(http.MethodGet, "/html/customers", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/customers, got %d", rec.Code)
	}

	// Test GET /html/users route
	reqUsers := httptest.NewRequest(http.MethodGet, "/html/users", nil)
	recUsers := httptest.NewRecorder()
	mux.ServeHTTP(recUsers, reqUsers)
	if recUsers.Code != http.StatusOK {
		t.Errorf("expected 200 on /html/users, got %d", recUsers.Code)
	}

	// Test GET /api/ping route still works
	reqPing := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	recPing := httptest.NewRecorder()
	mux.ServeHTTP(recPing, reqPing)
	if recPing.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/ping, got %d", recPing.Code)
	}
}

func TestHandlerHtml_Customer_SaveError_PreservesFormAndShowsBanner(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Save customer with invalid data (empty name)
	form := url.Values{}
	form.Set("add_vendor", "acme")
	form.Set("add_nickname", "invalido")
	form.Set("add_name", "")
	form.Set("add_document", "11144477735")
	form.Set("add_whatsapp", "+5511988887777")
	form.Set("add_email", "invalido@email.com")

	req := httptest.NewRequest(http.MethodPost, "/html/customers/salvar", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.CustomersSave(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	// 1. Error banner at top of results table
	if !strings.Contains(body, "Erro ao salvar cliente") {
		t.Errorf("expected error banner in results table, got: %s", body)
	}

	// 2. Preserved form values in OOB swap container
	if !strings.Contains(body, `id="formulario-cadastro-container" hx-swap-oob="innerHTML"`) {
		t.Errorf("expected OOB swap for form container")
	}
	if !strings.Contains(body, `value="invalido"`) {
		t.Errorf("expected form to preserve 'invalido' nickname")
	}
	if !strings.Contains(body, `value="invalido@email.com"`) {
		t.Errorf("expected form to preserve email")
	}

	// 3. Existing table content still rendered
	if !strings.Contains(body, "cliente_teste") {
		t.Errorf("expected table to still contain existing customers")
	}
}

func TestHandlerHtml_Customer_UpdateError_PreservesEditRowAndShowsBanner(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Update customer with invalid data (invalid status)
	form := url.Values{}
	form.Set("edit_name", "Nome Editado Sem Salvar")
	form.Set("edit_document", "11144477735")
	form.Set("edit_email", "editado@teste.com")
	form.Set("edit_whatsapp", "+5511999998888")
	form.Set("edit_status", "99") // invalid status
	form.Set("page", "1")

	req := httptest.NewRequest(http.MethodPost, "/html/customers/atualizar?id=10", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.CustomersUpdate(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	// 1. Error banner at top of results table
	if !strings.Contains(body, "Erro ao atualizar cliente") {
		t.Errorf("expected error banner in results table, got: %s", body)
	}

	// 2. Customer row remains in edit mode with typed values preserved
	if !strings.Contains(body, "Nome Editado Sem Salvar") {
		t.Errorf("expected edit row to preserve edited name, got: %s", body)
	}
	if !strings.Contains(body, "editado@teste.com") {
		t.Errorf("expected edit row to preserve edited email")
	}
	if !strings.Contains(body, "formulario-edit") {
		t.Errorf("expected row to still be in edit mode (formulario-edit)")
	}
}

func TestHandlerHtml_User_SaveError_PreservesFormAndShowsBanner(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Save user with invalid password (too short)
	form := url.Values{}
	form.Set("add_vendor", "acme")
	form.Set("add_username", "novousuarioinvalido")
	form.Set("add_name", "Usuario Invalido")
	form.Set("add_password", "123") // too short
	form.Set("add_whatsapp", "+5511977776666")
	form.Set("add_email", "invalido@usuario.com")

	req := httptest.NewRequest(http.MethodPost, "/html/users/salvar", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.UsersSave(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	// 1. Error banner at top of results table
	if !strings.Contains(body, "Erro ao salvar usuário") {
		t.Errorf("expected error banner in results table, got: %s", body)
	}

	// 2. Preserved form values in OOB swap container
	if !strings.Contains(body, `id="formulario-cadastro-container" hx-swap-oob="innerHTML"`) {
		t.Errorf("expected OOB swap for form container")
	}
	if !strings.Contains(body, `value="novousuarioinvalido"`) {
		t.Errorf("expected form to preserve username")
	}
	if !strings.Contains(body, `value="Usuario Invalido"`) {
		t.Errorf("expected form to preserve name")
	}
	if !strings.Contains(body, `value="invalido@usuario.com"`) {
		t.Errorf("expected form to preserve email")
	}

	// 3. Existing table content still rendered
	if !strings.Contains(body, "usuario_teste") {
		t.Errorf("expected table to still contain existing users")
	}
}

func TestHandlerHtml_User_UpdateError_PreservesEditRowAndShowsBanner(t *testing.T) {
	handler, _, _, _ := setupHTMLTest(t)

	// Update user with invalid status
	form := url.Values{}
	form.Set("edit_name", "Usuario Editado Sem Salvar")
	form.Set("edit_email", "editado@usuario.com")
	form.Set("edit_whatsapp", "+5511999997777")
	form.Set("edit_status", "99") // invalid status
	form.Set("page", "1")

	req := httptest.NewRequest(http.MethodPost, "/html/users/atualizar?id=20", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.UsersUpdate(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()

	// 1. Error banner at top of results table
	if !strings.Contains(body, "Erro ao atualizar usuário") {
		t.Errorf("expected error banner in results table, got: %s", body)
	}

	// 2. User row remains in edit mode with typed values preserved
	if !strings.Contains(body, "Usuario Editado Sem Salvar") {
		t.Errorf("expected edit row to preserve edited name, got: %s", body)
	}
	if !strings.Contains(body, "editado@usuario.com") {
		t.Errorf("expected edit row to preserve edited email")
	}
	if !strings.Contains(body, "formulario-edit") {
		t.Errorf("expected row to still be in edit mode (formulario-edit)")
	}
}
