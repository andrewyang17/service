package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/andrewyang17/service/business/sys/auth"
	"github.com/andrewyang17/service/business/web/v1"
	"github.com/andrewyang17/service/foundation/web"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			authStr := r.Header.Get("authorization")

			parts := strings.Split(authStr, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected authorization header format: bearer <token>")
				return v1.NewRequestError(err, http.StatusUnauthorized)
			}

			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return v1.NewRequestError(err, http.StatusUnauthorized)
			}

			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(roles ...string) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := auth.GetClaims(ctx)
			if err != nil {
				return v1.NewRequestError(
					fmt.Errorf("you are not authorized for that action, no claims"),
					http.StatusForbidden,
				)
			}

			if !claims.Authorized(roles...) {
				return v1.NewRequestError(
					fmt.Errorf("you are not authorized for that action, claims[%v] roles[%v]", claims.Roles, roles),
					http.StatusForbidden,
				)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
