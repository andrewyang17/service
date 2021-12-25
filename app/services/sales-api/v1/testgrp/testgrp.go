package testgrp

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world")
}
