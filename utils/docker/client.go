package docker

import (
	"strings"

	"github.com/docker/docker/api/types"
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
