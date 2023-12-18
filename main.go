package main

import (
	"github.com/pocketbase/pocketbase"
	"github.com/shynome/err0/try"
	_ "github.com/shynome/video2live/migrations"
)

var Version = "dev"

func main() {
	app := pocketbase.New()
	app.RootCmd.Version = Version
	initLive(app)
	try.To(app.Start())
}
