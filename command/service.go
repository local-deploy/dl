package command

import (
	"context"
	"slices"

	"github.com/compose-spec/compose-go/v2/types"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/local-deploy/dl/containers"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var servicesFlag []string

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
		upServiceCommand(),
	)

	serviceCmd.PersistentFlags().StringSliceVarP(&servicesFlag, "services", "s", nil, "Manage only specified services (comma separated values)")

	return serviceCmd
}

// getServicesContainer local services containers
func getServicesContainer() []types.ServiceConfig {
	configs := []types.ServiceConfig{
		containers.Traefik(),
		containers.Mail(),
		containers.Portainer(),
	}

	if len(servicesFlag) > 0 {
		var c []types.ServiceConfig
		for _, service := range configs {
			if isSetFlag(service.Name) {
				c = append(c, service)
			}
		}
		configs = c
	}

	return configs
}

// CheckOldNetwork deleting the old dl_default network created in previous versions of dl
func checkOldNetwork(ctx context.Context, client *docker.Client) {
	netFilters := filters.NewArgs(filters.Arg("name", "dl_default"))
	list, _ := client.DockerCli.Client().NetworkList(ctx, dockerTypes.NetworkListOptions{Filters: netFilters})
	if len(list) == 0 {
		return
	}

	inspect, err := client.DockerCli.Client().NetworkInspect(ctx, "dl_default", dockerTypes.NetworkInspectOptions{})
	if err != nil {
		return
	}

	for label, value := range inspect.Labels {
		if label == "com.docker.compose.network" && value == "dl_default" {
			return
		}
	}

	for _, con := range inspect.Containers {
		_ = client.DockerCli.Client().ContainerStop(ctx, con.Name, container.StopOptions{})
		_ = client.DockerCli.Client().ContainerRemove(ctx, con.Name, container.RemoveOptions{Force: true})
	}

	err = client.DockerCli.Client().NetworkRemove(ctx, "dl_default")
	if err != nil {
		return
	}

	pterm.FgYellow.Println("Successful removal containers of the previous version.")
}

func isDisabled(service string) bool {
	if service == "traefik" {
		return false
	}

	hasKeys := viper.IsSet("services")
	services := viper.GetStringSlice("services")
	if !hasKeys {
		services = append(services, "portainer", "mail")
	}

	index := slices.IndexFunc(services, func(v string) bool {
		return v == service
	})

	return index == -1
}

func isSetFlag(service string) bool {
	index := slices.IndexFunc(servicesFlag, func(v string) bool {
		return v == service
	})

	return index != -1
}
