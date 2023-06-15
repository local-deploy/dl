package command

import (
	"path/filepath"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration/network"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/spf13/cobra"
)

var source string
var servicesNetworkName = "dl_default"

var serviceCmd = &cobra.Command{
	Use:       "service",
	Short:     "Local services configuration",
	Long:      `Local services configuration (portainer, mailcatcher, traefik).`,
	ValidArgs: []string{"up", "down", "recreate", "restart"},
}

func serviceCommand() *cobra.Command {
	serviceCmd.AddCommand(
		downServiceCommand(),
		recreateServiceCommand(),
		restartServiceCommand(),
		upServiceCommand(),
	)
	return serviceCmd
}

// getServicesContainer local services containers
func getServicesContainer() []docker.Container {
	containers := []docker.Container{
		{
			Name:    "traefik",
			Image:   "traefik",
			Version: "latest",
			Cmd: []string{
				"--api.insecure=true",
				"--providers.docker",
				"--providers.docker.network=dl_default",
				"--providers.docker.exposedbydefault=false",
				"--providers.file.directory=/certs/conf",
				"--entrypoints.web.address=:80",
				"--entrypoints.websecure.address=:443",
				"--entrypoints.ws.address=:8081",
				"--entrypoints.wss.address=:8082",
				"--serversTransport.insecureSkipVerify=true",
			},
			Volumes: map[string]struct{}{"/var/run/docker.sock": {}},
			Labels: map[string]string{
				"traefik.enable":                                         "true",
				"com.docker.compose.project":                             "dl-services",
				"traefik.http.routers.traefik.entrypoints":               "web, websecure",
				"traefik.http.routers.traefik.rule":                      "Host(`traefik.localhost`) || HostRegexp(`traefik.{ip:.*}.nip.io`)",
				"traefik.http.services.traefik.loadbalancer.server.port": "8080",
				"traefik.http.middlewares.site-compress.compress":        "true",
				"traefik.http.routers.traefik.middlewares":               "site-compress",
			},
			Ports: []string{"0.0.0.0:80:80", "0.0.0.0:443:443", "0.0.0.0:8081:8081", "0.0.0.0:8082:8082"},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/run/docker.sock",
					Target:   "/var/run/docker.sock",
					ReadOnly: true,
				},
				{
					Type:     mount.TypeBind,
					Source:   filepath.Join(helper.ConfigDir(), "certs"),
					Target:   "/certs",
					ReadOnly: true,
				},
			},
			Env:     nil,
			Network: servicesNetworkName,
		},
		{
			Name:       "mail",
			Image:      "mailhog/mailhog",
			Version:    "latest",
			Cmd:        nil,
			Volumes:    nil,
			Entrypoint: nil,
			Labels: map[string]string{
				"com.docker.compose.project":                          "dl-services",
				"traefik.enable":                                      "true",
				"traefik.http.routers.mail.entrypoints":               "web, websecure",
				"traefik.http.routers.mail.rule":                      "Host(`mail.localhost`) || HostRegexp(`mail.{ip:.*}.nip.io`)",
				"traefik.http.services.mail.loadbalancer.server.port": "8025",
			},
			Ports:   []string{"0.0.0.0:1025:1025"},
			Network: servicesNetworkName,
		},
		{
			Name:    "portainer",
			Image:   "portainer/portainer-ce",
			Version: "latest",
			Cmd:     []string{"--no-analytics"},
			Volumes: map[string]struct{}{
				"/var/run/docker.sock:/var/run/docker.sock": {},
			},
			Labels: map[string]string{
				"com.docker.compose.project":                               "dl-services",
				"traefik.enable":                                           "true",
				"traefik.http.routers.portainer.entrypoints":               "web, websecure",
				"traefik.http.routers.portainer.rule":                      "Host(`portainer.localhost`) || HostRegexp(`portainer.{ip:.*}.nip.io`)",
				"traefik.http.services.portainer.loadbalancer.server.port": "9000",
			},
			// Ports: []string{"0.0.0.0:9000:9000"},
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
			Network: servicesNetworkName,
		},
	}

	if len(source) > 0 {
		for _, con := range containers {
			if con.Name == source {
				return []docker.Container{con}
			}
		}
	}

	return containers
}

func isNet(cli client.NetworkAPIClient) bool {
	net := network.IsNetworkAvailable(cli, servicesNetworkName)

	return net().Success()
}
