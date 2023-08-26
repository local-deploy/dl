package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// NewClient docker client initialization
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	c := &Client{
		Client: cli,
	}

	return c, err
}

// IsServiceRunning Checking if local services running
func (cli *Client) IsServiceRunning(ctx context.Context) bool {
	containerFilter := filters.NewArgs(
		filters.Arg("name", "traefik"),
		filters.Arg("label", fmt.Sprintf("%s=%s", api.ProjectLabel, "dl-services")),
	)
	traefikExists, _ := cli.ContainerList(ctx, types.ContainerListOptions{Filters: containerFilter})

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
