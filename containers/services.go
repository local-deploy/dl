package containers

import (
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/local-deploy/dl/utils"
)

// ServicesNetworkName network name for service containers
var ServicesNetworkName = "dl_default"

// Traefik local traefik
func Traefik() types.ServiceConfig {
	return types.ServiceConfig{
		Name:          "traefik",
		Image:         "traefik:v2.10.7",
		ContainerName: "traefik",
		Command: types.ShellCommand{
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
		Environment: nil,
		Ports: []types.ServicePortConfig{
			{Mode: "ingress", HostIP: "0.0.0.0", Target: 80, Published: "80", Protocol: "tcp"},
			{Mode: "ingress", HostIP: "0.0.0.0", Target: 443, Published: "443", Protocol: "tcp"},
			{Mode: "ingress", HostIP: "0.0.0.0", Target: 8081, Published: "8081", Protocol: "tcp"},
			{Mode: "ingress", HostIP: "0.0.0.0", Target: 8082, Published: "8082", Protocol: "tcp"},
		},
		Labels: types.Labels{
			"traefik.enable":                                         "true",
			"com.docker.compose.project":                             "dl-services",
			"traefik.http.routers.traefik.entrypoints":               "web, websecure",
			"traefik.http.routers.traefik.rule":                      "Host(`traefik.localhost`) || HostRegexp(`traefik.{ip:.*}.nip.io`)",
			"traefik.http.services.traefik.loadbalancer.server.port": "8080",
			"traefik.http.middlewares.site-compress.compress":        "true",
			"traefik.http.routers.traefik.middlewares":               "site-compress",
		},
		Networks: map[string]*types.ServiceNetworkConfig{
			ServicesNetworkName: nil,
		},
		Scale:      intPtr(1),
		PullPolicy: types.PullPolicyIfNotPresent,
		Restart:    types.RestartPolicyAlways,
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:     types.VolumeTypeBind,
				Source:   "/var/run/docker.sock",
				Target:   "/var/run/docker.sock",
				ReadOnly: true,
			},
			{
				Type:     types.VolumeTypeBind,
				Source:   filepath.Join(utils.ConfigDir(), "certs"),
				Target:   "/certs",
				ReadOnly: true,
			},
		},
	}
}

// Mail local mailhog
func Mail() types.ServiceConfig {
	return types.ServiceConfig{
		Name:          "mail",
		Image:         "mailhog/mailhog",
		ContainerName: "mail",
		Labels: types.Labels{
			"com.docker.compose.project":                          "dl-services",
			"traefik.enable":                                      "true",
			"traefik.http.routers.mail.entrypoints":               "web, websecure",
			"traefik.http.routers.mail.rule":                      "Host(`mail.localhost`) || HostRegexp(`mail.{ip:.*}.nip.io`)",
			"traefik.http.services.mail.loadbalancer.server.port": "8025",
		},
		Ports: []types.ServicePortConfig{
			{
				Mode:      "ingress",
				HostIP:    "0.0.0.0",
				Target:    1025,
				Published: "1025",
				Protocol:  "tcp",
			},
		},
		Networks: map[string]*types.ServiceNetworkConfig{
			ServicesNetworkName: {},
		},
		Scale:      intPtr(1),
		PullPolicy: types.PullPolicyIfNotPresent,
		Restart:    types.RestartPolicyAlways,
	}
}

// Portainer local portainer
func Portainer() types.ServiceConfig {
	return types.ServiceConfig{
		Name:          "portainer",
		Image:         "portainer/portainer-ce",
		ContainerName: "portainer",
		Command: types.ShellCommand{
			"--no-analytics",
		},
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:     types.VolumeTypeBind,
				Source:   "/var/run/docker.sock",
				Target:   "/var/run/docker.sock",
				ReadOnly: true,
			},
			{
				Type:     types.VolumeTypeVolume,
				Source:   "portainer_data",
				Target:   "/data",
				ReadOnly: false,
			},
		},
		Labels: map[string]string{
			"com.docker.compose.project":                               "dl-services",
			"traefik.enable":                                           "true",
			"traefik.http.routers.portainer.entrypoints":               "web, websecure",
			"traefik.http.routers.portainer.rule":                      "Host(`portainer.localhost`) || HostRegexp(`portainer.{ip:.*}.nip.io`)",
			"traefik.http.services.portainer.loadbalancer.server.port": "9000",
		},
		Networks: map[string]*types.ServiceNetworkConfig{
			ServicesNetworkName: {},
		},
		Scale:      intPtr(1),
		PullPolicy: types.PullPolicyIfNotPresent,
		Restart:    types.RestartPolicyAlways,
	}
}

func intPtr(i int) *int {
	return &i
}
