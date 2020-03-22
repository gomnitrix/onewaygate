package internal

import (
	"context"
	"errors"

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

func CreateTarget() (targetID string) {
	targetID = createNewContainer("target", config.Image, nil)[0:12] //TODO fix the name
	runContainer(targetID)
	return
}

func CreateRunManager(target string) string {
	managerID := createManager(target)
	runContainer(managerID)
	return managerID[0:12]
}

func runContainer(ID string) {
	if err := cli.ContainerStart(ctx, ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func createManager(target string) (managerID string) {
	pidConfig := container.PidMode("container:" + target)
	hostConfig := &container.HostConfig{
		PidMode: pidConfig,
		Sysctls: map[string]string{"net.ipv4.ip_forward": "0"},
	}
	managerID = createNewContainer(config.ManagerPrefix+target, config.MgrImage, hostConfig)
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

//func CheckContainerID(container string) bool {
//	containers,err := cli.ContainerList(ctx,types.ContainerListOptions{})
//}
