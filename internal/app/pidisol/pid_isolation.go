package pidisol

import (
	"context"
	"fmt"
	"os"
	"strings"

	"controller.com/config"
	"controller.com/internal"

	"github.com/docker/docker/api/types"
)

var adoreNgPath = config.AdoreNgPath
var dstPath = config.ExecFileDstPath
var cli = internal.InitClient()
var ctx = context.Background()

func PidIsolation(managerID string) {
	// TODO need to be refactored
	prepareAdoreNg(managerID)
	insmodCmd := []string{"insmod", config.KoPath}
	// TODO need to provide a method to remove the adore-ng from the manager
	internal.RunCommandInManager(managerID, insmodCmd)
	realPids := getRealPidsInManager(managerID)
	prepareHiderInManager(managerID)
	internal.RunCommandInManager(managerID, []string{config.HiderPath, "hide", strings.Join(realPids, " ")})
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
