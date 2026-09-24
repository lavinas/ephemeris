package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"signup/internal/domain"
)

func setupTestRouter() (*http.ServeMux, *memoryRepo, *mockLogger) {
	repo := newMemoryRepo()
	logger := newMockLogger()
	pub := &mockPublisher{}
	router := NewAPIRoutes(repo, logger, pub)
	return router, repo, logger
}

func setupTestRouterWithPublisher() (*http.ServeMux, *memoryRepo, *mockLogger, *mockPublisher) {
	repo := newMemoryRepo()
	logger := newMockLogger()
	pub := &mockPublisher{}
	router := NewAPIRoutes(repo, logger, pub)
	return router, repo, logger, pub
}

func executeRequest(router *http.ServeMux, method, path string, body []byte) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// -----------------------------------------------------------------------------
// /customer/create Tests
// -----------------------------------------------------------------------------

func TestCustomerCreate_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/customer/create", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestCustomerCreate_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPost, "/customer/create", []byte("invalid-json{"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestCustomerCreate_BadRequest_ValidationFailed(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Missing required fields
	payload := []byte(`{"vendor":"acme"}`)
	rec := executeRequest(router, http.MethodPost, "/customer/create", payload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for incomplete payload, got %d", rec.Code)
	}

	// Vendor does not exist
	payloadInvalidVendor := []byte(`{
		"vendor": "nonexistent",
		"name": "Client Name",
		"nickname": "clientone",
		"document": "11144477735",
		"email": "client@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec2 := executeRequest(router, http.MethodPost, "/customer/create", payloadInvalidVendor)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for non-existent vendor, got %d", rec2.Code)
	}
}

func TestCustomerCreate_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"name": "Client One",
		"nickname": "clientone",
		"document": "11144477735",
		"email": "clientone@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec := executeRequest(router, http.MethodPost, "/customer/create", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %v", response["status"])
	}

	// Verify customer saved in repo
	if len(repo.Customers) != 1 {
		t.Fatalf("expected 1 customer in repo, found %d", len(repo.Customers))
	}
	var savedCustomer *domain.Customer
	for _, c := range repo.Customers {
		savedCustomer = c
		break
	}
	if savedCustomer.Nickname != "clientone" {
		t.Errorf("expected nickname 'clientone', got %s", savedCustomer.Nickname)
	}
	if savedCustomer.VendorID != 1 {
		t.Errorf("expected vendor ID 1, got %d", savedCustomer.VendorID)
	}
}

func TestCustomerCreate_DuplicateNickname(t *testing.T) {
	router, _, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"name": "Client One",
		"nickname": "clientone",
		"document": "11144477735",
		"email": "clientone@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec1 := executeRequest(router, http.MethodPost, "/customer/create", payload)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first creation expected 200, got %d", rec1.Code)
	}

	// Try creating again with duplicate nickname
	rec2 := executeRequest(router, http.MethodPost, "/customer/create", payload)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on duplicate nickname, got %d", rec2.Code)
	}
}

// -----------------------------------------------------------------------------
// /customer/update Tests
// -----------------------------------------------------------------------------

func TestCustomerUpdate_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/customer/update", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestCustomerUpdate_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPatch, "/customer/update", []byte("invalid-json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestCustomerUpdate_CustomerNotFound(t *testing.T) {
	router, _, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"nickname": "nonexistent",
		"name": "New Name"
	}`)
	rec := executeRequest(router, http.MethodPatch, "/customer/update", payload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for nonexistent customer, got %d", rec.Code)
	}
}

