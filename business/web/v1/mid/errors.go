package mid

import (
	"context"
	"net/http"

	"github.com/andrewyang17/service/business/sys/validate"
	"github.com/andrewyang17/service/business/web/v1"
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

				var er v1.ErrorResponse
				var status int

				switch act := validate.Cause(err).(type) {

				case validate.FieldErrors:
					er = v1.ErrorResponse{
						Error:  "data validation error",
						Fields: act.Error(),
					}
					status = http.StatusBadRequest

				case *v1.RequestError:
					er = v1.ErrorResponse{
						Error: act.Error(),
					}
					status = act.Status

				default:
					er = v1.ErrorResponse{
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
