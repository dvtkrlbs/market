package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// WithCORS permits credentialed browser requests from explicitly configured origins.
// It must wrap request validation so preflights do not require authentication.
func WithCORS(next http.Handler, origins []string) (http.Handler, error) {
	allowed := make(map[string]bool)
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
			parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return nil, fmt.Errorf("invalid CORS origin %q: use an HTTP(S) origin without a path", origin)
		}
		allowed[origin] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := request.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, request)
			return
		}
		scheme := "http"
		if request.TLS != nil {
			scheme = "https"
		}
		if !allowed[origin] && origin != scheme+"://"+request.Host {
			writeAPIError(w, http.StatusForbidden, "origin not allowed")
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if request.Method == http.MethodOptions && request.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")
			switch request.Header.Get("Access-Control-Request-Method") {
			case http.MethodGet, http.MethodHead, http.MethodPost:
			default:
				writeAPIError(w, http.StatusMethodNotAllowed, "CORS method not allowed")
				return
			}
			for _, header := range strings.Split(request.Header.Get("Access-Control-Request-Headers"), ",") {
				switch strings.ToLower(strings.TrimSpace(header)) {
				case "", "content-type", "x-request-id":
				default:
					writeAPIError(w, http.StatusForbidden, "CORS header not allowed")
					return
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, request)
	}), nil
}
