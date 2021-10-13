package command

import (
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration/network"
	"github.com/spf13/cobra"
)

//localServicesContainer list of container names of the local stack
type localServicesContainer struct {
	Name    string
	Image   string
	Version string
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var localNetworkName = "dl_default"

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Local services configuration",
	Long:  `Local services configuration (portainer, mailcatcher, nginx).`,
}

func newLocalServicesContainer(name string, version string) *localServicesContainer {
	return &localServicesContainer{Name: name, Version: version}
}

//getServicesContainer container names
func getServicesContainer() []localServicesContainer {
	containers := []localServicesContainer{
		{Name: "dl-traefik", Image: "traefik", Version: "latest"},
		//{Name: "dl-mailcatcher", Image: "mailhog/mailhog", Version: "latest"},
		//{Name: "dl-portainer", Image: "portainer/portainer", Version: "latest"},
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
