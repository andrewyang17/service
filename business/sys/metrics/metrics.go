package metrics

import (
	"context"
	"expvar"
	"runtime"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package is already based on a singleton
// for the different metrics that are registered with the package.
var m *metrics

// metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently.
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// The metrics value is stored in a package level variable since everything
// inside of expvar is registered as a singleton.
func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

type ctxKey int

const key ctxKey = 1

func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		if v.requests.Value()%100 == 0 {
			v.goroutines.Set(int64(runtime.NumGoroutine()))
		}
	}
}

func AddRequest(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.requests.Add(1)
	}
}

func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.errors.Add(1)
	}
}

func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.panics.Add(1)
	}
}
