package docker

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerClient "github.com/docker/docker/client"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/project"
	"github.com/sirupsen/logrus"
)

// TODO: refactoring this!

// UpDbContainer Run db container before dump
func UpDbContainer() error {
	ctx := context.Background()
	w := progress.ContextWriter(ctx)

	w.Event(progress.StartingEvent("Starting db container"))

	cli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv, dockerClient.WithAPIVersionNegotiation())
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Failed to connect to socket", fmt.Sprint(err)))
		return nil
	}

	site := project.Env.GetString("HOST_NAME")
	var siteDb = site + "_db"
	containerFilter := filters.NewArgs(filters.Arg("name", siteDb))
	containerExists, err := cli.ContainerList(ctx, container.ListOptions{Filters: containerFilter})

	if len(containerExists) == 0 {
		logrus.Info("db container not running")
		bin, option := helper.GetCompose()
		Args := []string{bin}
		preArgs := []string{"-p", project.Env.GetString("NETWORK_NAME"), "up", "-d", "db"}

		if len(option) > 0 {
			Args = append(Args, option)
		}

		Args = append(Args, preArgs...)

		logrus.Infof("Run command: %s, args: %s", bin, Args)
		cmdCompose := &exec.Cmd{
			Path: bin,
			Dir:  project.Env.GetString("PWD"),
			Args: Args,
			Env:  project.CmdEnv(),
		}

		err = cmdCompose.Run()
		if err != nil {
			return err
		}

		w.Event(progress.StartedEvent("Starting db container"))

		return nil
	}
	return nil
}
