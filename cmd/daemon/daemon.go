package daemon

import (
	"controller.com/api/server"
	"controller.com/config"
	"controller.com/internal"

	//_ "controller.com/internal/app/daemon"

	"github.com/kataras/iris"
)

var templatePath = internal.JoinPath(config.TemplatePath)
var staticPath = internal.JoinPath(config.StaticPath)

func RunServerWithDaemon() {
	app := iris.New()
	initApp(app)
	server.StartServe(app)
}

func registerRouter(app *iris.Application) {
	app.OnErrorCode(iris.StatusNotFound, server.Error404)
	app.Post("/run", server.NewRunHandler)
	app.Post("/stop", server.NewStopHandler)
	web := app.Party("/owm")
	{
		web.Get("/login", server.WebLoginHandler)
		web.Get("/main", server.WebRunHandler)
		web.Get("/table", server.WebTableViewHandler)
	}
}

func initApp(app *iris.Application) {
	temp := iris.Django(templatePath, ".html")
	temp.Reload(true)
	app.HandleDir("/static", staticPath)
	app.RegisterView(temp)
	registerRouter(app)
}
