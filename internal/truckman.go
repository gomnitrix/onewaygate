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

func creatNewContainer(containerName string, hostConfig *container.HostConfig) string {
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
	targetID = creatNewContainer("target", nil)[0:12] //TODO fix the name
	runContainer(targetID)
	return
}

func CreatRunManager(target string) string {
	managerID := creatManager(target)
	runContainer(managerID)
	return managerID[0:12]
}

func runContainer(ID string) {
	if err := cli.ContainerStart(ctx, ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
}

func creatManager(target string) (managerID string) {
	pidConfig := container.PidMode("container:" + target)
	hostConfig := &container.HostConfig{
		PidMode: pidConfig,
	}
	managerID = creatNewContainer("manager_for_"+target, hostConfig)
	return
}
