package main_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	db "shorty/internal/db/sqlc"
	httpapi "shorty/internal/http"
)

type linkResp struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

type errResp struct {
	Error string `json:"error"`
}

var (
	testSQL  *sql.DB
	testPool *pgxpool.Pool
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		panic("DATABASE_URL is required for tests")
	}

	var err error

	testSQL, err = sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(testSQL, "db/migrations"); err != nil {
		panic(err)
	}

	testPool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}

	code := m.Run()

	testPool.Close()
	_ = testSQL.Close()

	os.Exit(code)
}

func truncateLinks(t *testing.T) {
	t.Helper()

	_, err := testSQL.Exec(`TRUNCATE links RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
}

func newRouter(t *testing.T) http.Handler {
	t.Helper()

	q := db.New(testPool)
	return httpapi.NewRouter(q, "https://short.io")
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var r *http.Request
	if body == nil {
		r = httptest.NewRequest(method, path, nil)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		r = httptest.NewRequest(method, path, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to decode json: %v, body=%s", err, w.Body.String())
	}
	return v
}

func TestLinksCRUD(t *testing.T) {
	truncateLinks(t)
	h := newRouter(t)

	// POST
	w := doJSON(t, h, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/long-url",
		"short_name":   "exmpl",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("POST expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	created := decodeJSON[linkResp](t, w)
	if created.OriginalURL != "https://example.com/long-url" {
		t.Fatalf("unexpected original_url: %q", created.OriginalURL)
	}
	if created.ShortName != "exmpl" {
		t.Fatalf("unexpected short_name: %q", created.ShortName)
	}
	if created.ShortURL != "https://short.io/r/exmpl" {
		t.Fatalf("unexpected short_url: %q", created.ShortURL)
	}
	if created.ID <= 0 {
		t.Fatalf("unexpected id: %d", created.ID)
	}

	idPath := "/api/links/" + strconv.FormatInt(created.ID, 10)

	// GET /:id
	w = doJSON(t, h, http.MethodGet, idPath, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("GET expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	got := decodeJSON[linkResp](t, w)
	if got.ID != created.ID {
		t.Fatalf("unexpected id: %d", got.ID)
	}

	// PUT /:id
	w = doJSON(t, h, http.MethodPut, idPath, map[string]any{
		"original_url": "https://example.com/updated",
		"short_name":   "exmpl2",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("PUT expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	updated := decodeJSON[linkResp](t, w)
	if updated.OriginalURL != "https://example.com/updated" {
		t.Fatalf("unexpected original_url: %q", updated.OriginalURL)
	}
	if updated.ShortName != "exmpl2" {
		t.Fatalf("unexpected short_name: %q", updated.ShortName)
	}
	if updated.ShortURL != "https://short.io/r/exmpl2" {
		t.Fatalf("unexpected short_url: %q", updated.ShortURL)
	}

	// LIST
	w = doJSON(t, h, http.MethodGet, "/api/links", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("LIST expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	list := decodeJSON[[]linkResp](t, w)
	if len(list) != 1 {
		t.Fatalf("expected 1 link, got %d", len(list))
	}

	// DELETE
	w = doJSON(t, h, http.MethodDelete, idPath, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE expected 204, got %d, body=%s", w.Code, w.Body.String())
	}

	// GET after delete -> 404
	w = doJSON(t, h, http.MethodGet, idPath, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET after delete expected 404, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCreateGeneratesShortNameWhenMissing(t *testing.T) {
	truncateLinks(t)
	h := newRouter(t)

	w := doJSON(t, h, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/long-url",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	created := decodeJSON[linkResp](t, w)
	if created.ShortName == "" {
		t.Fatalf("expected generated short_name, got empty")
	}
	if created.ShortURL != "https://short.io/r/"+created.ShortName {
		t.Fatalf("unexpected short_url: %q", created.ShortURL)
	}
}

func TestShortNameConflictReturns409(t *testing.T) {
	truncateLinks(t)
	h := newRouter(t)

	w := doJSON(t, h, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://a.com",
		"short_name":   "dup",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", w.Code, w.Body.String())
	}

	w = doJSON(t, h, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://b.com",
		"short_name":   "dup",
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d, body=%s", w.Code, w.Body.String())
	}

	er := decodeJSON[errResp](t, w)
	if er.Error == "" {
		t.Fatalf("expected error message, got empty")
	}
}

func TestNotFoundReturns404(t *testing.T) {
	truncateLinks(t)
	h := newRouter(t)

	w := doJSON(t, h, http.MethodGet, "/api/links/999999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestInvalidJSONReturns400(t *testing.T) {
	truncateLinks(t)
	h := newRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewReader([]byte(`{"original_url":`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", w.Code, w.Body.String())
	}
}
