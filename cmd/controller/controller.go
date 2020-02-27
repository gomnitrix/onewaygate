package main

import (
	"flag"
	"fmt"
	"os"

	"controller.com/internal/app/mntisol"

	"controller.com/internal/app/pidisol"

	"controller.com/cmd/hider"

	"controller.com/internal"
)

func main() {
	runContCmd := flag.NewFlagSet("run", flag.ExitOnError)
	target := *runContCmd.String("t", "", "the target container")
	manager := *runContCmd.String("m", "", "the manager container")

	hideProcCmd := flag.NewFlagSet("hide", flag.ExitOnError)
	if len(os.Args) < 2 {
		fmt.Println("subcommand is required.")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "run":
		runContCmd.Parse(os.Args[2:])
		if target != "" {
			fmt.Println("the target:", target)
			fmt.Println("Trying to create new manager for the target...")
		} else {
			fmt.Println("no target, trying to create.")
			target = internal.CreateTarget()
		}
		if manager == "" {
			manager = internal.CreateRunManager(target)
			fmt.Println("Create successfully, the manager ID is ", manager)
		}
		pidisol.PidIsolation(manager)
		mntisol.MountIsolation(manager, target)
	case "hide":
		hideProcCmd.Parse(os.Args[2:]) //TODO mayby not used
		hider.Hide(os.Args[2:])
	}

}

///C:/Users/omnitrix/GoLand/controller.com
