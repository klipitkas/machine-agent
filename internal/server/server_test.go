package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	os.Unsetenv("TOKEN")
	srv := New(":0")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}
}

func TestMetadataEndpoint(t *testing.T) {
	os.Unsetenv("TOKEN")
	srv := New(":0")
	req := httptest.NewRequest("GET", "/metadata", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMetadataWithInclude(t *testing.T) {
	os.Unsetenv("TOKEN")
	srv := New(":0")
	req := httptest.NewRequest("GET", "/metadata?include=host", nil)
	w := httptest.NewRecorder()
	srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareRejectsNoToken(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareAcceptsHeader(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer secret")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareAcceptsQueryParam(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/?token=secret", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthMiddlewareRejectsWrongToken(t *testing.T) {
	handler := authMiddleware("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/?token=wrong", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestParseSectionsDefault(t *testing.T) {
	req := httptest.NewRequest("GET", "/metadata", nil)
	sections := parseSections(req)

	if !sections["host"] || !sections["cpu"] || !sections["memory"] {
		t.Error("default sections should include host, cpu, memory")
	}
	if sections["docker"] || sections["processes"] {
		t.Error("docker and processes should not be in defaults")
	}
}

func TestParseSectionsAll(t *testing.T) {
	req := httptest.NewRequest("GET", "/metadata?include=all", nil)
	sections := parseSections(req)

	if !sections["docker"] || !sections["processes"] || !sections["host"] {
		t.Error("include=all should include all sections")
	}
}

func TestParseSectionsSpecific(t *testing.T) {
	req := httptest.NewRequest("GET", "/metadata?include=docker,disks", nil)
	sections := parseSections(req)

	if !sections["docker"] || !sections["disks"] {
		t.Error("should include docker and disks")
	}
	if sections["host"] || sections["cpu"] {
		t.Error("should not include host or cpu")
	}
}

func TestParseSectionsInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/metadata?include=fakesection", nil)
	sections := parseSections(req)

	if !sections["host"] {
		t.Error("invalid section should fall back to defaults")
	}
}
