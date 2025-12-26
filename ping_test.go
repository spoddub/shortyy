package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	db "shorty/internal/db/sqlc"
	httpapi "shorty/internal/http"
)

func TestPing(t *testing.T) {
	router := httpapi.NewRouter(&db.Queries{}, "https://short.io")

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "pong" {
		t.Fatalf("expected body %q, got %q", "pong", w.Body.String())
	}
}
