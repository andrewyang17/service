package mid

import (
	"context"
	"net/http"

	"github.com/andrewyang17/service/business/sys/metrics"
	"github.com/andrewyang17/service/foundation/web"
)

// Metrics updates program counters.
func Metrics() web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			ctx = metrics.Set(ctx)

			err := handler(ctx, w, r)

			metrics.AddRequest(ctx)
			metrics.AddGoroutines(ctx)

			if err != nil {
				metrics.AddErrors(ctx)
			}

			return err
		}

		return h
	}

	return m
}
