package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"controller.com/internal"

	"controller.com/cmd/hider"

	"controller.com/config"
)

var log = config.ELog

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

func (ctrlCli *ControllerCli) CmdRun(args ...string) error { // TODO fix all the Cmd method error
	var cmdName string = "run"
	runContCmd := ctrlCli.subCmd(cmdName, "", "init and assign a new manager container for a target container")
	target := *runContCmd.String("-t", "", "provide an exist target container")
	if err := runContCmd.Parse(args); err != nil {
		fmt.Fprintf(ctrlCli.err, "Error: Args parse failed in "+cmdName)
		ctrlCli.CmdRun("--help")
		return nil
	}
	jsonBody := map[string]string{"target": target}
	res := ctrlCli.SendRequest(jsonBody, "/run")
	if !res {
		fmt.Fprint(ctrlCli.err, "send /run failed")
	}
	return nil
}

func (ctrlCli *ControllerCli) CmdStop(args ...string) error {
	stopContCmd := ctrlCli.subCmd("stop", "", "stop the container you appoint")
	manager := *stopContCmd.String("m", "", "the manager container you want to destory, this is a necessary param if you not provide the target manager")
	target := *stopContCmd.String("t", "", "the target container you want to destory, this is a alternative param. If you provide this param, this will destory the target and the manager")
	var managerID, targetID string
	if target != "" {
		targetID, err := internal.GetContainerFullID(target)
		if err != nil {
			log.Println(err)
		}
		if manager == "" {
			managerID = "manager_for_" + targetID[0:12]
		}
	} else if manager == "" {
		ctrlCli.CmdStop("--help")
		return errors.New("manager and target could not both be empty")
	} else {
		managerID = manager
	}

	managerID, err := internal.GetContainerFullID(managerID)
	if err != nil {
		log.Println(err)
	}
	jsonBody := map[string]string{"manager": managerID, "target": targetID}
	res := ctrlCli.SendRequest(jsonBody, "/stop")
	if !res {
		fmt.Fprint(ctrlCli.err, "send /stop failed")
	}
	return nil
}

func (ctrlCli *ControllerCli) CmdHide(args ...string) error {
	hider.Hide(args)
	return nil
}
func (ctrlCli *ControllerCli) SendRequest(jsonBody map[string]string, url string) bool {

	postBody, err := json.Marshal(jsonBody)
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return false
	}
	postAddr := config.Host + config.Addr + url
	resp, err := http.Post(postAddr, "application/json;charset=UTF-8", bytes.NewReader(postBody))
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return false
	}
	respBtyes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(ctrlCli.err, err.Error())
		return false
	}
	fmt.Fprintf(ctrlCli.out, string(respBtyes))
	return false
}
