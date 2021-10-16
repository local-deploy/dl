package command

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	serviceCmd.AddCommand(downCmd)
	downCmd.Flags().StringVarP(&source, "service", "s", "", "Stop and remove single service")
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove services",
	Long:  `Stop and remove services.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func down() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	handleError(err)

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
		return
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilters})
	handleError(err)

	for _, container := range containers {
		fmt.Print("Stopping container ", strings.TrimPrefix(container.Names[0], "/"), "... ")
		err := cli.ContainerStop(ctx, container.ID, nil)
		err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})

		handleError(err)
		fmt.Println("Success")
	}

	if isNet(cli) && len(source) == 0 {
		netFilters := filters.NewArgs(filters.Arg("name", localNetworkName))
		list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: netFilters})
		err = cli.NetworkRemove(ctx, list[0].ID)

		handleError(err)
	}
}
