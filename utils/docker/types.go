package docker

import (
	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/docker/api/types/mount"
)

// Container contains container data needed to run
type Container struct {
	Name       string
	Image      string
	Version    string
	Cmd        []string
	Volumes    map[string]struct{}
	Entrypoint []string
	Labels     map[string]string
	Ports      []string
	Mounts     []mount.Mount
	Env        []string
	Network    string
	AddNetwork string
}

// Containers container array
type Containers []Container

// Client docker client
type Client struct {
	DockerCli command.Cli
	Backend   api.Service
}

type ContainerSummary struct {
	ID         string
	Name       string
	State      string
	Health     string
	IPAddress  string
	ExitCode   int
	Publishers PortPublishers
}

type PortPublishers []PortPublisher

type PortPublisher struct {
	URL           string
	TargetPort    int
	PublishedPort int
	Protocol      string
}
