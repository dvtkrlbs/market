package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSRequests(t *testing.T) {
	for _, tt := range []struct {
		name            string
		origin          string
		method          string
		preflightMethod string
		headers         string
		status          int
		called          bool
	}{
		{"allowed login", "http://localhost:5176", http.MethodPost, "", "", http.StatusOK, true},
		{"no origin", "", http.MethodGet, "", "", http.StatusOK, true},
		{"same origin", "http://localhost:8080", http.MethodPost, "", "", http.StatusOK, true},
		{"denied origin", "http://untrusted.example", http.MethodPost, "", "", http.StatusForbidden, false},
		{"preflight", "http://localhost:5176", http.MethodOptions, http.MethodPost, "content-type", http.StatusNoContent, false},
		{"custom header preflight", "http://localhost:5176", http.MethodOptions, http.MethodGet, "Content-Type, X-Request-Id", http.StatusNoContent, false},
		{"store update preflight", "http://localhost:5176", http.MethodOptions, http.MethodPatch, "content-type", http.StatusNoContent, false},
		{"store delete preflight", "http://localhost:5176", http.MethodOptions, http.MethodDelete, "", http.StatusNoContent, false},
		{"denied preflight method", "http://localhost:5176", http.MethodOptions, http.MethodPut, "", http.StatusMethodNotAllowed, false},
		{"denied preflight header", "http://localhost:5176", http.MethodOptions, http.MethodPost, "authorization", http.StatusForbidden, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler, err := WithCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}), []string{" http://localhost:5176 ", ""})
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(tt.method, "http://localhost:8080/auth/login", nil)
			request.Header.Set("Origin", tt.origin)
			request.Header.Set("Access-Control-Request-Method", tt.preflightMethod)
			request.Header.Set("Access-Control-Request-Headers", tt.headers)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, request)
			if w.Code != tt.status || called != tt.called {
				t.Fatalf("response = %d, handler called = %v; want %d, %v", w.Code, called, tt.status, tt.called)
			}
			if !strings.Contains(w.Header().Get("Vary"), "Origin") {
				t.Fatal("missing Vary: Origin")
			}
			if tt.origin != "" && tt.origin != "http://untrusted.example" {
				if w.Header().Get("Access-Control-Allow-Origin") != tt.origin || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
					t.Fatal("missing credentialed CORS headers")
				}
			} else if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Fatal("unexpected CORS origin header")
			}
		})
	}
}

func TestCORSRejectsInvalidOrigins(t *testing.T) {
	for _, origin := range []string{"*", "null", "ftp://example.com", "http://example.com/path", "http://example.com/", "http://user:password@example.com", "http://example.com?x=1", "http://example.com#x"} {
		if _, err := WithCORS(http.NotFoundHandler(), []string{origin}); err == nil {
			t.Errorf("accepted invalid origin %q", origin)
		}
	}
}
