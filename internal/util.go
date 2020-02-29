package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"controller.com/config"
	"github.com/docker/docker/api/types"

	"github.com/docker/docker/client"
)

func InitClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(config.DockerVersion))
	if err != nil {
		panic(err)
	}
	return cli
}

func Split(s rune) bool {
	if s == '(' || s == ',' {
		return true
	}
	return false
}

func IsExist(value string, array []string) bool {
	for _, v := range array {
		if value == v {
			return true
		}
	}
	return false
}

func FirstIndexOf(value string, array []string) int {
	for k, v := range array {
		if v == value {
			return k
		}
	}
	return -1
}

func RunCommandInManager(managerID string, cmd []string) error {
	fmt.Println(strings.Join(cmd, " "))
	execId, err := cli.ContainerExecCreate(ctx, managerID, types.ExecConfig{
		Privileged: true,
		Cmd:        cmd,
	})
	if err != nil {
		return err
	}
	err = cli.ContainerExecStart(ctx, execId.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	return nil
}

func IsDigitAll(targetString string) bool {
	pattern := "^-?\\d+(\\.\\d+)?$"
	result, _ := regexp.MatchString(pattern, targetString)
	return result
}

func GetParamJson(req *http.Request) map[string]string {
	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	var param map[string]string
	decoder.Decode(param)
	return param
}

//func GetLogPath() string {
//	dir,err := os.Getwd()
//	if err!=nil{
//		panic("get log path failed")
//	}
//	return
//}
