package command

import (
	"context"
	"fmt"
	"github.com/docker/docker/integration/network"
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
	if err != nil {
		panic(err)
	}

	imageName := "traefik:latest"

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)

	isNet := network.IsNetworkNotAvailable(cli, "traefik_default")
	if isNet().Success() {
		_, err = cli.NetworkCreate(ctx, "traefik_default", types.NetworkCreate{})
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		ExposedPorts: nil,
		Env:          nil,
		Cmd:          []string{"--api.insecure=true", " --providers.docker", " --providers.docker.network=traefik_default", " --providers.docker.exposedByDefault=false"},
		Image:        imageName,
		Volumes: map[string]struct{}{
			"/var/run/docker.sock": {},
		},
		WorkingDir: "",
		Entrypoint: []string{"/entrypoint.sh"},
		OnBuild:    nil,
		Labels:     nil,
	}, &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock:ro",
		},
		NetworkMode: "traefik_default",
		IpcMode:     "private",
		PortBindings: nat.PortMap{
			nat.Port("8080/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		AutoRemove:      false,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		Links:           nil,
		Privileged:      false,
		PublishAllPorts: false,
	}, nil, nil, "dl-traefik")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)
}
