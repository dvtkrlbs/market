package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dvtkrlbs/market/internal/api/common"
	authservice "github.com/dvtkrlbs/market/internal/auth"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func authenticationResolver(service *authservice.Service) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		if input.SecuritySchemeName != "BearerAuth" || input.SecurityScheme == nil ||
			input.SecurityScheme.Type != "http" || !strings.EqualFold(input.SecurityScheme.Scheme, "bearer") {
			return input.NewError(fmt.Errorf("unsupported security scheme %q", input.SecuritySchemeName))
		}
		if service == nil {
			return input.NewError(authservice.ErrInvalidSession)
		}
		request := input.RequestValidationInput.Request
		parts := strings.Fields(request.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return input.NewError(authservice.ErrInvalidSession)
		}
		principal, err := service.Authenticate(ctx, parts[1])
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
		w.Header().Set("WWW-Authenticate", "Bearer")
	} else if status == http.StatusNotFound {
		message = "not found"
	} else if status == http.StatusMethodNotAllowed {
		message = "method not allowed"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(common.Error{Message: message})
}
