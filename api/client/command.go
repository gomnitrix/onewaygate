package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"controller.com/internal"

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
	internal.PrintTitle()
	help := fmt.Sprintf("Usage: Controller [OPTIONS] COMMAND [arg...]\n\nA container manager controller program which based on docker.\n\nCommands:\n", config.DEFAULTUNIXSOCKET)
	for _, command := range [][]string{} {
		help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
	}
	fmt.Fprintf(ctrlCli.out, "%s\n", help)
	return nil
}

func (ctrlCli *ControllerCli) CmdRun(args ...string) error { // TODO fix all the Cmd method error
	var cmdName string = "run"
	var target, name string
	runContCmd := ctrlCli.subCmd(cmdName, "", "init and assign a new manager container for a target container")
	runContCmd.StringVar(&target, "t", "", "provide an exist target container")
	runContCmd.StringVar(&name, "name", "", "the target container name you want to assign to a new container,this is an alternative param")
	if err := runContCmd.Parse(args); err != nil {
		fmt.Fprintf(ctrlCli.err, "Error: Args parse failed in "+cmdName)
		ctrlCli.CmdRun("--help")
		return nil
	}
	jsonBody := map[string]string{"target": target, "TgtName": name}
	res := ctrlCli.SendRequest(jsonBody, "/run")
	if !res {
		fmt.Fprint(ctrlCli.err, "send /run failed")
	}
	return nil
}

func (ctrlCli *ControllerCli) CmdRemove(args ...string) error {
	var manager, target string
	stopContCmd := ctrlCli.subCmd("remove", "", "stop the container you appoint")
	stopContCmd.StringVar(&manager, "m", "", "the manager container you want to destory, this is a necessary param if you not provide the target manager")
	stopContCmd.StringVar(&target, "t", "", "the target container you want to destory, this is a alternative param. If you provide this param, this will destory the target and the manager")
	stopContCmd.Parse(args)
	var managerID, targetID string
	if target == "" && manager == "" {
		ctrlCli.CmdRemove("--help")
		return errors.New("manager and target could not both be empty")
	}
	if manager == "" {
		//TODO 数据库查询manager id
	} else {
		managerID = manager
	}
	if target != "" {
		var err error
		targetID, err = internal.GetContainerFullID(target)
		if err != nil {
			fmt.Fprint(ctrlCli.err, err.Error())
			os.Exit(1)
		}
	}
	managerID, err := internal.GetContainerFullID(managerID)
	if err != nil {
		fmt.Fprint(ctrlCli.err, err.Error())
		os.Exit(1)
	}
	managerID = managerID[0:12]
	targetID = targetID[0:12]
	jsonBody := map[string]string{"manager": managerID, "target": targetID}
	res := ctrlCli.SendRequest(jsonBody, "/stop")
	if !res {
		fmt.Fprint(ctrlCli.err, "send /stop failed\n")
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
	resp, err := http.Post(postAddr, "application/json;charset=utf-8", bytes.NewReader(postBody))
	defer resp.Body.Close()
	if err != nil {
		fmt.Fprintln(ctrlCli.err, err.Error())
		return false
	}
	respBtyes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintln(ctrlCli.err, err.Error())
		return false
	}
	fmt.Fprintln(ctrlCli.out, string(respBtyes))
	return true
}
