package daemon

import (
	"controller.com/api/server"
	//_ "controller.com/internal/app/daemon"

	"github.com/kataras/iris"
)

func RunServerWithDaemon() {
	app := iris.New()
	server.StartServe(app)
}
