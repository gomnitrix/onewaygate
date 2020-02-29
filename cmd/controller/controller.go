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

//hideProcCmd := flag.NewFlagSet("hide", flag.ExitOnError)
//if len(os.Args) < 2 {
//	fmt.Println("subcommand is required.")
//	os.Exit(1)
//}
//switch os.Args[1] {
//case "run":
//	runContCmd.Parse(os.Args[2:])
//	if target != "" {
//		fmt.Println("the target:", target)
//		fmt.Println("Trying to create new manager for the target...")
//	} else {
//		fmt.Println("no target, trying to create.")
//		target = internal.CreateTarget()
//	}
//	if manager == "" {
//		manager = internal.CreateRunManager(target)
//		fmt.Println("Create successfully, the manager ID is ", manager)
//	}
//	pidisol.PidIsolation(manager)
//	mntisol.MountIsolation(manager, target)
//case "hide":
//	hideProcCmd.Parse(os.Args[2:]) //TODO mayby not used
//	hider.Hide(os.Args[2:])
//}

///C:/Users/omnitrix/GoLand/controller.com
