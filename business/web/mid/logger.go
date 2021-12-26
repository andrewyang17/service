package mid

import (
	"context"
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/andrewyang17/service/foundation/web"
)

func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			log.Infow("request started", "traceID", v.TraceID, "statusCode", v.StatusCode, "method", r.Method,
				"path", r.URL.Path, "remoteAddr", r.RemoteAddr)

			err = handler(ctx, w, r)

			log.Infow("request completed", "traceID", v.TraceID, "statusCode", v.StatusCode, "method", r.Method,
				"path", r.URL.Path, "remoteAddr", r.RemoteAddr, "since", time.Since(v.Now))

			return err
		}

		return h
	}

	return m
}
