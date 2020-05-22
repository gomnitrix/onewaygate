package mntisol

import (
	"context"
	"strings"

	"controller.com/internal/OwmError"

	"github.com/codeskyblue/go-sh"

	"controller.com/config"
	"controller.com/internal"
)

var adoreNgPath = config.AdoreNgPath
var dstPath = config.ExecFileDstPath
var cli = internal.InitClient()
var ctx = context.Background()

func MountIsolation(managerID, targetID string) {
	defer OwmError.Pack()
	managerID, err := internal.GetContainerFullID(managerID)
	OwmError.Check(err, "")
	targetID, err = internal.GetContainerFullID(targetID)
	OwmError.Check(err, "")
	managerMntPath := getAufsLayerOfCont(managerID)
	targetMntPath := getAufsLayerOfCont(targetID)
	dstMntPath := managerMntPath + config.MntDstPath
	mountTargetFs(targetMntPath, dstMntPath)
}

func UmountTarget(managerID string) {
	defer OwmError.Pack()
	managerID, err := internal.GetContainerFullID(managerID)
	OwmError.Check(err, "")
	managerMntPath := getAufsLayerOfCont(managerID)
	dstMntPath := managerMntPath + config.MntDstPath
	err = sh.Command("sudo", "-S", "umount", dstMntPath).SetInput(config.RootPw).Run()
	OwmError.Check(err, "")
}
func getAufsLayerOfCont(containerID string) (layerPath string) {
	defer OwmError.Pack()
	mntID := getMountIdOfCont(containerID)
	layerPath = config.ContLayerPath + mntID
	return
}

func mountTargetFs(targetFs, dstMntPath string) {
	defer OwmError.Pack()
	err := sh.Command("sudo", "-S", "mount", "--bind", targetFs, dstMntPath).SetInput(config.RootPw).Run()
	OwmError.Check(err, "")
}

func getMountIdOfCont(containerID string) (mntID string) {
	defer OwmError.Pack()
	out, err := sh.Command("sudo", "-S", "cat", config.MountIDPath+containerID+"/mount-id").SetInput(config.RootPw).Output()
	OwmError.Check(err, "")
	mntID = strings.Trim(string(out), "%")
	return
}
