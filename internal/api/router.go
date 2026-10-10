package api

import (
	"context"
	"fmt"
	"net/http"

	apidoc "github.com/dvtkrlbs/market/dist"
	authapi "github.com/dvtkrlbs/market/internal/api/auth"
	healthapi "github.com/dvtkrlbs/market/internal/api/health"
	productsapi "github.com/dvtkrlbs/market/internal/api/products"
	storesapi "github.com/dvtkrlbs/market/internal/api/stores"
	authservice "github.com/dvtkrlbs/market/internal/services/auth"
	productsservice "github.com/dvtkrlbs/market/internal/services/products"
	storeservice "github.com/dvtkrlbs/market/internal/services/stores"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

func NewRouter(authentication *authservice.Service, products *productsservice.Service, signer *authservice.CookieSigner, storeService *storeservice.Service) (http.Handler, error) {
	if signer == nil {
		return nil, fmt.Errorf("session cookie signer is required")
	}
	mux := http.NewServeMux()
	storesapi.HandlerFromMux(
		storesapi.NewStrictHandlerWithOptions(storesapi.NewServer(storeService, authentication), nil, storesapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		}),
		mux,
	)

	healthapi.HandlerFromMux(
		healthapi.NewStrictHandlerWithOptions(healthapi.NewServer(), nil, healthapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		}),
		mux,
	)
	authapi.HandlerFromMux(
		authapi.NewStrictHandlerWithOptions(authapi.NewServer(authentication, signer), nil, authapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		}),
		mux,
	)
	productsapi.HandlerFromMux(
		productsapi.NewStrictHandlerWithOptions(productsapi.NewServer(products), nil, productsapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  requestErrorHandler,
			ResponseErrorHandlerFunc: responseErrorHandler,
		}),
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
			AuthenticationFunc: authenticationResolver(authentication, signer),
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
