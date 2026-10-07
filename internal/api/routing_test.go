package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRouting(t *testing.T) {
	handler, err := NewRouter(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodHead, "/health", http.StatusOK},
		{http.MethodGet, "/orders", http.StatusUnauthorized},
		{http.MethodDelete, "/entries", http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
		{http.MethodPost, "/orders", http.StatusMethodNotAllowed},
		{http.MethodGet, "/missing", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d", response.Code, tt.status)
			}
		})
	}
}

func TestHandlerHealth(t *testing.T) {
	handler, err := NewRouter(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Body.String(); got != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q, want health JSON", got)
	}
}

func TestHandlerOrdersPlaceholder(t *testing.T) {
	handler, err := NewRouter(nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if got := response.Body.String(); got != "{\"message\":\"authentication required\"}\n" {
		t.Fatalf("body = %q, want authentication error", got)
	}
}

func TestAuthRoutesArePublicOrProtectedAsDocumented(t *testing.T) {
	handler, err := NewRouter(nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodPost, "/auth/register", http.StatusBadRequest},
		{http.MethodPost, "/auth/login", http.StatusBadRequest},
		{http.MethodPost, "/auth/logout", http.StatusUnauthorized},
		{http.MethodPost, "/auth/refresh", http.StatusUnauthorized},
		{http.MethodGet, "/auth/whoami", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(tt.method, tt.path, nil)
			handler.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.status, response.Body.String())
			}
			if tt.status == http.StatusUnauthorized && response.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("WWW-Authenticate = %q, want Bearer", response.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
