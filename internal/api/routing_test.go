package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRouting(t *testing.T) {
	handler := NewRouter(nil)
	if _, ok := handler.(*http.ServeMux); !ok {
		t.Fatalf("NewRouter returned %T, want *http.ServeMux", handler)
	}

	tests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodHead, "/health", http.StatusOK},
		{http.MethodGet, "/orders", http.StatusOK},
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
	handler := NewRouter(nil)
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
	handler := NewRouter(nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orders", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Body.String(); got != "{\"status\":\"placeholder\"}\n" {
		t.Fatalf("body = %q, want placeholder JSON", got)
	}
}
