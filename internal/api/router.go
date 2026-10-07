package api

import (
	"context"
	"fmt"
	"net/http"

	apidoc "github.com/dvtkrlbs/market/dist"
	authapi "github.com/dvtkrlbs/market/internal/api/auth"
	"github.com/dvtkrlbs/market/internal/api/entries"
	"github.com/dvtkrlbs/market/internal/api/health"
	"github.com/dvtkrlbs/market/internal/api/orders"
	authservice "github.com/dvtkrlbs/market/internal/auth"
	"github.com/dvtkrlbs/market/internal/bookkeeping"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func NewRouter(service *bookkeeping.Service, authentication *authservice.Service) (http.Handler, error) {
	mux := http.NewServeMux()

	health.HandlerFromMux(
		health.NewStrictHandler(health.NewServer(), nil),
		mux,
	)
	entries.HandlerFromMux(
		entries.NewStrictHandler(entries.NewServer(service), nil),
		mux,
	)
	orders.HandlerFromMux(
		orders.NewStrictHandler(orders.NewServer(), nil),
		mux,
	)
	authapi.HandlerFromMux(
		authapi.NewStrictHandler(authapi.NewServer(authentication), nil),
		mux,
	)

	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromData(apidoc.OpenAPI)
	if err != nil {
		return nil, fmt.Errorf("load embedded OpenAPI contract: %w", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate embedded OpenAPI contract: %w", err)
	}
	validator := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: authenticationResolver(authentication),
		},
		ErrorHandlerWithOpts: validationErrorHandler,
		Skipper: func(request *http.Request) bool {
			pathItem := spec.Paths.Find(request.URL.Path)
			if pathItem == nil {
				return false
			}
			if request.Method == http.MethodHead {
				// net/http serves HEAD through a GET pattern. Permit that only
				// for public GET operations; protected GETs must still pass auth.
				if pathItem.Get == nil {
					return false
				}
				security := spec.Security
				if pathItem.Get.Security != nil {
					security = *pathItem.Get.Security
				}
				return len(security) == 0
			}
			// Let ServeMux preserve its standard 405 response for methods the
			// contract does not define. No endpoint handler can run in this case.
			return pathItem.GetOperation(request.Method) == nil
		},
	})
	return validator(authapi.RequestMetadataMiddleware(mux)), nil
}
