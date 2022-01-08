package v1

import (
	"net/http"

	"github.com/andrewyang17/service/app/services/sales-api/handlers/v1/usergrp"
	"github.com/andrewyang17/service/business/core/user"
	"github.com/andrewyang17/service/business/sys/auth"
	"github.com/andrewyang17/service/business/web/v1/mid"
	"github.com/andrewyang17/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Config struct {
	Log  *zap.SugaredLogger
	Auth *auth.Auth
	DB   *sqlx.DB
}

func Routes(app *web.App, cfg Config) {
	const version = "v1"
	authen := mid.Authenticate(cfg.Auth)
	admin := mid.Authorize(auth.RoleAdmin)

	// Register user management and authentication endpoints.
	ugh := usergrp.Handlers{
		User: user.NewCore(cfg.Log, cfg.DB),
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/users/token", ugh.Token)
	app.Handle(http.MethodGet, version, "/users/:page/:rows", ugh.Query, authen, admin)
	app.Handle(http.MethodGet, version, "/users/:id", ugh.QueryByID, authen)
	app.Handle(http.MethodPost, version, "/users", ugh.Create, authen, admin)
	app.Handle(http.MethodPut, version, "/users/:id", ugh.Update, authen, admin)
	app.Handle(http.MethodDelete, version, "/users/:id", ugh.Delete, authen, admin)
}
