package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dvtkrlbs/market/internal/api/common"
	authservice "github.com/dvtkrlbs/market/internal/services/auth"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func authenticationResolver(service *authservice.Service, signer *authservice.CookieSigner) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if input.SecuritySchemeName != "SessionAuth" || input.SecurityScheme == nil ||
			input.SecurityScheme.Type != "apiKey" || input.SecurityScheme.In != "cookie" || input.SecurityScheme.Name != "session" {
			return input.NewError(fmt.Errorf("unsupported security scheme %q", input.SecuritySchemeName))
		}
		if service == nil || signer == nil {
			return input.NewError(authservice.ErrInvalidSession)
		}
		request := input.RequestValidationInput.Request
		cookie, err := request.Cookie("session")
		if err != nil || cookie.Value == "" {
			return input.NewError(authservice.ErrInvalidSession)
		}
		token, err := signer.Verify(cookie.Value)
		if err != nil {
			return input.NewError(authservice.ErrInvalidSession)
		}
		principal, err := service.Authenticate(ctx, token)
		if err != nil {
			return input.NewError(authservice.ErrInvalidSession)
		}
		*request = *request.WithContext(authservice.ContextWithPrincipal(request.Context(), principal))
		return nil
	}
}

func validationErrorHandler(_ context.Context, err error, w http.ResponseWriter, _ *http.Request, opts nethttpmiddleware.ErrorHandlerOpts) {
	status := opts.StatusCode
	message := "invalid request"
	var securityErr *openapi3filter.SecurityRequirementsError
	if errors.As(err, &securityErr) {
		status = http.StatusUnauthorized
		message = "authentication required"
	} else if status == http.StatusNotFound {
		message = "not found"
	} else if status == http.StatusMethodNotAllowed {
		message = "method not allowed"
	}
	writeAPIError(w, status, message)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(common.Error{Message: message})
}
