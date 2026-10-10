package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dvtkrlbs/market/internal/api/common"
	authservice "github.com/dvtkrlbs/market/internal/auth"
	"github.com/dvtkrlbs/market/internal/ent"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type Server struct {
	auth   *authservice.Service
	signer *authservice.CookieSigner
}

type metadataContextKey struct{}
type secureCookieContextKey struct{}

var _ StrictServerInterface = (*Server)(nil)

func NewServer(service *authservice.Service, signer *authservice.CookieSigner) *Server {
	return &Server{auth: service, signer: signer}
}

func (s *Server) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	if request.Body == nil {
		return badRegisterRequest("request body is required"), nil
	}
	created, err := s.auth.Register(ctx, string(request.Body.Email), request.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, authservice.ErrInvalidEmail), errors.Is(err, authservice.ErrInvalidPassword):
			return badRegisterRequest(err.Error()), nil
		case errors.Is(err, authservice.ErrEmailAlreadyExists):
			return Register409JSONResponse(common.Error{Message: err.Error()}), nil
		default:
			return nil, err
		}
	}
	return Register201JSONResponse(userResponse(created)), nil
}

func (s *Server) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	if request.Body == nil {
		return badLoginRequest("request body is required"), nil
	}
	token, err := s.auth.Login(ctx, string(request.Body.Email), request.Body.Password, metadataFromContext(ctx))
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			return Login401JSONResponse{UnauthorizedJSONResponse: UnauthorizedJSONResponse{Message: err.Error()}}, nil
		}
		return nil, err
	}
	return loginCookieResponse{Login200JSONResponse(sessionResponse(token)), sessionCookie(ctx, token, s.signer)}, nil
}

func (s *Server) Logout(ctx context.Context, _ LogoutRequestObject) (LogoutResponseObject, error) {
	principal, ok := authservice.PrincipalFromContext(ctx)
	if !ok {
		return unauthorizedLogout(), nil
	}
	if err := s.auth.Logout(ctx, principal); err != nil {
		if errors.Is(err, authservice.ErrInvalidSession) {
			return unauthorizedLogout(), nil
		}
		return nil, err
	}
	cookie := sessionCookie(ctx, authservice.SessionToken{}, s.signer)
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(1, 0).UTC()
	return logoutCookieResponse{Logout204Response{}, cookie}, nil
}

func (s *Server) Whoami(ctx context.Context, _ WhoamiRequestObject) (WhoamiResponseObject, error) {
	principal, ok := authservice.PrincipalFromContext(ctx)
	if !ok {
		return Whoami401JSONResponse{UnauthorizedJSONResponse: UnauthorizedJSONResponse{Message: "authentication required"}}, nil
	}
	return Whoami200JSONResponse(Whoami{
		User:             userResponse(principal.User),
		SessionExpiresAt: principal.Session.ExpiresAt,
	}), nil
}

func (s *Server) Refresh(ctx context.Context, _ RefreshRequestObject) (RefreshResponseObject, error) {
	principal, ok := authservice.PrincipalFromContext(ctx)
	if !ok {
		return Refresh401JSONResponse{UnauthorizedJSONResponse: UnauthorizedJSONResponse{Message: "authentication required"}}, nil
	}
	token, err := s.auth.Refresh(ctx, principal)
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidSession) {
			return Refresh401JSONResponse{UnauthorizedJSONResponse: UnauthorizedJSONResponse{Message: "authentication required"}}, nil
		}
		return nil, err
	}
	return refreshCookieResponse{Refresh200JSONResponse(sessionResponse(token)), sessionCookie(ctx, token, s.signer)}, nil
}

