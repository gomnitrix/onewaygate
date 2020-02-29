package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"controller.com/cmd/hider"

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
	postBody, err := json.Marshal(map[string]string{"target": target})
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return nil
	}
	postAddr := config.Host + config.Addr + "/run"
	resp, err := http.Post(postAddr, "application/json;charset=UTF-8", bytes.NewReader(postBody))
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return nil
	}
	respBtyes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return nil
	}
	fmt.Fprintf(ctrlCli.out, string(respBtyes))
	return nil
}

func (ctrlCli *ControllerCli) CmdHide(args ...string) error {
	hider.Hide(args)
	return nil
}
