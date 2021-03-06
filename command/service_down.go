package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	serviceCmd.AddCommand(downServiceCmd)
	downServiceCmd.Flags().StringVarP(&source, "service", "s", "", "Stop and remove single service")
}

var downServiceCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove services",
	Long: `Stops and removes portainer, mailcatcher and traefik containers.
Valid parameters for the "--service" flag: portainer, mail, traefik`,
	Example: "dl down\ndl down -s portainer",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		err := progress.Run(ctx, downService)
		if err != nil {
			return err
		}

		return nil
	},
	ValidArgs: []string{"--service"},
}

func downService(ctx context.Context) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	handleError(err)

	err = removeContainers(ctx, cli)
	if err != nil {
		return err
	}

	if isNet(cli) && len(source) == 0 {
		eg.Go(func() error {
			eventName := fmt.Sprintf("Network %q", localNetworkName)
			w.Event(progress.RemovingEvent(eventName))

			netFilters := filters.NewArgs(filters.Arg("name", localNetworkName))
			list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: netFilters})
			err = cli.NetworkRemove(ctx, list[0].ID)

			if err != nil {
				w.Event(progress.ErrorMessageEvent(eventName, fmt.Sprint(err)))
				return nil
			}

			w.Event(progress.RemovedEvent(eventName))
			return nil
		})
	}

	return eg.Wait()
}

func removeContainers(ctx context.Context, cli *client.Client) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	localContainers := getServicesContainer()

	containerFilters := filters.NewArgs()
	for _, container := range localContainers {
		if len(source) > 0 && source != container.Name {
			continue
		}
		containerFilters.Add("name", container.Name)
	}

	if containerFilters.Len() == 0 {
		fmt.Println("Unknown service")
		return nil
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilters})
	handleError(err)

	for _, container := range containers {
		container := container
		containerName := strings.TrimPrefix(container.Names[0], "/")

		eg.Go(func() error {
			eventName := fmt.Sprintf("Container %q", containerName)

			w.Event(progress.StoppingEvent(eventName))
			err := cli.ContainerStop(ctx, container.ID, nil)
			if err != nil {
				w.Event(progress.ErrorMessageEvent(eventName, fmt.Sprint(err)))
				return nil
			}
			w.Event(progress.StoppedEvent(eventName))

			w.Event(progress.RemovingEvent(eventName))
			err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
				// RemoveVolumes: true,
				Force: true,
			})
			if err != nil {
				w.Event(progress.ErrorMessageEvent(eventName, fmt.Sprint(err)))
				return nil
			}
			w.Event(progress.RemovedEvent(eventName))

			return nil
		})
	}

	return eg.Wait()
}
