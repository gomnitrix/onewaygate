package internal

import (
	"context"
	"errors"

	"controller.com/Model"

	"controller.com/internal/OwmError"

	"controller.com/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

var cli = InitClient()
var ctx = context.Background()

func createNewContainer(containerName, image string, hostConfig *container.HostConfig) string {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Cmd:   []string{"/bin/bash"},
		Image: image,
		Tty:   true,
	}, hostConfig, nil, containerName)
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	return resp.ID
}

func CreateTarget(targetName string) (targetID string) {
	targetID = createNewContainer(targetName, config.Image, nil)[0:12] //TODO fix the name
	runContainer(targetID)
	return
}

func runContainer(ID string) {
	if err := cli.ContainerStart(ctx, ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func CreateRunManager(target string, mgrName string) string {
	managerID := createManager(target, mgrName)
	runContainer(managerID)
	return managerID[0:12]
}

func createManager(target string, mgrName string) (managerID string) {
	pidConfig := container.PidMode("container:" + target)
	hostConfig := &container.HostConfig{
		PidMode: pidConfig,
		Sysctls: map[string]string{"net.ipv4.ip_forward": "0"},
	}
	managerID = createNewContainer(mgrName, config.MgrImage, hostConfig)
	return
}

/*
	identity: could be the name or the brief Id of a container
*/
func GetContainerFullID(identity string) (string, error) {
	containerConfig, err := cli.ContainerInspect(ctx, identity)
	if err == nil && containerConfig.ID != "" {
		return containerConfig.ID, nil
	}
	id, err := GetIDByName(identity)
	if id == "" {
		return "", errors.New("no such container")
	} else {
		return id, err
	}
}

func RmContainer(identity string) error {
	return cli.ContainerRemove(ctx, identity, types.ContainerRemoveOptions{
		Force: true,
	})
}

func ListContainer() ([]types.Container, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	return containers, err
}

func GetNameByID(contID string) string {
	defer OwmError.Pack()
	infos, err := cli.ContainerInspect(ctx, contID)
	OwmError.Check(err, false, "Inspect Container %s failed\n", contID)
	return infos.Name[1:]
}

func GetIDByName(name string) (string, error) {
	containers, err := ListContainer()
	if err != nil {
		return "", err
	}
	for i := 0; i < len(containers); i++ {
		curContainer := containers[i]
		for _, n := range curContainer.Names {
			if n == name {
				return curContainer.ID, nil
			}
		}
	}
	return "", errors.New("no container has this name: " + name)
}

func GetNetInfo(containerID string) (string, error) {
	infos, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", err
	}
	ipv4Addr := infos.NetworkSettings.IPAddress
	return ipv4Addr, nil
}

func GetContTableInfo(contID string, row *Model.ContainerRow) {
	defer OwmError.Pack()
	infos, err := cli.ContainerInspect(ctx, contID)
	OwmError.Check(err, false, "Get Information of Container %s Failed\n", contID)
	row.Name = infos.Name[1:]
	row.ID = infos.ID[0:12]
	row.Status = infos.State.Status
}

func FilterContainerID(identity string) string {
	defer OwmError.Pack()
	if identity == "" {
		OwmError.Check(errors.New("Empty Container Id or Name Detected\n"), false, "FilterID failed\n")
	}
	containerConfig, err := cli.ContainerInspect(ctx, identity)
	if err == nil && containerConfig.ID != "" {
		return containerConfig.ID[0:12]
	}
	id, err := GetIDByName(identity)
	OwmError.Check(err, false, "GetIDByName failed while filter ID: %s\n", identity)
	return id[0:12]
}

func GetTty(contID string) (tty types.HijackedResponse) {
	defer OwmError.Pack()
	ir, err := cli.ContainerExecCreate(ctx, contID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash"},
		Tty:          true,
	})
	OwmError.Check(err, false, "Create Exec Failed, container ID: %s\n", contID)

	tty, err = cli.ContainerExecAttach(ctx, ir.ID, types.ExecStartCheck{Detach: false, Tty: true})
	OwmError.Check(err, false, "Attach Exec Failed, container ID: %s\n", contID)
	return
}
