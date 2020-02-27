package client

import (
	"fmt"

	"controller.com/internal"
	"controller.com/internal/app/mntisol"
	"controller.com/internal/app/pidisol"

	"controller.com/config"
)

func (ctrlCli *ControllerCli) CmdHelp(args ...string) error {
	if len(args) >= 1 {
		method, exists := ctrlCli.getMethod(args[1])
		if !exists {
			fmt.Fprintf(ctrlCli.err, "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}
	help := fmt.Sprintf("Usage: Controller [OPTIONS] COMMAND [arg...]\n\nA container manager controller program which based on docker.\n\nCommands:\n", config.DEFAULTUNIXSOCKET)
	for _, command := range [][]string{} {
		help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
	}
	fmt.Fprintf(ctrlCli.out, "%s\n", help)
	return nil
}

func (ctrlCli *ControllerCli) CmdRun(args ...string) error {
	var cmdName string = "run"
	runContCmd := ctrlCli.subCmd(cmdName, "", "init and assign a new manager container for a target container")
	target := *runContCmd.String("-t", "", "provide an exist target container")
	if err := runContCmd.Parse(args); err != nil {
		fmt.Fprintf(ctrlCli.err, "Error: Args parse failed in "+cmdName)
		ctrlCli.CmdRun("--help")
		return nil
	}
	if target != "" {
		// TODO check the target ID
		fmt.Println("the target:", target)
		fmt.Println("Trying to create new manager for the target...")
	} else {
		fmt.Println("no target, trying to create.")
		target = internal.CreateTarget()
	}
	manager := internal.CreateRunManager(target)
	fmt.Fprintf(ctrlCli.out, "Create successfully, the manager ID is ", manager)

	// init the isolation relationship between manager and target
	pidisol.PidIsolation(manager)
	mntisol.MountIsolation(manager, target)
	//TODO add User ns and Net ns isolation
}
