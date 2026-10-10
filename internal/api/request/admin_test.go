package request

import (
	"context"
	"errors"
	"testing"

	"github.com/dvtkrlbs/market/internal/ent"
	"github.com/dvtkrlbs/market/internal/services/auth"
)

type stubAdminChecker struct {
	allowed bool
	err     error
	userID  int
}

func (s *stubAdminChecker) IsAdmin(_ context.Context, userID int) (bool, error) {
	s.userID = userID
	return s.allowed, s.err
}

func TestRequireAdmin(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tt := range []struct {
		name         string
		principal    *auth.Principal
		allowed      bool
		lookupErr    error
		want         error
		lookupUserID int
	}{
		{"signed out", nil, false, nil, ErrUnauthenticated, 0},
		{"missing user", &auth.Principal{}, false, nil, ErrUnauthenticated, 0},
		{"invalid user", &auth.Principal{User: &ent.User{}}, true, nil, ErrUnauthenticated, 0},
		{"not admin", &auth.Principal{User: &ent.User{ID: 7}}, false, nil, ErrAdminRequired, 7},
		{"admin", &auth.Principal{User: &ent.User{ID: 7}}, true, nil, nil, 7},
		{"lookup failure", &auth.Principal{User: &ent.User{ID: 7}}, false, dbErr, dbErr, 7},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := auth.ContextWithPrincipal(context.Background(), tt.principal)
			checker := &stubAdminChecker{allowed: tt.allowed, err: tt.lookupErr}
			if err := RequireAdmin(ctx, checker); !errors.Is(err, tt.want) {
				t.Fatalf("RequireAdmin error = %v, want %v", err, tt.want)
			}
			if checker.userID != tt.lookupUserID {
				t.Fatalf("lookup user = %d, want %d", checker.userID, tt.lookupUserID)
			}
		})
	}
}
