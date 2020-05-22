package internal

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"

	"controller.com/internal/OwmError"

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

func GetProjRoot() string {
	_, cur, _, ok := runtime.Caller(1)
	if !ok {
		panic(errors.New("Can not get current file info"))
	}
	root := filepath.Dir(filepath.Dir(cur))
	return root
}

func JoinPath(relaPath string) string {
	return filepath.Join(GetProjRoot(), relaPath)
}

//func GetParamJson(ctx iris.Context) (map[string]string, error) {
//	var param map[string]string
//	err := json.NewDecoder(ctx.Body).Decode(&param)
//	defer req.Body.Close()
//	return param, err
//}

func BuildPreRule(tgtIP, mgrIP string) []string {
	return []string{"-s", tgtIP, "-j", "TEE", "--gateway", mgrIP}
}

func BuildPostRule(tgtIP, mgrIP string) []string {
	return []string{"-d", tgtIP, "-j", "TEE", "--gateway", mgrIP}
}

func GetJsonData(v interface{}) string {
	defer OwmError.Pack()
	tmp, err := json.Marshal(v)
	OwmError.Check(err, false, "Json Marshal failed\n")
	return string(tmp)
}

func StoM(v interface{}) map[string]string {
	defer OwmError.Pack()
	tmp, err := json.Marshal(v)
	OwmError.Check(err, false, "Transformation failed\n")
	var tmpMap map[string]string
	err = json.Unmarshal(tmp, &tmpMap)
	OwmError.Check(err, false, "Transformation failed\n")
	return tmpMap
}

func StrToMD5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str
}
