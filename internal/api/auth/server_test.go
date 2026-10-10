package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authservice "github.com/dvtkrlbs/market/internal/services/auth"
)

func testCookieSigner(t *testing.T) *authservice.CookieSigner {
	t.Helper()
	signer, err := authservice.NewCookieSigner(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func TestForwardedIPPrecedence(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.3:1234"
	request.Header.Set("X-Forwarded-For", "198.51.100.2, 192.0.2.3")
	request.Header.Set("Forwarded", `for="[2001:db8::1]:4711";proto=https`)

	got := forwardedIP(request)
	if got == nil || *got != "2001:db8::1" {
		t.Fatalf("forwardedIP() = %v, want 2001:db8::1", got)
	}
}

func TestSessionCookieResponses(t *testing.T) {
	signer := testCookieSigner(t)
	token := authservice.SessionToken{Token: "test-token", ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second)}
	for _, secure := range []bool{false, true} {
		ctx := context.WithValue(context.Background(), secureCookieContextKey{}, secure)
		cookie := sessionCookie(ctx, token, signer)
		for _, response := range []struct {
			name  string
			visit func(http.ResponseWriter) error
		}{
			{"login", loginCookieResponse{Login200JSONResponse(sessionResponse(token)), cookie}.VisitLoginResponse},
			{"refresh", refreshCookieResponse{Refresh200JSONResponse(sessionResponse(token)), cookie}.VisitRefreshResponse},
		} {
			t.Run(response.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				if err := response.visit(w); err != nil {
					t.Fatal(err)
				}
				cookies := w.Result().Cookies()
				if len(cookies) != 1 {
					t.Fatalf("cookies = %v, want one session cookie", cookies)
				}
				got := cookies[0]
				verified, err := signer.Verify(got.Value)
				if err != nil || verified != token.Token {
					t.Fatalf("cookie signature verification = %q, %v", verified, err)
				}
				if got.Name != "session" || got.Value == token.Token || got.Path != "/" ||
					!got.Expires.Equal(token.ExpiresAt) || !got.HttpOnly || got.Secure != secure || got.SameSite != http.SameSiteLaxMode {
					t.Fatalf("unexpected session cookie: %v", got)
				}
				if w.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", w.Code)
				}
				var body map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if len(body) != 1 || body["expiresAt"] != token.ExpiresAt.Format(time.RFC3339) {
					t.Fatalf("body = %v, want only session expiry", body)
				}
			})
		}
	}
}

func TestLogoutCookieResponse(t *testing.T) {
	cookie := sessionCookie(context.Background(), authservice.SessionToken{}, testCookieSigner(t))
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0).UTC()
	w := httptest.NewRecorder()
	if err := (logoutCookieResponse{Logout204Response{}, cookie}).VisitLogoutResponse(w); err != nil {
		t.Fatal(err)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "" || cookies[0].MaxAge != -1 || cookies[0].Path != "/" {
		t.Fatalf("unexpected logout cookies: %v", cookies)
	}
	if w.Code != http.StatusNoContent || w.Body.Len() != 0 {
		t.Fatalf("response = %d %q, want empty 204", w.Code, w.Body.String())
	}
}

func TestRequestMetadataMiddlewareCookieSecurity(t *testing.T) {
	for _, url := range []string{"http://example.com/auth/login", "https://example.com/auth/login"} {
		request := httptest.NewRequest(http.MethodPost, url, nil)
		w := httptest.NewRecorder()
		RequestMetadataMiddleware(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			http.SetCookie(w, sessionCookie(request.Context(), authservice.SessionToken{Token: "test-token"}, testCookieSigner(t)))
		})).ServeHTTP(w, request)
		if got := w.Result().Cookies()[0].Secure; got != (request.TLS != nil) {
			t.Fatalf("Secure = %v for %s", got, url)
		}
	}
}

func TestForwardedIPFallsBackToRemoteAddress(t *testing.T) {
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.3:1234"

	got := forwardedIP(request)
	if got == nil || *got != "192.0.2.3" {
		t.Fatalf("forwardedIP() = %v, want 192.0.2.3", got)
	}
}
