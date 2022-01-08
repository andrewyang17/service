package mid

import (
	"context"
	"net/http"

	"github.com/andrewyang17/service/business/sys/validate"
	v1Web "github.com/andrewyang17/service/business/web/v1"
	"github.com/andrewyang17/service/foundation/web"
	"go.uber.org/zap"
)

func Errors(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			if err := handler(ctx, w, r); err != nil {
				// Log the error.
				log.Errorw("ERROR", "traceID", v.TraceID, "ERROR", err)

				var er v1Web.ErrorResponse
				var status int

				switch {
				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					er = v1Web.ErrorResponse{
						Error:  "data validation error",
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case v1Web.IsRequestError(err):
					reqErr := v1Web.GetRequestError(err)
					er = v1Web.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					er = v1Web.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				if err := web.Response(ctx, w, status, er); err != nil {
					return err
				}

				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}
