package main

import (
	"flag"
	"os"

	"controller.com/cmd/daemon"

	"controller.com/api/client"

	"controller.com/config"
)

var flDaemon = flag.Bool("d", false, "Enable daemon mode")

func main() {
	flag.Parse()

	if *flDaemon {
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
