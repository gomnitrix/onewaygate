package internal

import (
	"context"

	"controller.com/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
)

var cli = InitClient()
var ctx = context.Background()
var image = config.Image

func createNewContainer(containerName string, hostConfig *container.HostConfig) string {
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
	targetID = createNewContainer("target", nil)[0:12] //TODO fix the name
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
	}
	managerID = createNewContainer(config.ManagerPrefix+target, hostConfig)
	return
}

/*
	identity: could be the name or the brief Id of a container
*/
func GetContainerFullID(identity string) (string, error) {
	containerConfig, err := cli.ContainerInspect(ctx, identity)
	if err != nil {
		return "", err
	}
	return containerConfig.ID, nil
}

func RmContainer(identity string) error {
	return cli.ContainerRemove(ctx, identity, types.ContainerRemoveOptions{})
}

//func CheckContainerID(container string) bool {
//	containers,err := cli.ContainerList(ctx,types.ContainerListOptions{})
//}