func TestCustomerUpdate_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	// Seed existing customer
	doc := "11144477735"
	email := "old@example.com"
	phone := "+5511999998888"
	status := 1
	existing := &domain.Customer{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         10,
		VendorID:   1,
		Name:       "Old Name",
		Nickname:   "existingclient",
		Document:   &doc,
		Email:      &email,
		Whatsapp:   &phone,
		Status:     &status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Customers[existing.ID] = existing

	payload := []byte(`{
		"vendor": "acme",
		"nickname": "existingclient",
		"name": "New Name Updated",
		"whatsapp": "+5511999997777"
	}`)
	rec := executeRequest(router, http.MethodPatch, "/customer/update", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	updated := repo.Customers[existing.ID]
	if updated.Name != "New Name Updated" {
		t.Errorf("expected name 'New Name Updated', got '%s'", updated.Name)
	}
	if *updated.Whatsapp != "+5511999997777" {
		t.Errorf("expected whatsapp '+5511999997777', got '%s'", *updated.Whatsapp)
	}
}

// -----------------------------------------------------------------------------
// /customer/list Tests
// -----------------------------------------------------------------------------

func TestCustomerList_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPost, "/customer/list", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestCustomerList_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/customer/list", []byte("invalid-json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestCustomerList_BadRequest_ValidationFailed(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Page <= 0
	payload := []byte(`{"vendor":"acme","page":0,"page_size":10}`)
	rec := executeRequest(router, http.MethodGet, "/customer/list", payload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for page <= 0, got %d", rec.Code)
	}

	// Missing vendor
	payloadNoVendor := []byte(`{"page":1,"page_size":10}`)
	rec2 := executeRequest(router, http.MethodGet, "/customer/list", payloadNoVendor)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing vendor, got %d", rec2.Code)
	}
}

func TestCustomerList_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	// Seed customers
	doc1 := "11144477735"
	status1 := 1
	repo.Customers[1] = &domain.Customer{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         1,
		VendorID:   1,
		Name:       "Customer Alpha",
		Nickname:   "alpha",
		Document:   &doc1,
		Status:     &status1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	doc2 := "52998224725"
	status2 := 1
	repo.Customers[2] = &domain.Customer{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         2,
		VendorID:   1,
		Name:       "Customer Beta",
		Nickname:   "beta",
		Document:   &doc2,
		Status:     &status2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	payload := []byte(`{
		"vendor": "acme",
		"page": 1,
		"page_size": 10
	}`)
	rec := executeRequest(router, http.MethodGet, "/customer/list", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %v", response["status"])
	}

	customersRaw, ok := response["customers"].([]interface{})
	if !ok {
		t.Fatalf("expected customers array in response, got %v", response["customers"])
	}
	if len(customersRaw) != 2 {
		t.Errorf("expected 2 customers listed, got %d", len(customersRaw))
	}
}

// -----------------------------------------------------------------------------
// /user/create Tests
// -----------------------------------------------------------------------------

func TestUserCreate_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/user/create", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestUserCreate_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPost, "/user/create", []byte("invalid-json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestUserCreate_BadRequest_ValidationFailed(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Missing fields
	payload := []byte(`{"vendor":"acme"}`)
	rec := executeRequest(router, http.MethodPost, "/user/create", payload)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for incomplete user, got %d", rec.Code)
	}

	// Invalid password (< 8 chars)
	payloadShortPass := []byte(`{
		"vendor": "acme",
		"name": "John Doe",
		"username": "johndoe",
		"password": "123",
		"email": "john@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec2 := executeRequest(router, http.MethodPost, "/user/create", payloadShortPass)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for short password, got %d", rec2.Code)
	}
}

func TestUserCreate_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"name": "John Doe",
		"username": "johndoe",
		"password": "Password@123",
		"email": "john@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec := executeRequest(router, http.MethodPost, "/user/create", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %v", response["status"])
	}

	// Verify user saved in repo
	if len(repo.Users) != 1 {
		t.Fatalf("expected 1 user in repo, found %d", len(repo.Users))
	}
	var savedUser *domain.User
	for _, u := range repo.Users {
		savedUser = u
		break
	}
	if savedUser.Username != "johndoe" {
		t.Errorf("expected username 'johndoe', got %s", savedUser.Username)
	}
	if savedUser.PassHash == "Password@123" || !strings.HasPrefix(savedUser.PassHash, "$2a$") {
		t.Errorf("expected password to be hashed with bcrypt, got '%s'", savedUser.PassHash)
	}
}

func TestUserCreate_DuplicateUsername(t *testing.T) {
	router, _, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"name": "John Doe",
		"username": "johndoe",
		"password": "Password@123",
		"email": "john1@example.com",
		"whatsapp": "+5511999998888"
	}`)
	rec1 := executeRequest(router, http.MethodPost, "/user/create", payload)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first creation expected 200, got %d", rec1.Code)
	}

	// Second attempt with same username
	rec2 := executeRequest(router, http.MethodPost, "/user/create", payload)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on duplicate username, got %d", rec2.Code)
	}
}

