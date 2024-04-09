package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

// NewClient docker client initialization
func NewClient() (*Client, error) {
	cli, composeService, err := newComposeService()
	if err != nil {
		return nil, err
	}
	c := &Client{
		DockerCli: cli,
		Backend:   composeService,
	}

	return c, err
}

func newComposeService() (*command.DockerCli, api.Service, error) {
	dockerCli, err := newDockerCli()
	if err != nil {
		return nil, nil, err
	}

	return dockerCli, compose.NewComposeService(dockerCli), err
}

func newDockerCli() (*command.DockerCli, error) {
	dockerCLI, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}

	options := flags.NewClientOptions()
	options.LogLevel = "fatal"

	err = dockerCLI.Initialize(options)
	if err != nil {
		return nil, err
	}

	return dockerCLI, err
}

// IsServiceRunning Checking if local services running
func (cli *Client) IsServiceRunning(ctx context.Context) bool {
	containerFilter := filters.NewArgs(
		filters.Arg("name", "traefik"),
		filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, "dl-services")),
	)
	traefikExists, _ := cli.DockerCli.Client().ContainerList(ctx, container.ListOptions{Filters: containerFilter})

	return len(traefikExists) > 0
}

// DisplayablePorts returns formatted string representing open ports of container
func (cli *Client) DisplayablePorts(c ContainerSummary) string {
	if c.Publishers == nil {
		return ""
	}

	ports := make([]types.Port, len(c.Publishers))
	for i, pub := range c.Publishers {
		ports[i] = types.Port{
			IP:          pub.URL,
			PrivatePort: uint16(pub.TargetPort),
			PublicPort:  uint16(pub.PublishedPort),
			Type:        pub.Protocol,
		}
	}

	return formatter.DisplayablePorts(ports)
}

func getContainerProgressName(c types.Container) string {
	return "Container " + GetCanonicalContainerName(c)
}

func GetCanonicalContainerName(c types.Container) string {
	if len(c.Names) == 0 {
		return c.ID[:12]
	}

	for _, name := range c.Names {
		if strings.LastIndex(name, "/") == 0 {
			return name[1:]
		}
	}
	return c.Names[0][1:]
}
