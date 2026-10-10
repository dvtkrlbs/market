package stores

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	apidoc "github.com/dvtkrlbs/market/dist"
	"github.com/dvtkrlbs/market/internal/auth"
	"github.com/dvtkrlbs/market/internal/ent"
	storeservice "github.com/dvtkrlbs/market/internal/stores"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	nethttpmiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type stubService struct {
	err     error
	called  bool
	mode    string
	userID  int
	storeID int
	name    string
	update  storeservice.Update
}

func (s *stubService) List(_ context.Context, userID int, mode string) ([]*ent.Store, error) {
	s.called, s.userID, s.mode = true, userID, mode
	return []*ent.Store{}, s.err
}
func (s *stubService) Create(_ context.Context, userID int, name string) (*ent.Store, error) {
	s.called, s.userID, s.name = true, userID, name
	return &ent.Store{ID: 2, Name: name}, s.err
}
func (s *stubService) Get(_ context.Context, userID, storeID int) (*ent.Store, error) {
	s.called, s.userID, s.storeID = true, userID, storeID
	return &ent.Store{ID: storeID, Name: "Shop", Approved: true}, s.err
}
func (s *stubService) Update(_ context.Context, userID, storeID int, input storeservice.Update) (*ent.Store, error) {
	s.called, s.userID, s.storeID, s.update = true, userID, storeID, input
	return &ent.Store{ID: storeID, Name: "Updated", Approved: true}, s.err
}
func (s *stubService) Delete(_ context.Context, userID, storeID int) error {
	s.called, s.userID, s.storeID = true, userID, storeID
	return s.err
}

func (s *stubService) Approve(_ context.Context, storeID int) (*ent.Store, error) {
	s.called, s.storeID = true, storeID
	return &ent.Store{ID: storeID, Name: "Shop", Approved: true}, s.err
}

type stubAdmins struct {
	allowed bool
	err     error
	userID  int
}

func (s *stubAdmins) IsAdmin(_ context.Context, userID int) (bool, error) {
	s.userID = userID
	return s.allowed, s.err
}

