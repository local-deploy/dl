package command

import (
	"context"
	"slices"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/local-deploy/dl/containers"
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var recreate bool

func upServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start local services",
		Long:  `Start portainer, mailcatcher and traefik containers.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			err := upServiceRun(ctx)
			if err != nil {
				return err
			}

			// check for new version
			utils.CheckUpdates()

			return nil
		},
		ValidArgs: []string{"--service", "--restart"},
	}
	cmd.Flags().BoolVarP(&recreate, "recreate", "r", false, "Recreate running containers")
	return cmd
}

func MapsAppend[T comparable, U any](target map[T]U, source map[T]U) map[T]U {
	if target == nil {
		return source
	}
	if source == nil {
		return target
	}
	for key, value := range source {
		if _, ok := target[key]; !ok {
			target[key] = value
		}
	}
	return target
}

func upServiceRun(ctx context.Context) error {
	if !utils.WpdeployCheck() {
		return nil
	}

	client, _ := docker.NewClient()
	checkOldNetwork(ctx, client)

	services := types.Services{}
	servicesContainers := getServicesContainer()
	for _, service := range servicesContainers {
		if !isEnable(service.Name) {
			continue
		}
		services[service.Name] = service
	}

	project := &types.Project{
		Name:       "dl-services",
		WorkingDir: "",
		Services:   services,
		Networks: map[string]types.NetworkConfig{
			containers.ServicesNetworkName: {
				Name: containers.ServicesNetworkName,
			},
		},
	}

	err := client.StartContainers(ctx, project, recreate)
	if err != nil {
		return err
	}

	return nil
}

func isEnable(service string) bool {
	if service == "traefik" {
		return true
	}

	hasKeys := viper.IsSet("services")
	services := viper.GetStringSlice("services")
	if !hasKeys {
		services = append(services, "portainer", "mail")
	}

	index := slices.IndexFunc(services, func(v string) bool {
		return v == service
	})

	return index != -1
}