// -----------------------------------------------------------------------------
// /user/update Tests
// -----------------------------------------------------------------------------

func TestUserUpdate_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/user/update", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestUserUpdate_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPatch, "/user/update", []byte("invalid-json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestUserUpdate_UserNotFound(t *testing.T) {
	router, _, _ := setupTestRouter()

	payload := []byte(`{
		"vendor": "acme",
		"username": "nonexistent",
		"name": "New Name"
	}`)
	rec := executeRequest(router, http.MethodPatch, "/user/update", payload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for nonexistent user, got %d", rec.Code)
	}
}

func TestUserUpdate_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	// Seed existing user
	email := "johndoe@example.com"
	phone := "+5511999998888"
	status := 1
	existing := &domain.User{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         20,
		VendorID:   1,
		Name:       "John Doe",
		Username:   "johndoe",
		PassHash:   "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
		Email:      &email,
		Whatsapp:   &phone,
		Status:     &status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Users[existing.ID] = existing

	payload := []byte(`{
		"vendor": "acme",
		"username": "johndoe",
		"name": "John Doe Senior"
	}`)
	rec := executeRequest(router, http.MethodPatch, "/user/update", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	updated := repo.Users[existing.ID]
	if updated.Name != "John Doe Senior" {
		t.Errorf("expected name 'John Doe Senior', got '%s'", updated.Name)
	}
}

// -----------------------------------------------------------------------------
// /user/list Tests
// -----------------------------------------------------------------------------

func TestUserList_MethodNotAllowed(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodPost, "/user/list", nil)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", rec.Code)
	}
}

func TestUserList_BadRequest_InvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter()
	rec := executeRequest(router, http.MethodGet, "/user/list", []byte("invalid-json"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestUserList_BadRequest_ValidationFailed(t *testing.T) {
	router, _, _ := setupTestRouter()

	// Page <= 0
	payload := []byte(`{"vendor":"acme","page":0,"page_size":10}`)
	rec := executeRequest(router, http.MethodGet, "/user/list", payload)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for page <= 0, got %d", rec.Code)
	}

	// Missing vendor
	payloadNoVendor := []byte(`{"page":1,"page_size":10}`)
	rec2 := executeRequest(router, http.MethodGet, "/user/list", payloadNoVendor)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing vendor, got %d", rec2.Code)
	}
}

