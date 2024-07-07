package command

import (
	"context"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/local-deploy/dl/containers"
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/spf13/cobra"
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
		ValidArgs: []string{"--services", "--restart"},
	}
	cmd.Flags().BoolVarP(&recreate, "recreate", "r", false, "Recreate running containers")
	return cmd
}

func upServiceRun(ctx context.Context) error {
	if !utils.WpdeployCheck() {
		return nil
	}

	client, _ := docker.NewClient()
	checkOldNetwork(ctx, client)

	srv := types.Services{}
	servicesContainers := getServicesContainer()
	for _, service := range servicesContainers {
		if isDisabled(service.Name) {
			continue
		}
		srv[service.Name] = service
	}

	project := &types.Project{
		Name:       "dl-services",
		WorkingDir: "",
		Services:   srv,
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
