package command

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/docker"
	"github.com/spf13/cobra"
)

var restart bool

func upServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start local services",
		Long:  `Start portainer, mailcatcher and traefik containers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			err := progress.Run(ctx, upServiceRun)
			if err != nil {
				return err
			}

			// check for new version
			utils.CheckUpdates()

			return nil
		},
		ValidArgs: []string{"--service", "--restart"},
	}
	cmd.Flags().StringVarP(&source, "service", "s", "", "Start single service")
	cmd.Flags().BoolVarP(&restart, "restart", "r", false, "Restart running containers")
	return cmd
}

func upServiceRun(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	if !helper.WpdeployCheck() {
		return nil
	}

	serviceContainers := getServicesContainer()
	cli, err := docker.NewClient()
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Docker", "Failed connect to socket"))
		return err
	}

	// Check for images
	err = cli.PullRequiredImages(ctx, serviceContainers)
	if err != nil {
		return err
	}

	// Check network
	if cli.IsNetworkNotAvailable(servicesNetworkName) {
		err := cli.CreateNetwork(ctx, servicesNetworkName)
		if err != nil {
			return err
		}
	}

	// Create portainer data volume
	volumeResponse, err := cli.VolumeList(ctx, filters.NewArgs(filters.Arg("name", "portainer_data")))

	//goland:noinspection GoNilness
	if len(volumeResponse.Volumes) == 0 {
		eventName := fmt.Sprintf("Volume %q", "portainer_data")
		w.Event(progress.CreatingEvent(eventName))
		_, err = cli.VolumeCreate(ctx, volume.VolumeCreateBody{Name: "portainer_data", Driver: "local"})
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Volume", fmt.Sprint(err)))
			return nil
		}
		w.Event(progress.CreatedEvent(eventName))
	}

	err = cli.StartContainers(ctx, serviceContainers, restart)
	if err != nil {
		return err
	}

	return err
}
