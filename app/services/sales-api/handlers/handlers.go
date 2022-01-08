// Package handlers manages the different versions of the API.
package handlers

import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/andrewyang17/service/app/services/sales-api/handlers/debug/checkgrp"
	v1 "github.com/andrewyang17/service/app/services/sales-api/handlers/v1"
	"github.com/andrewyang17/service/business/sys/auth"
	"github.com/andrewyang17/service/business/web/v1/mid"
	"github.com/andrewyang17/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// WithCORS provides configuration options for CORS.
func WithCORS(origin string) func(opts *Options) {
	return func(opts *Options) {
		opts.corsOrigin = origin
	}
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig, options ...func(opts *Options)) http.Handler {
	var opts Options

	for _, option := range options {
		option(&opts)
	}

	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	if opts.corsOrigin != "" {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return nil
		}

		app.Handle(http.MethodOptions, "", "/*", h, mid.Cors(opts.corsOrigin))
	}

	v1.Routes(app, v1.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})

	return app
}

func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}
