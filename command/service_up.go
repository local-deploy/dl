package command

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
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
		ports := local.Ports
		busyPort := false
		for _, port := range ports {
			rawIP, hostPort, _ := splitParts(port)

			conn, _ := net.DialTimeout("tcp", net.JoinHostPort(rawIP, hostPort), time.Second)
			if conn != nil {
				defer func(conn net.Conn) {
					_ = conn.Close()
				}(conn)
				busyPort = true
				fmt.Printf("Unable to start container %s: port %s is busy.\n", local.Name, hostPort)
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
		exposedPorts, portBindings, _ := nat.ParsePortSpecs(local.Ports)

		resp, err := cli.ContainerCreate(ctx,
			&container.Config{
				Cmd:          local.Cmd,
				Image:        local.Image,
				Volumes:      local.Volumes,
				Entrypoint:   local.Entrypoint,
				Labels:       local.Labels,
				ExposedPorts: exposedPorts,
				Env:          local.Env,
			},
			&container.HostConfig{
				NetworkMode:   container.NetworkMode(localNetworkName),
				RestartPolicy: container.RestartPolicy{Name: "always"},
				PortBindings:  portBindings,
				Mounts:        local.Mounts,
			}, nil, nil, local.Name)
		handleError(err)

		// Start containers
		err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
		handleError(err)

		fmt.Println("Success")
	}

}

func splitParts(rawPort string) (string, string, string) {
	parts := strings.Split(rawPort, ":")
	n := len(parts)
	containerPort := parts[n-1]

	switch n {
	case 1:
		return "", "", containerPort
	case 2:
		return "", parts[0], containerPort
	case 3:
		return parts[0], parts[1], containerPort
	default:
		return strings.Join(parts[:n-2], ":"), parts[n-2], containerPort
	}
}