func sessionCookie(ctx context.Context, token authservice.SessionToken, signer *authservice.CookieSigner) *http.Cookie {
	secure, _ := ctx.Value(secureCookieContextKey{}).(bool)
	value := ""
	if token.Token != "" {
		value = signer.Sign(token.Token)
	}
	return &http.Cookie{
		Name:     "session",
		Value:    value,
		Path:     "/",
		Expires:  token.ExpiresAt,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

type loginCookieResponse struct {
	Login200JSONResponse
	cookie *http.Cookie
}

func (response loginCookieResponse) VisitLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, response.cookie)
	return response.Login200JSONResponse.VisitLoginResponse(w)
}

type refreshCookieResponse struct {
	Refresh200JSONResponse
	cookie *http.Cookie
}

func (response refreshCookieResponse) VisitRefreshResponse(w http.ResponseWriter) error {
	http.SetCookie(w, response.cookie)
	return response.Refresh200JSONResponse.VisitRefreshResponse(w)
}

type logoutCookieResponse struct {
	Logout204Response
	cookie *http.Cookie
}

func (response logoutCookieResponse) VisitLogoutResponse(w http.ResponseWriter) error {
	http.SetCookie(w, response.cookie)
	return response.Logout204Response.VisitLogoutResponse(w)
}

func userResponse(account *ent.User) User {
	return User{
		ID:        int64(account.ID),
		Email:     openapi_types.Email(account.Email),
		CreatedAt: account.CreatedAt,
	}
}

func sessionResponse(token authservice.SessionToken) SessionExpiry {
	return SessionExpiry{ExpiresAt: token.ExpiresAt}
}

func badRegisterRequest(message string) Register400JSONResponse {
	return Register400JSONResponse{BadRequestJSONResponse: BadRequestJSONResponse{Message: message}}
}

func badLoginRequest(message string) Login400JSONResponse {
	return Login400JSONResponse{BadRequestJSONResponse: BadRequestJSONResponse{Message: message}}
}

func unauthorizedLogout() Logout401JSONResponse {
	return Logout401JSONResponse{UnauthorizedJSONResponse: UnauthorizedJSONResponse{Message: "authentication required"}}
}

func metadataFromRequest(request *http.Request) authservice.SessionMetadata {
	if request == nil {
		return authservice.SessionMetadata{}
	}
	return authservice.SessionMetadata{
		UserAgent: optionalString(request.UserAgent()),
		IP:        forwardedIP(request),
	}
}

func metadataFromContext(ctx context.Context) authservice.SessionMetadata {
	metadata, _ := ctx.Value(metadataContextKey{}).(authservice.SessionMetadata)
	return metadata
}

// RequestMetadataMiddleware captures connection metadata for newly-created sessions.
func RequestMetadataMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		ctx := context.WithValue(request.Context(), metadataContextKey{}, metadataFromRequest(request))
		ctx = context.WithValue(ctx, secureCookieContextKey{}, request.TLS != nil)
		next.ServeHTTP(w, request.WithContext(ctx))
	})
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func forwardedIP(request *http.Request) *string {
	if forwarded := request.Header.Get("Forwarded"); forwarded != "" {
		for _, parameter := range strings.Split(strings.Split(forwarded, ",")[0], ";") {
			name, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
			if ok && strings.EqualFold(name, "for") {
				if ip := parseForwardedIP(value); ip != nil {
					return ip
				}
			}
		}
	}
	if forwardedFor := request.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if ip := parseForwardedIP(strings.Split(forwardedFor, ",")[0]); ip != nil {
			return ip
		}
	}
	if host, _, err := net.SplitHostPort(request.RemoteAddr); err == nil {
		return parseForwardedIP(host)
	}
	return parseForwardedIP(request.RemoteAddr)
}

func parseForwardedIP(value string) *string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if strings.HasPrefix(value, "[") {
		if host, _, err := net.SplitHostPort(value); err == nil {
			value = host
		}
	} else if strings.Count(value, ":") == 1 {
		if host, _, err := net.SplitHostPort(value); err == nil {
			value = host
		}
	}
	ip := net.ParseIP(strings.Trim(value, "[]"))
	if ip == nil {
		return nil
	}
	normalized := ip.String()
	return &normalized
}
