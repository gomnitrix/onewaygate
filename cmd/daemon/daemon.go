package daemon

import (
	"os"

	"controller.com/api/server"
	"controller.com/config"
	//_ "controller.com/internal/app/daemon"
)

func RunServerWithDaemon() {
	server := server.NewServer(config.Addr, os.Stdin, os.Stdout, os.Stderr)
	server.StartServe()
}
