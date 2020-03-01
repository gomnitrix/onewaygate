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
	//for i:=0;i<len(config.Title);i++{
	//	if string(config.Title[i])=="|"{
	//		fmt.Printf("\n %c[1;48;32m%s%c[0m\n\n", 0x1B, config.Title[i], 0x1B)
	//	}else {
	//		fmt.Printf("\n %c[0;48;32m%s%c[0m\n\n", 0x1B, config.Title[i], 0x1B)
	//	}
	//}
	fmt.Printf("\n %c[1;48;32m%s%c[0m\n\n", 0x1B, config.Title, 0x1B)
	//其中0x1B是标记，[开始定义颜色，1代表高亮，48代表黑色背景，32代表绿色前景，0代表恢复默认颜色。
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
