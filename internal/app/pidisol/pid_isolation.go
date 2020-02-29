package pidisol

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"controller.com/config"
	"controller.com/internal"

	"github.com/docker/docker/api/types"
)

var adoreNgPath = config.AdoreNgPath
var dstPath = config.ExecFileDstPath
var cli = internal.InitClient()
var ctx = context.Background()
var chMap map[string](chan bool)

type HiderBoss struct {
	realPidsLast  [config.MaxPrePidsNum]string
	managerContID string
}

func PidIsolation(managerID string) {
	newBoss := getNewBoss(managerID)
	newBoss.Isolation()
}

func getNewBoss(managerID string) *HiderBoss {
	//TODO check id
	return &HiderBoss{
		realPidsLast:  [5]string{},
		managerContID: managerID,
	}
}

func (boss *HiderBoss) Isolation() {
	managerID := boss.managerContID
	prepareHideEnv(managerID)
	timeout := make(chan bool)
	for {
		go func() {
			time.Sleep(config.SleepTime)
			timeout <- true
		}()
		select {
		case <-chMap[managerID]:

			return
		case <-timeout:
			realPids := getRealPidsInManager(managerID)
			internal.RunCommandInManager(managerID, []string{config.HiderPath, "hide", strings.Join(realPids, " ")})
		}
	}
}
func (boss HiderBoss) removeAdoreNg() {
	rmmodCmd := []string{config.AvaPath, "U"}
	err := internal.RunCommandInManager(boss.managerContID, rmmodCmd)
	if err != nil {

	}
}

func prepareHideEnv(managerID string) {
	// TODO need to be refactored
	prepareAdoreNg(managerID)
	insmodCmd := []string{"insmod", config.KoPath}
	// TODO need to provide a method to remove the adore-ng from the manager
	internal.RunCommandInManager(managerID, insmodCmd)
	prepareHiderInManager(managerID)
}

func getRealPidsInManager(managerID string) (pids []string) {
	body, err := cli.ContainerTop(ctx, managerID, []string{})
	if err != nil {
		panic(err)
	}
	fmt.Println("top done")
	fmt.Println(body)
	index := internal.FirstIndexOf("PID", body.Titles)
	pids = make([]string, 0, config.MaxPrePidsNum)
	for _, v := range body.Processes {
		pids = append(pids, v[index])
	}
	return
}

func prepareHiderInManager(managerID string) {
	hiderReader := getHider()
	err := cli.CopyToContainer(ctx, managerID, dstPath, hiderReader, types.CopyToContainerOptions{})
	hiderReader.Close()
	if err != nil {
		panic(err)
	}
}

func getHider() (reader *os.File) {
	reader, err := os.Open(config.HiderTarPath)
	if err != nil {
		panic(err)
	}
	return
}
func getAdoreNg() (reader *os.File) {
	reader, err := os.Open(adoreNgPath)
	if err != nil {
		panic(err)
	}
	return
}

func prepareAdoreNg(managerID string) {
	adoreReader := getAdoreNg()
	err := cli.CopyToContainer(ctx, managerID, dstPath, adoreReader, types.CopyToContainerOptions{})
	adoreReader.Close()
	if err != nil {
		panic(err)
	}
}
