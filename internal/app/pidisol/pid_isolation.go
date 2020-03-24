package pidisol

import (
	"context"
	"os"
	"time"

	"controller.com/internal/app/sqlhelper"

	"controller.com/config"
	"controller.com/internal"

	"github.com/docker/docker/api/types"
)

var adoreNgPath = config.AdoreNgPath
var dstPath = config.ExecFileDstPath
var cli = internal.InitClient()
var ctx = context.Background()
var chMap map[string]chan bool
var log = config.ELog

type HiderBoss struct {
	realPidsLast  []string
	managerContID string
}

func InitMap() {
	if chMap != nil {
		return
	}
	var myDb = sqlhelper.GetNewHelper()
	defer myDb.Close()
	chMap = myDb.GetChMap()
	for mgr, _ := range chMap {
		go PidIsolation(mgr)
	}
}

func PidIsolation(managerID string) {
	newBoss := getNewBoss(managerID)
	newBoss.Isolation()
}

func getNewBoss(managerID string) *HiderBoss {
	//TODO check id
	return &HiderBoss{
		realPidsLast:  []string{},
		managerContID: managerID,
	}
}

func (boss *HiderBoss) Isolation() {
	managerID := boss.managerContID
	if _, ok := chMap[managerID]; !ok {
		chMap[managerID] = make(chan bool)
		prepareHideEnv(managerID)
	}
	timeout := make(chan bool)
	for {
		go func() {
			time.Sleep(config.SleepTime)
			timeout <- true
		}()
		select {
		case <-chMap[managerID]:
			boss.removeAdoreNg()
			close(chMap[managerID])
			<-timeout
			close(timeout)
			delete(chMap, managerID)
			return
		case <-timeout:
			boss.hideRealPids()
		}
	}
}

func SendStopToChan(managerID string) {
	chMap[managerID] <- true
	return
}

func (boss *HiderBoss) hideRealPids() {
	realPids := getRealPidsInManager(boss.managerContID)
	purePids := boss.getDiffPidsFromLast(realPids)
	if len(purePids) == 0 {
		return
	}
	boss.saveLastPids(realPids)
	cmdHide := append([]string{config.HiderPath, "hide"}, purePids...)
	err := internal.RunCommandInManager(boss.managerContID, cmdHide)
	if err != nil {
		log.Println(err)
	}
}

func (boss *HiderBoss) saveLastPids(realPids []string) {
	boss.realPidsLast = realPids
}

func (boss HiderBoss) getDiffPidsFromLast(realPids []string) []string {
	newPids := []string{}
	for _, pid := range realPids {
		if exist := internal.FirstIndexOf(pid, boss.realPidsLast); exist == -1 {
			newPids = append(newPids, pid)
		}
	}
	return newPids
}

func (boss HiderBoss) removeAdoreNg() {
	rmmodCmd := []string{config.AvaPath, "U"}
	err := internal.RunCommandInManager(boss.managerContID, rmmodCmd)
	if err != nil {
		log.Println(err)
	}
}

func prepareHideEnv(managerID string) {
	// TODO need to be refactored
	//prepareAdoreNg(managerID)
	insmodCmd := []string{"insmod", config.KoPath}
	err := internal.RunCommandInManager(managerID, insmodCmd)
	if err != nil {
		log.Println(err)
	}
	prepareHiderInManager(managerID)
}

func getRealPidsInManager(managerID string) (pids []string) {
	body, err := cli.ContainerTop(ctx, managerID, []string{})
	if err != nil {
		panic(err)
	}
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

//将controller复制到目标路径，以隐藏进程
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
