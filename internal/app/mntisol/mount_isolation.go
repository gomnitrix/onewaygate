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
var log = config.ELog

func MountIsolation(managerID, targetID string) {
	managerMntPath := getAufsLayerOfCont(managerID)
	targetMntPath := getAufsLayerOfCont(targetID)
	dstMntPath := managerMntPath + config.MntDstPath
	mountTargetFs(targetMntPath, dstMntPath)
}

func UmountTarget(managerID string) {
	managerMntPath := getAufsLayerOfCont(managerID)
	dstMntPath := managerMntPath + config.MntDstPath
	err := sh.Command("sudo", "-S", "umount", dstMntPath).SetInput(config.RootPw).Run()
	if err != nil {
		log.Println(err)
	}
}
func getAufsLayerOfCont(containerID string) (layerPath string) {
	mntID := getMountIdOfCont(containerID)
	layerPath = config.ContLayerPath + mntID
	return
}

func mountTargetFs(targetFs, dstMntPath string) {
	err := sh.Command("sudo", "-S", "mount", "--bind", targetFs, dstMntPath).SetInput(config.RootPw).Run()
	if err != nil {
		log.Println(err)
	}
}

func getMountIdOfCont(containerID string) (mntID string) {
	out, err := sh.Command("sudo", "-S", "cat", config.MountIDPath+containerID+"/mount-id").SetInput(config.RootPw).Output()
	if err != nil {
		log.Println(err)
	}
	mntID = strings.Trim(string(out), "%")
	return
}
