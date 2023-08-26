package docker

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
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
	*client.Client
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
