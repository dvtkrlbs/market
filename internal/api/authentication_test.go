package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authservice "github.com/dvtkrlbs/market/internal/auth"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
)

func testCookieSigner(t *testing.T) *authservice.CookieSigner {
	t.Helper()
	signer, err := authservice.NewCookieSigner(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func TestAuthenticationRejectsMissingOrInvalidSessionCookies(t *testing.T) {
	resolve := authenticationResolver(authservice.New(nil), testCookieSigner(t))
	for _, cookie := range []string{"", "session=", "session=invalid", "other=invalid",
		"session=AAAAAAAAAAAAAAAAAAAAAA.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"session=AAAAAAAAAAAAAAAAAAAAAA.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.invalid"} {
		t.Run(cookie, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/auth/whoami", nil)
			request.Header.Set("Cookie", cookie)
			// A structurally valid token would reach the database if the old
			// Authorization header were still accepted.
			request.Header.Set("Authorization", "Bearer AAAAAAAAAAAAAAAAAAAAAA.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			err := resolve(context.Background(), &openapi3filter.AuthenticationInput{
				RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request},
				SecuritySchemeName:     "SessionAuth",
				SecurityScheme:         &openapi3.SecurityScheme{Type: "apiKey", In: "cookie", Name: "session"},
			})
			if err == nil {
				t.Fatal("expected authentication failure")
			}
		})
	}
}