func TestStoreHTTP(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData(apidoc.OpenAPI)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		err    error
		status int
		called bool
	}{
		{"shop default", "GET", "/users/1/stores", "", nil, 200, true},
		{"shop", "GET", "/users/1/stores?for=shop", "", nil, 200, true},
		{"manage", "GET", "/users/1/stores?for=manage", "", nil, 200, true},
		{"admin", "GET", "/users/1/stores?for=admin", "", nil, 200, true},
		{"invalid mode", "GET", "/users/1/stores?for=other", "", nil, 400, false},
		{"invalid user", "GET", "/users/0/stores", "", nil, 400, false},
		{"get", "GET", "/users/1/stores/2", "", nil, 200, true},
		{"get missing", "GET", "/users/1/stores/2", "", storeservice.ErrNotFound, 404, true},
		{"create", "POST", "/users/1/stores", `{"name":"Shop"}`, nil, 201, true},
		{"create missing name", "POST", "/users/1/stores", `{}`, nil, 400, false},
		{"create empty name", "POST", "/users/1/stores", `{"name":""}`, nil, 400, false},
		{"create whitespace", "POST", "/users/1/stores", `{"name":"  "}`, storeservice.ErrNameRequired, 400, true},
		{"cannot create approved", "POST", "/users/1/stores", `{"name":"Shop","approved":true}`, nil, 400, false},
		{"update", "PATCH", "/users/1/stores/2", `{"name":"Updated","approved":true}`, nil, 200, true},
		{"update false", "PATCH", "/users/1/stores/2", `{"approved":false}`, nil, 200, true},
		{"empty update", "PATCH", "/users/1/stores/2", `{}`, nil, 400, false},
		{"update unknown field", "PATCH", "/users/1/stores/2", `{"deleted_at":null}`, nil, 400, false},
		{"update blank name", "PATCH", "/users/1/stores/2", `{"name":" "}`, storeservice.ErrNameRequired, 400, true},
		{"update forbidden", "PATCH", "/users/1/stores/2", `{"approved":true}`, storeservice.ErrForbidden, 403, true},
		{"update missing", "PATCH", "/users/1/stores/2", `{"approved":true}`, storeservice.ErrNotFound, 404, true},
		{"delete", "DELETE", "/users/1/stores/2", "", nil, 204, true},
		{"delete forbidden", "DELETE", "/users/1/stores/2", "", storeservice.ErrForbidden, 403, true},
		{"delete missing", "DELETE", "/users/1/stores/2", "", storeservice.ErrNotFound, 404, true},
		{"internal error", "GET", "/users/1/stores/2", "", errors.New("database secret"), 500, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &stubService{err: tt.err}
			handler := Handler(NewStrictHandlerWithOptions(NewServer(service, &stubAdmins{allowed: true}), nil, StrictHTTPServerOptions{
				RequestErrorHandlerFunc:  func(w http.ResponseWriter, _ *http.Request, _ error) { http.Error(w, "invalid request", 400) },
				ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) { http.Error(w, "internal server error", 500) },
			}))
			validator := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
				Options: openapi3filter.Options{AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
					r := input.RequestValidationInput.Request
					*r = *r.WithContext(auth.ContextWithPrincipal(r.Context(), &auth.Principal{User: &ent.User{ID: 1}}))
					return nil
				}},
			})
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			request.AddCookie(&http.Cookie{Name: "session", Value: "test"})
			if tt.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			response := httptest.NewRecorder()
			validator(handler).ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tt.status, response.Body.String())
			}
			if service.called != tt.called {
				t.Fatalf("service called = %t, want %t", service.called, tt.called)
			}
			if service.called && service.userID != 1 {
				t.Fatalf("userID = %d", service.userID)
			}
			if (tt.name == "manage" || tt.name == "admin") && service.mode != tt.name {
				t.Fatalf("mode = %q", service.mode)
			}
			if tt.name == "shop default" && service.mode != "shop" {
				t.Fatalf("mode = %q", service.mode)
			}
			if tt.name == "update false" && (service.update.Approved == nil || *service.update.Approved) {
				t.Fatal("false approval was not passed through")
			}
			if tt.name == "create" && response.Body.String() != "{\"approved\":false,\"id\":2,\"name\":\"Shop\"}\n" {
				t.Fatalf("created JSON = %s", response.Body.String())
			}
			if tt.name == "shop default" && response.Body.String() != "[]\n" {
				t.Fatalf("empty list = %s", response.Body.String())
			}
			if tt.status == 500 && strings.Contains(response.Body.String(), "database secret") {
				t.Fatal("internal error leaked")
			}
		})
	}
}

func TestStoresRequireMatchingPrincipal(t *testing.T) {
	for _, userID := range []int{0, 2} {
		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			t.Run(method+fmtID(userID), func(t *testing.T) {
				service := &stubService{}
				handler := Handler(NewStrictHandler(NewServer(service, &stubAdmins{allowed: true}), nil))
				path, body := "/users/1/stores/2", ""
				if method == "POST" {
					path, body = "/users/1/stores", `{"name":"Shop"}`
				}
				if method == "PATCH" {
					body = `{"approved":true}`
				}
				request := httptest.NewRequest(method, path, strings.NewReader(body))
				if userID != 0 {
					request = request.WithContext(auth.ContextWithPrincipal(request.Context(), &auth.Principal{User: &ent.User{ID: userID}}))
				}
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				want := 401
				if userID != 0 {
					want = 403
				}
				if response.Code != want || service.called {
					t.Fatalf("status = %d, service called = %t", response.Code, service.called)
				}
			})
		}
	}
}

func fmtID(id int) string {
	if id == 0 {
		return "/unauthenticated"
	}
	return "/other-user"
}

