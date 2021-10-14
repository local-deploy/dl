package command

import (
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration/network"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

//localServicesContainer list of container names of the local stack
type localServicesContainer struct {
	Name         string
	Image        string
	Version      string
	Cmd          []string
	Volumes      map[string]struct{}
	Entrypoint   []string
	Labels       map[string]string
	Ports        nat.PortSet
	Mounts       []mount.Mount
	PortBindings nat.PortMap
	Env          []string
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var localNetworkName = "dl_default"

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Local services configuration",
	Long:  `Local services configuration (portainer, mailcatcher, traefik).`,
}

//getServicesContainer local services containers
func getServicesContainer() []localServicesContainer {
	home, err := os.UserHomeDir()
	portainerDataDir := filepath.Join(home, ".dl/portainer_data")

	handleError(err)

	defaultLabels := map[string]string{"com.docker.compose.project": "dl-services"}
	containers := []localServicesContainer{
		{
			Name:       "dl-traefik",
			Image:      "traefik",
			Version:    "latest",
			Cmd:        []string{"--api.insecure=true", "--providers.docker", "--providers.docker.network=dl_default", "--providers.docker.exposedByDefault=false"},
			Volumes:    map[string]struct{}{"/var/run/docker.sock": {}},
			Entrypoint: []string{"/entrypoint.sh"},
			Labels:     defaultLabels,
			Ports: nat.PortSet{
				"8080/tcp": {},
				"80/tcp":   {},
			},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/run/docker.sock",
					Target:   "/var/run/docker.sock",
					ReadOnly: true,
				},
			},
			PortBindings: nat.PortMap{
				nat.Port("8080/tcp"): []nat.PortBinding{{HostPort: "8080"}},
				nat.Port("80/tcp"):   []nat.PortBinding{{HostPort: "80"}},
			},
			Env: nil,
		},
		{
			Name:       "dl-mailcatcher",
			Image:      "mailhog/mailhog",
			Version:    "latest",
			Cmd:        nil,
			Volumes:    nil,
			Entrypoint: nil,
			Labels:     defaultLabels,
			Ports: nat.PortSet{
				"8025/tcp": {},
				"1025/tcp": {},
			},
			Mounts: nil,
			PortBindings: nat.PortMap{
				nat.Port("8025/tcp"): []nat.PortBinding{{HostPort: "8025"}},
				nat.Port("1025/tcp"): []nat.PortBinding{{HostPort: "1025"}},
			},
			Env: nil,
		},
		{
			Name:       "dl-portainer",
			Image:      "portainer/portainer",
			Version:    "latest",
			Cmd:        []string{"--no-analytics"},
			Volumes:    map[string]struct{}{"/var/run/docker.sock:/var/run/docker.sock": {}},
			Entrypoint: nil,
			Labels:     defaultLabels,
			Ports: nat.PortSet{
				"9000/tcp": {},
			},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   "/var/run/docker.sock",
					Target:   "/var/run/docker.sock",
					ReadOnly: true,
				},
				{
					Type:     mount.TypeBind,
					Source:   portainerDataDir,
					Target:   "/data",
					ReadOnly: false,
				},
			},
			PortBindings: nat.PortMap{
				nat.Port("9000/tcp"): []nat.PortBinding{{HostPort: "9000"}},
			},
			Env: []string{"VIRTUAL_HOST=portainer.localhost"},
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
