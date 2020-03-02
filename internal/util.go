package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"controller.com/config"
	"github.com/docker/docker/api/types"

	"github.com/docker/docker/client"
)

func PrintTitle() {
	fl := false
	for _, ch := range config.Title {
		char := string(ch)
		color := 32
		fmtHL := 0
		highLight := (char == "|" || char == "/" || char == "\\")
		bracket := (char == "(" || char == ")")
		if highLight && fl {
			fmtHL = 1
			fl = !fl
		} else if highLight {
			color = 34
			fl = !fl
		} else if bracket {
			color = 37
		} else {
			color = 34
			if char == "\n" {
				fl = false
			}
		}
		fmt.Printf("%c[%d;48;%dm%s%c[0m", 0x1B, fmtHL, color, char, 0x1B)
	}
	fmt.Print("\n")
}

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

func GetParamJson(req *http.Request) (map[string]string, error) {
	var param map[string]string
	err := json.NewDecoder(req.Body).Decode(&param)
	defer req.Body.Close()
	return param, err
}

//func GetLogPath() string {
//	dir,err := os.Getwd()
//	if err!=nil{
//		panic("get log path failed")
//	}
//	return
//}