func TestApproveStoreHTTP(t *testing.T) {
	spec, err := openapi3.NewLoader().LoadFromData(apidoc.OpenAPI)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name       string
		path       string
		userID     int
		admin      bool
		adminErr   error
		serviceErr error
		status     int
		called     bool
	}{
		{"admin approves", "/stores/2/approve", 7, true, nil, nil, 200, true},
		{"signed out", "/stores/2/approve", 0, false, nil, nil, 401, false},
		{"store owner is not platform admin", "/stores/2/approve", 1, false, nil, nil, 403, false},
		{"lookup error fails closed", "/stores/2/approve", 7, true, errors.New("database secret"), nil, 500, false},
		{"store missing", "/stores/2/approve", 7, true, nil, storeservice.ErrNotFound, 404, true},
		{"service error", "/stores/2/approve", 7, true, nil, errors.New("database secret"), 500, true},
		{"invalid store ID", "/stores/0/approve", 7, true, nil, nil, 400, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := &stubService{err: tt.serviceErr}
			admins := &stubAdmins{allowed: tt.admin, err: tt.adminErr}
			handler := Handler(NewStrictHandlerWithOptions(NewServer(service, admins), nil, StrictHTTPServerOptions{
				ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) { http.Error(w, "internal server error", 500) },
			}))
			validator := nethttpmiddleware.OapiRequestValidatorWithOptions(spec, &nethttpmiddleware.Options{
				Options: openapi3filter.Options{AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error { return nil }},
			})
			req := httptest.NewRequest("POST", tt.path, nil)
			if tt.userID != 0 {
				req = req.WithContext(auth.ContextWithPrincipal(req.Context(), &auth.Principal{User: &ent.User{ID: tt.userID}}))
			}
			response := httptest.NewRecorder()
			validator(handler).ServeHTTP(response, req)
			if response.Code != tt.status || service.called != tt.called {
				t.Fatalf("status = %d, called = %t, body = %s", response.Code, service.called, response.Body.String())
			}
			if service.called && service.storeID != 2 {
				t.Fatalf("store ID = %d", service.storeID)
			}
			if tt.status == 200 {
				if admins.userID != tt.userID {
					t.Fatalf("checked user ID = %d", admins.userID)
				}
				if response.Body.String() != "{\"approved\":true,\"id\":2,\"name\":\"Shop\"}\n" {
					t.Fatalf("JSON = %s", response.Body.String())
				}
			}
			if strings.Contains(response.Body.String(), "database secret") {
				t.Fatal("internal error leaked")
			}
		})
	}
}

func TestPatchApprovalRequiresGlobalAdmin(t *testing.T) {
	for _, body := range []string{`{"approved":true}`, `{"approved":false}`} {
		service := &stubService{}
		admins := &stubAdmins{}
		handler := Handler(NewStrictHandler(NewServer(service, admins), nil))
		req := httptest.NewRequest("PATCH", "/users/1/stores/2", strings.NewReader(body))
		req = req.WithContext(auth.ContextWithPrincipal(req.Context(), &auth.Principal{User: &ent.User{ID: 1}}))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != 403 || service.called || admins.userID != 1 {
			t.Fatalf("status = %d, called = %t, checked user = %d", response.Code, service.called, admins.userID)
		}
	}
	// Store owners can still edit names without platform admin membership.
	service := &stubService{}
	admins := &stubAdmins{}
	handler := Handler(NewStrictHandler(NewServer(service, admins), nil))
	req := httptest.NewRequest("PATCH", "/users/1/stores/2", strings.NewReader(`{"name":"Renamed"}`))
	req = req.WithContext(auth.ContextWithPrincipal(req.Context(), &auth.Principal{User: &ent.User{ID: 1}}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != 200 || !service.called || admins.userID != 0 {
		t.Fatalf("status = %d", response.Code)
	}
}
