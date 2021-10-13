package command

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
	handleError(err)

	imageName := "traefik:2.5.3"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	_, err = io.Copy(os.Stdout, out)
	handleError(err)

	if isNotNet(cli) {
		_, err = cli.NetworkCreate(ctx, localNetworkName, types.NetworkCreate{})
	}

	containerFilters := filters.NewArgs(filters.Arg("name", "dl-traefik"))
	isExists, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilters})

	if len(isExists) > 0 {
		err := cli.ContainerRestart(ctx, isExists[0].ID, nil)
		handleError(err)

		fmt.Println("Container restarted")
		return
	}

	//TODO: https://tilrnt.github.io/golang/network/2017/12/30/golang-check-for-open-ports.html

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Cmd:        []string{"--api.insecure=true", " --providers.docker", " --providers.docker.network=dl_default", " --providers.docker.exposedByDefault=false"},
			Image:      imageName,
			Volumes:    map[string]struct{}{"/var/run/docker.sock": {}},
			Entrypoint: []string{"/entrypoint.sh"},
			Labels:     map[string]string{"com.docker.compose.project": "dl-services"},
			ExposedPorts: nat.PortSet{
				"8080/tcp": {},
				"80/tcp":   {},
			},
		},
		&container.HostConfig{
			Binds:         []string{"/var/run/docker.sock:/var/run/docker.sock:ro"},
			NetworkMode:   container.NetworkMode(localNetworkName),
			RestartPolicy: container.RestartPolicy{Name: "always", MaximumRetryCount: 10},
			PortBindings: nat.PortMap{
				nat.Port("8080/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
				nat.Port("80/tcp"):   []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "80"}},
			},
		}, nil, nil, "dl-traefik")
	handleError(err)

	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	handleError(err)

	fmt.Println(resp.ID)
}
