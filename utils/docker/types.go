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
