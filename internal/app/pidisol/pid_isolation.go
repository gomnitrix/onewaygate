package pidisol

import (
	"context"
	"os"
	"time"

	"controller.com/internal/OwmError"
	"github.com/kataras/golog"

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

type HiderBoss struct {
	realPidsLast  []string
	managerContID string
}

func InitMap(mylog *golog.Logger) {
	if chMap != nil {
		return
	}
	var myDb = sqlhelper.GetNewHelper()
	defer myDb.Close()
	chMap = myDb.GetChMap()
	for mgr, _ := range chMap {
		go PidIsolation(mgr, mylog)
	}
}

func PidIsolation(managerID string, mylog *golog.Logger) {
	defer HandlerRecoverForPidIsol(mylog)
	newBoss := getNewBoss(managerID)
	newBoss.Isolation()
}

func getNewBoss(managerID string) *HiderBoss {
	return &HiderBoss{
		realPidsLast:  []string{},
		managerContID: managerID,
	}
}

func (boss *HiderBoss) Isolation() {
	defer OwmError.Pack()
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
	defer OwmError.Pack()
	realPids := getRealPidsInManager(boss.managerContID)
	purePids := boss.getDiffPidsFromLast(realPids)
	if len(purePids) == 0 {
		return
	}
	boss.saveLastPids(realPids)
	cmdHide := append([]string{config.HiderPath, "hide"}, purePids...)
	internal.RunCommandInManager(boss.managerContID, cmdHide)
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
	defer OwmError.Pack()
	rmmodCmd := []string{config.AvaPath, "U"}
	internal.RunCommandInManager(boss.managerContID, rmmodCmd)
}

func prepareHideEnv(managerID string) {
	defer OwmError.Pack()
	insmodCmd := []string{"insmod", config.KoPath}
	internal.RunCommandInManager(managerID, insmodCmd)
	prepareHiderInManager(managerID)
}

func getRealPidsInManager(managerID string) (pids []string) {
	defer OwmError.Pack()
	body, err := cli.ContainerTop(ctx, managerID, []string{})
	OwmError.Check(err, "Get Container Top process failed.\n")
	index := internal.FirstIndexOf("PID", body.Titles)
	pids = make([]string, 0, config.MaxPrePidsNum)
	for _, v := range body.Processes {
		pids = append(pids, v[index])
	}
	return
}

func prepareHiderInManager(managerID string) {
	defer OwmError.Pack()
	hiderReader := getHider()
	defer hiderReader.Close()
	err := cli.CopyToContainer(ctx, managerID, dstPath, hiderReader, types.CopyToContainerOptions{})
	OwmError.Check(err, "Copy file to Container failed.\n")
}

//将controller复制到目标路径，以隐藏进程
func getHider() (reader *os.File) {
	defer OwmError.Pack()
	reader, err := os.Open(config.HiderTarPath)
	OwmError.Check(err, "Open Tar file path failed.\n")
	return
}

func HandlerRecoverForPidIsol(flog *golog.Logger) {
	if p := recover(); p != nil {
		if e, ok := p.(OwmError.Error); ok {
			flog.Errorf("%+v", e.Prev)
		} else {
			flog.Error(p)
		}
		panic(p) //使程序崩溃 避免在错误下继续运行
	}
}
