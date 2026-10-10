package request

import (
	"context"
	"errors"

	"github.com/dvtkrlbs/market/internal/services/auth"
)

var (
	ErrUnauthenticated = errors.New("authentication required")
	ErrAdminRequired   = errors.New("admin access required")
)

type AdminChecker interface {
	IsAdmin(context.Context, int) (bool, error)
}

// RequireAdmin checks the authenticated principal against current admin membership.
// Never trust a user ID supplied by the caller or cached role data.
func RequireAdmin(ctx context.Context, admins AdminChecker) error {
	principal, ok := auth.PrincipalFromContext(ctx)
	if !ok || principal.User == nil || principal.User.ID <= 0 {
		return ErrUnauthenticated
	}
	allowed, err := admins.IsAdmin(ctx, principal.User.ID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrAdminRequired
	}
	return nil
}