func TestUserList_Success(t *testing.T) {
	router, repo, _ := setupTestRouter()

	// Seed users
	status1 := 1
	repo.Users[1] = &domain.User{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         1,
		VendorID:   1,
		Name:       "User Alpha",
		Username:   "alpha",
		PassHash:   "hash1",
		Status:     &status1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	status2 := 1
	repo.Users[2] = &domain.User{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         2,
		VendorID:   1,
		Name:       "User Beta",
		Username:   "beta",
		PassHash:   "hash2",
		Status:     &status2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	payload := []byte(`{
		"vendor": "acme",
		"page": 1,
		"page_size": 10
	}`)
	rec := executeRequest(router, http.MethodGet, "/user/list", payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response JSON: %v", err)
	}
	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %v", response["status"])
	}

	usersRaw, ok := response["users"].([]interface{})
	if !ok {
		t.Fatalf("expected users array in response, got %v", response["users"])
	}
	if len(usersRaw) != 2 {
		t.Errorf("expected 2 users listed, got %d", len(usersRaw))
	}
}

// -----------------------------------------------------------------------------
// Routes Registration Test
// -----------------------------------------------------------------------------

func TestAPIRoutesRegistration(t *testing.T) {
	router, _, _ := setupTestRouter()

	routes := []struct {
		method         string
		path           string
		expectedNot404 bool
	}{
		{http.MethodPost, "/customer/create", true},
		{http.MethodPost, "/customer/update", true},
		{http.MethodGet, "/customer/list", true},
		{http.MethodPost, "/user/create", true},
		{http.MethodPost, "/user/update", true},
		{http.MethodGet, "/user/list", true},
		{http.MethodGet, "/ping", true},
		{http.MethodGet, "/unregistered/path", false},
	}

	for _, tc := range routes {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			rec := executeRequest(router, tc.method, tc.path, nil)
			if tc.expectedNot404 && rec.Code == http.StatusNotFound {
				t.Errorf("route %s was expected to exist but returned 404", tc.path)
			}
			if !tc.expectedNot404 && rec.Code != http.StatusNotFound {
				t.Errorf("route %s was expected to return 404, got %d", tc.path, rec.Code)
			}
		})
	}
}

func TestCustomerCreate_EmitsEvent(t *testing.T) {
	router, repo, _, pub := setupTestRouterWithPublisher()

	// Seed vendor
	repo.Vendors[1] = &domain.Vendor{
		ID:       1,
		Nickname: "acme",
	}

	payload := []byte(`{
		"vendor": "acme",
		"name": "Sync Client",
		"nickname": "syncclient",
		"document": "11144477735",
		"email": "sync@example.com",
		"whatsapp": "+5511999998888"
	}`)

	rec := executeRequest(router, http.MethodPost, "/customer/create", payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(pub.PublishedCreated) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.PublishedCreated))
	}

	event := pub.PublishedCreated[0]
	if event.Nickname != "syncclient" {
		t.Errorf("expected nickname 'syncclient', got '%s'", event.Nickname)
	}
	if event.Name != "Sync Client" {
		t.Errorf("expected name 'Sync Client', got '%s'", event.Name)
	}
}

func TestCustomerUpdate_EmitsEvent(t *testing.T) {
	router, repo, _, pub := setupTestRouterWithPublisher()

	repo.Vendors[1] = &domain.Vendor{
		ID:       1,
		Nickname: "acme",
	}

	doc := "11144477735"
	email := "sync@example.com"
	phone := "+5511999998888"
	status := 1
	existing := &domain.Customer{
		DomainBase: domain.DomainBase{Repo: repo},
		ID:         10,
		VendorID:   1,
		Name:       "Sync Client",
		Nickname:   "syncclient",
		Document:   &doc,
		Email:      &email,
		Whatsapp:   &phone,
		Status:     &status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	repo.Customers[existing.ID] = existing

	payload := []byte(`{
		"vendor": "acme",
		"nickname": "syncclient",
		"name": "Sync Client Updated",
		"whatsapp": "+5511988887777"
	}`)

	rec := executeRequest(router, http.MethodPatch, "/customer/update", payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(pub.PublishedUpdated) != 1 {
		t.Fatalf("expected 1 update event published, got %d", len(pub.PublishedUpdated))
	}

	event := pub.PublishedUpdated[0]
	if event.Nickname != "syncclient" {
		t.Errorf("expected nickname 'syncclient', got '%s'", event.Nickname)
	}
	if event.Name != "Sync Client Updated" {
		t.Errorf("expected name 'Sync Client Updated', got '%s'", event.Name)
	}
	if event.Whatsapp == nil || *event.Whatsapp != "+5511988887777" {
		t.Errorf("expected updated whatsapp, got %v", event.Whatsapp)
	}
}
