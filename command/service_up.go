package command

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"
)

func init() {
	serviceCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start local services",
	Long:  `Start local services.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func up() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	home, err := os.UserHomeDir()
	portainerDataDir := filepath.Join(home, ".dl/portainer_data")
	if _, err := os.Stat(portainerDataDir); err != nil {
		if os.IsNotExist(err) {
			_ = os.Mkdir(portainerDataDir, 0755)
		}
	}

	handleError(err)

	if isNotNet(cli) {
		_, err = cli.NetworkCreate(ctx, localNetworkName, types.NetworkCreate{})
	}

	localContainers := getServicesContainer()

	for _, local := range localContainers {
		// Check running containers
		containerFilter := filters.NewArgs(filters.Arg("name", local.Name))
		isExists, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilter})
		if len(isExists) > 0 {
			fmt.Print("Restarting container ", local.Name, "... ")

			err := cli.ContainerRestart(ctx, isExists[0].ID, nil)
			handleError(err)

			fmt.Println("Success")

			continue
		}

		// Check ports
		ports := local.PortBindings
		busyPort := false
		for _, port := range ports {
			conn, _ := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", port[0].HostPort), time.Second)
			if conn != nil {
				defer func(conn net.Conn) {
					_ = conn.Close()
				}(conn)
				busyPort = true
				fmt.Printf("Unable to start container %s: port %s is busy.\n", local.Name, port[0].HostPort)
			}
		}
		if busyPort {
			continue
		}

		// Check for images
		imageFiler := filters.NewArgs(filters.Arg("reference", local.Image+":"+local.Version))
		isImageExists, err := cli.ImageList(ctx, types.ImageListOptions{All: true, Filters: imageFiler})
		if len(isImageExists) == 0 {
			fmt.Print("Pulling image ", local.Image, "... ")

			out, err := cli.ImagePull(ctx, local.Image+":"+local.Version, types.ImagePullOptions{})
			_, err = ioutil.ReadAll(out)
			handleError(err)

			fmt.Println("Success")
		}

		fmt.Print("Starting container ", local.Name, "... ")

		// Create containers
		resp, err := cli.ContainerCreate(ctx,
			&container.Config{
				Cmd:          local.Cmd,
				Image:        local.Image,
				Volumes:      local.Volumes,
				Entrypoint:   local.Entrypoint,
				Labels:       local.Labels,
				ExposedPorts: local.Ports,
				Env:          local.Env,
			},
			&container.HostConfig{
				NetworkMode:   container.NetworkMode(localNetworkName),
				RestartPolicy: container.RestartPolicy{Name: "always"},
				PortBindings:  local.PortBindings,
				Mounts:        local.Mounts,
			}, nil, nil, local.Name)
		handleError(err)

		// Start containers
		err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		handleError(err)

		fmt.Println("Success")
	}

}
