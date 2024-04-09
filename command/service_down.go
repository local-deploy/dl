package command

import (
	"context"
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/local-deploy/dl/containers"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/spf13/cobra"
)

func downServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop and remove services",
		Long: `Stops and removes portainer, mailcatcher and traefik containers.
Valid parameters for the "--service" flag: portainer, mail, traefik`,
		Example: "dl down\ndl down -s portainer",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			err := downServiceRun(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

func downServiceRun(ctx context.Context) error {
	client, _ := docker.NewClient()
	checkOldNetwork(ctx, client)

	services := types.Services{}
	servicesContainers := getServicesContainer()
	for _, service := range servicesContainers {
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

	timeoutValue := 30 * time.Second
	err := client.Backend.Down(ctx, project.Name, api.DownOptions{
		RemoveOrphans: false,
		Project:       project,
		Timeout:       &timeoutValue,
		Volumes:       false,
	})
	if err != nil {
		return err
	}

	return nil
}
