package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/shynome/err0/try"
	_ "github.com/shynome/video2live/migrations"
)

var Version = "dev"

func main() {
	app := pocketbase.New()
	app.RootCmd.Version = Version
	initLive(app)
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/", func(c echo.Context) error {
			return c.Redirect(http.StatusTemporaryRedirect, "/_/")
		})
		return nil
	})
	try.To(app.Start())
}
