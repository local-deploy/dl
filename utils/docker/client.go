package docker

import (
	"github.com/docker/docker/client"
)

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
