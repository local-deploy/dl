package command

import (
	"context"

	"github.com/docker/compose/v2/pkg/progress"
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
			ctx := context.Background()
			err := progress.Run(ctx, downServiceRun)
			if err != nil {
				return err
			}

			return nil
		},
		ValidArgs: []string{"--service"},
	}
	cmd.Flags().StringVarP(&source, "service", "s", "", "Stop and remove single service")
	return cmd
}

func downServiceRun(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	serviceContainers := getServicesContainer()
	cli, err := docker.NewClient()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Docker", "Failed connect to socket"))
		return err
	}

	err = cli.RemoveContainers(ctx, serviceContainers)
	if err != nil {
		return err
	}

	if cli.IsNetworkAvailable(servicesNetworkName) && len(source) == 0 {
		err := cli.RemoveNetwork(ctx, servicesNetworkName)
		if err != nil {
			return err
		}
	}

	return err
}
