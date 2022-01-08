// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	mux      *httptreemux.ContextMux
	otmux    http.Handler
	shutdown chan os.Signal
	mv       []Middleware
}

func NewApp(shutdown chan os.Signal, mv ...Middleware) *App {
	mux := httptreemux.NewContextMux()

	return &App{
		mux:      mux,
		otmux:    otelhttp.NewHandler(mux, "request"),
		shutdown: shutdown,
		mv:       mv,
	}
}

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

func (a *App) Handle(method string, group string, path string, handler Handler, mv ...Middleware) {
	// Wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mv, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mv, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Capture the parent request span from the context.
		span := trace.SpanFromContext(ctx)

		v := Values{
			TraceID: span.SpanContext().TraceID().String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(ctx, key, &v)

		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.mux.Handle(method, finalPath, h)
}
