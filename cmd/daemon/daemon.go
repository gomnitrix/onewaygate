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
	app.OnErrorCode(iris.StatusForbidden, server.Error403)
	app.OnErrorCode(iris.StatusInternalServerError, server.Error500)
	app.Post("/run", server.NewRunHandler)
	app.Post("/stop", server.NewRemoveHandler)
	web := app.Party("/owm")
	{
		web.Get("/login", server.WebLoginHandler)
		web.Post("/login", server.LoginHandler)
		web.Post("/register", server.RegisterHandler)
		web.Post("/logout", server.LogoutHandler)
		web.Get("/container/{name:string}/", server.CheckSession, server.WebContainerHandler)
		web.Get("/connect/{name:string}/{contid:string}", server.CheckSession, server.WebTerminalHandler)
		web.Get("/terminal/{name:string}/{contid:string}", server.CheckSession, server.TerminalHandler)
		web.Get("/main/{name:string}", server.CheckSession, server.WebMainViewHandler)
		web.Get("/table/{name:string}", server.CheckSession, server.WebTableViewHandler)
		web.Get("/add/{name:string}", server.CheckSession, server.WebAddHandler)
		web.Post("/add/{name:string}", server.CheckSession, server.AddHandler)
		web.Post("/remove/{name:string}", server.CheckSession, server.WebRemoveHandler)
	}
}

//{contId,contName,contType string}
func initApp(app *iris.Application) {
	//MyUpgrader := kgorilla.Upgrader(*mytmpu)
	//ws := websocket.New(MyUpgrader, websocket.Events{})
	//ws.OnConnect = server.TerminalHandler
	//ws.OnDisconnect = func(c *websocket.Conn) {
	//	log.Printf("[%s] Disconnected from server", c.ID())
	//}
	temp := iris.Django(templatePath, ".html")
	temp.Reload(true)
	app.HandleDir("/static", staticPath)
	app.RegisterView(temp)
	registerRouter(app)
}
