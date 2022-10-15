package docker

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

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

type Containers []Container

type Client struct {
	*client.Client
}
