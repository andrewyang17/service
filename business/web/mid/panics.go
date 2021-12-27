package mid

import (
	"context"
	"fmt"
	"github.com/andrewyang17/service/business/sys/metrics"
	"net/http"
	"runtime/debug"

	"github.com/andrewyang17/service/foundation/web"
)

// Panics recovers from panics and converts the panic to an error, so it is
// reported in Metrics and handled in Errors.
func Panics() web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			defer func() {
				if rec := recover(); rec != nil {

					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

					metrics.AddPanics(ctx)
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}