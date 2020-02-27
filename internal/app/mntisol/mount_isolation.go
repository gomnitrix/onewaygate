package mntisol

import (
	"context"
	"strings"

	"github.com/codeskyblue/go-sh"

	"controller.com/config"
	"controller.com/internal"
)

var adoreNgPath = config.AdoreNgPath
var dstPath = config.ExecFileDstPath
var cli = internal.InitClient()
var ctx = context.Background()

func MountIsolation(managerID, targetID string) {
	managerID = internal.GetContainerFullID(managerID)
	targetID = internal.GetContainerFullID(targetID)
	managerMntPath := getAufsLayerOfCont(managerID)
	targetMntPath := getAufsLayerOfCont(targetID)
	dstMntPath := managerMntPath + config.MntDstPath
	mountTargetFs(targetMntPath, dstMntPath)
}

func getAufsLayerOfCont(containerID string) (layerPath string) {
	mntID := getMountIdOfCont(containerID)
	layerPath = config.ContLayerPath + mntID
	return
}

func mountTargetFs(targetFs, dstMntPath string) {
	err := sh.Command("sudo", "-S", "mount", "--bind", targetFs, dstMntPath).SetInput(config.RootPw).Run()
	if err != nil {
		panic(err)
	}
}

func getMountIdOfCont(containerID string) (mntID string) {
	out, err := sh.Command("sudo", "-S", "cat", config.MountIDPath+containerID+"/mount-id").SetInput(config.RootPw).Output()
	if err != nil {
		panic(err)
	}
	mntID = strings.Trim(string(out), "%")
	return
}
