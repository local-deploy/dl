package command

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration/network"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"strings"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Down project",
	Long:  `Down project.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func down() {
	helper.LoadEnv()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	handleError(err)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	handleError(err)

	for _, container := range containers {
		containerName := strings.TrimPrefix(container.Names[0], "/")

		if !strings.Contains(containerName, helper.ProjectEnv.GetString("APP_NAME")) {
			continue
		}

		spinnerStopping, _ := pterm.DefaultSpinner.Start("Stopping and remove container " + containerName)
		err := cli.ContainerStop(ctx, container.ID, nil)

		spinnerStopping.UpdateText("Container " + containerName + " deleted")
		err = cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		})

		if err != nil {
			spinnerStopping.Fail("Error while deleting container " + containerName)
			continue
		}

		spinnerStopping.Success()
	}

	if isProjectNet(cli) {
		spinnerNetwork, _ := pterm.DefaultSpinner.Start("Deleting network")
		netFilters := filters.NewArgs(filters.Arg("name", helper.ProjectEnv.GetString("NETWORK_NAME")))
		list, err := cli.NetworkList(ctx, types.NetworkListOptions{Filters: netFilters})
		err = cli.NetworkRemove(ctx, list[0].ID)

		if err != nil {
			spinnerNetwork.Fail("Network deleting error")
			return
		}
		spinnerNetwork.UpdateText("Network deleted")
		spinnerNetwork.Success()
	}
}

func isProjectNet(cli *client.Client) bool {
	net := network.IsNetworkAvailable(cli, helper.ProjectEnv.GetString("NETWORK_NAME")+"_default")

	return net().Success()
}
