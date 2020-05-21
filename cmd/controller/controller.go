package main

import (
	"flag"
	"os"

	"controller.com/cmd/daemon"

	"controller.com/api/client"

	"controller.com/config"
)

func main() {
	var flDaemon, fullDaemon, flDebug bool
	flag.BoolVar(&flDaemon, "d", false, "Enable daemon mode")
	flag.BoolVar(&fullDaemon, "daemon", false, "Enable daemon mode")
	flag.BoolVar(&flDebug, "debug", false, "Enable debug mode")
	flag.Parse()

	if flDaemon || fullDaemon || flDebug {
		daemon.RunServerWithDaemon()
		return
	}
	host := config.Host
	addr := config.Addr
	var ctrlCli *client.ControllerCli
	ctrlCli = client.CreateNewClient(os.Stdin, os.Stdout, os.Stderr, host, addr)
	if err := ctrlCli.Cmd(flag.Args()...); err != nil {
		panic(err) //TODO better handle method
	}
}
