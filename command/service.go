package command

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration/network"
	"github.com/spf13/cobra"
)

//localServicesContainer list of container names of the local stack
type localServicesContainer struct {
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
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var source string
var localNetworkName = "dl_default"

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Local services configuration",
	Long:  `Local services configuration (portainer, mailcatcher, traefik).`,
}

//getServicesContainer local services containers
func getServicesContainer() []localServicesContainer {
	defaultLabels := map[string]string{"com.docker.compose.project": "dl-services"}
	containers := []localServicesContainer{
		{
			Name:    "traefik",
			Image:   "traefik",
			Version: "latest",
			Cmd: []string{
				"--api.insecure=true",
				"--providers.docker",
				"--providers.docker.network=dl_default",
				"--providers.docker.exposedbydefault=false",
				"--entrypoints.web.address=:80",
				"--entrypoints.websecure.address=:443",
			},
			Volumes: map[string]struct{}{"/var/run/docker.sock": {}},
			Labels: map[string]string{
				"traefik.enable":             "true",
				"com.docker.compose.project": "dl-services",
			},
			Ports: []string{"0.0.0.0:8080:8080", "0.0.0.0:80:80", "0.0.0.0:443:443"},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/run/docker.sock",
					Target:   "/var/run/docker.sock",
					ReadOnly: true,
				},
			},
			Env: nil,
		},
		{
			Name:       "mailcatcher",
			Image:      "mailhog/mailhog",
			Version:    "latest",
			Cmd:        nil,
			Volumes:    nil,
			Entrypoint: nil,
			Labels:     defaultLabels,
			Ports:      []string{"0.0.0.0:8025:8025", "0.0.0.0:1025:1025"},
		},
		{
			Name:    "portainer",
			Image:   "portainer/portainer",
			Version: "latest",
			Cmd:     []string{"--no-analytics"},
			Volumes: map[string]struct{}{
				"/var/run/docker.sock:/var/run/docker.sock": {},
			},
			Labels: map[string]string{
				"com.docker.compose.project":                               "dl-services",
				"traefik.enable":                                           "true",
				"traefik.http.routers.portainer.entrypoints":               "web",
				"traefik.http.routers.portainer.rule":                      "Host(`portainer.localhost`)",
				"traefik.http.services.portainer.loadbalancer.server.port": "9000",
			},
			Ports: []string{"0.0.0.0:9000:9000"},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/run/docker.sock",
					Target:   "/var/run/docker.sock",
					ReadOnly: true,
				},
				{
					Type:     mount.TypeVolume,
					Source:   "portainer_data",
					Target:   "/data",
					ReadOnly: false,
				},
			},
		},
	}

	return containers
}

func isNet(cli *client.Client) bool {
	net := network.IsNetworkAvailable(cli, localNetworkName)

	return net().Success()
}

func isNotNet(cli *client.Client) bool {
	net := network.IsNetworkNotAvailable(cli, localNetworkName)

	return net().Success()
}
