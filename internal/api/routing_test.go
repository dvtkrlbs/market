package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type routingServer struct{}

func (routingServer) GetHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (routingServer) ListEntries(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func (routingServer) CreateEntry(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

func TestHandlerRouting(t *testing.T) {
	handler := Handler(routingServer{})
	if _, ok := handler.(*http.ServeMux); !ok {
		t.Fatalf("Handler returned %T, want *http.ServeMux", handler)
	}

	tests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/health", http.StatusNotImplemented},
		{http.MethodHead, "/health", http.StatusNotImplemented},
		{http.MethodGet, "/entries", http.StatusNotImplemented},
		{http.MethodPost, "/entries", http.StatusNotImplemented},
		{http.MethodDelete, "/entries", http.StatusMethodNotAllowed},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
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
	handler := Handler(NewStrictHandler(NewServer(nil), nil))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Body.String(); got != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q, want health JSON", got)
	}
}
