package testgrp

import (
	"context"
	"github.com/andrewyang17/service/foundation/web"
	"math/rand"
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		//return errors.New("untrusted error")
		//return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
		return web.NewShutdownError("shutting down")
	}

	data := struct{
		Status string
	}{
		Status: "ok",
	}

	return web.Response(ctx, w, http.StatusOK, data)
}
