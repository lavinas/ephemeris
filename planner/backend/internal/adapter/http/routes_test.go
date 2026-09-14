package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

type dummyLogger struct{}

func (d *dummyLogger) IPrintf(level int, format string, v ...interface{}) {}
func (d *dummyLogger) Close()                                             {}

type dummyRepo struct{}

func (d *dummyRepo) BeginTransaction() error      { return nil }
func (d *dummyRepo) CommitTransaction() error     { return nil }
func (d *dummyRepo) RollbackTransaction() error   { return nil }
func (d *dummyRepo) Save(model interface{}) error { return nil }
func (d *dummyRepo) Find(page, pagesize int, conditions map[string]interface{}, orderBy ...string) ([]interface{}, error) {
	return nil, nil
}
func (d *dummyRepo) FindCount(conditions map[string]interface{}) (int64, error) { return 0, nil }
func (d *dummyRepo) FindGroup(conditions map[string]interface{}, groupField string) ([]map[string]interface{}, error) {
	return nil, nil
}
func (d *dummyRepo) Close() error { return nil }

func TestRootEndpointServesStaticIndex(t *testing.T) {
	// Change working directory to planner/backend for test to resolve web/static
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get working dir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir("../../.."); err != nil {
		t.Fatalf("could not chdir to planner/backend: %v", err)
	}

	templateData := []byte(`{{define "index"}}<html><body>Sessions</body></html>{{end}}`)
	routes, err := NewRoutes(&dummyRepo{}, &dummyLogger{}, templateData)
	if err != nil {
		t.Fatalf("failed to create routes: %v", err)
	}

	// 1. Test GET /
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	routes.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 on /, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "subframe-conteudo") {
		t.Errorf("expected index.html to contain 'subframe-conteudo'")
	}
	if !strings.Contains(body, `/html/sessoes`) {
		t.Errorf("expected index.html to reference '/html/sessoes'")
	}

	// 2. Test unknown route returns 404
	req404 := httptest.NewRequest(http.MethodGet, "/nao-existe", nil)
	rec404 := httptest.NewRecorder()
	routes.ServeHTTP(rec404, req404)

	if rec404.Code != http.StatusNotFound {
		t.Errorf("expected status 404 on /nao-existe, got %d", rec404.Code)
	}

	// 3. Test GET /html/ping
	reqPing := httptest.NewRequest(http.MethodGet, "/html/ping", nil)
	recPing := httptest.NewRecorder()
	routes.ServeHTTP(recPing, reqPing)

	if recPing.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/ping, got %d", recPing.Code)
	}

	// 4. Test GET /static/logo.png
	reqLogo := httptest.NewRequest(http.MethodGet, "/static/logo.png", nil)
	recLogo := httptest.NewRecorder()
	routes.ServeHTTP(recLogo, reqLogo)

	if recLogo.Code != http.StatusOK {
		t.Errorf("expected status 200 on /static/logo.png, got %d", recLogo.Code)
	}

	// 5. Test GET /html/sessoes (the subframe target)
	reqSessoes := httptest.NewRequest(http.MethodGet, "/html/sessoes", nil)
	recSessoes := httptest.NewRecorder()
	routes.ServeHTTP(recSessoes, reqSessoes)

	if recSessoes.Code != http.StatusOK {
		t.Errorf("expected status 200 on /html/sessoes, got %d", recSessoes.Code)
	}
	if !strings.Contains(recSessoes.Body.String(), "Sessions") {
		t.Errorf("expected /html/sessoes response to contain 'Sessions'")
	}
}
