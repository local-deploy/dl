package command

import (
	"context"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(upServiceCmd)
	upServiceCmd.Flags().StringVarP(&source, "service", "s", "", "Start single service")
	upServiceCmd.Flags().BoolVarP(&restart, "restart", "r", false, "Restart running containers")
}

var restart bool
var upServiceCmd = &cobra.Command{
	Use:   "up",
	Short: "Start local services",
	Long:  `Start portainer, mailcatcher and traefik containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		upService()
	},
	ValidArgs: []string{"--service", "--restart"},
}

func upService() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		pterm.Fatal.Printfln("Failed to connect to socket")
		return
	}

	// Create portainer data volume
	volumeResponse, err := cli.VolumeList(ctx, filters.NewArgs(filters.Arg("name", "portainer_data")))

	//goland:noinspection GoNilness
	if len(volumeResponse.Volumes) == 0 {
		_, err = cli.VolumeCreate(ctx, volume.VolumeCreateBody{Name: "portainer_data"})
		if err != nil {
			pterm.Warning.Printfln("Failed to create portainer_data volume")
		}
	}

	// Check network
	if isNotNet(cli) {
		spinnerNet, _ := pterm.DefaultSpinner.Start("Network creation")
		_, err = cli.NetworkCreate(ctx, localNetworkName, types.NetworkCreate{})
		if err != nil {
			spinnerNet.Fail("Network creation error")
			return
		}
		spinnerNet.Success()
	}

	localContainers := getServicesContainer()

	for _, local := range localContainers {
		if len(source) > 0 && source != local.Name {
			continue
		}

		// Check dl-services running containers
		containerFilter := filters.NewArgs(filters.Arg("name", local.Name), filters.Arg("label", "com.docker.compose.project=dl-services"))
		isExists, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilter})
		if len(isExists) > 0 {
			if !restart {
				continue
			}

			spinnerRecreate, _ := pterm.DefaultSpinner.Start("Restarting container " + local.Name)
			err := cli.ContainerRestart(ctx, isExists[0].ID, nil)
			if err != nil {
				spinnerRecreate.Warning("Container " + local.Name + " cannot be recreated")
			}
			spinnerRecreate.Success()

			continue
		}

		// Check name running containers
		busyName := false
		containerNameFilter := filters.NewArgs(filters.Arg("name", local.Name))
		isExistsName, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerNameFilter})
		if len(isExistsName) > 0 {
			busyName = true
			pterm.Warning.Printfln("Unable to start container %s: name is already in use by container %s", isExistsName[0].ID, local.Name)
		}
		if busyName {
			continue
		}

		// Check ports
		ports := local.Ports
		busyPort := false
		for _, port := range ports {
			rawIP, hostPort, _ := splitParts(port)
			conn, _ := net.DialTimeout("tcp", net.JoinHostPort(rawIP, hostPort), time.Second)
			if conn != nil {
				//goland:noinspection GoDeferInLoop
				defer func(conn net.Conn) {
					_ = conn.Close()
				}(conn)
				busyPort = true
				pterm.Warning.Printfln("Unable to start container %s: port %s is busy.", local.Name, hostPort)
			}
		}
		if busyPort {
			continue
		}

		// Check for images
		imageFiler := filters.NewArgs(filters.Arg("reference", local.Image+":"+local.Version))
		isImageExists, err := cli.ImageList(ctx, types.ImageListOptions{All: true, Filters: imageFiler})
		if len(isImageExists) == 0 {
			spinnerPulling, _ := pterm.DefaultSpinner.WithRemoveWhenDone(true).Start("Pulling image " + local.Image)

			out, err := cli.ImagePull(ctx, local.Image+":"+local.Version, types.ImagePullOptions{})
			_, err = ioutil.ReadAll(out)
			if err != nil {
				spinnerPulling.Warning("Unable to load image: " + err.Error())
				return
			}

			spinnerPulling.Success()
		}

		spinnerStarting, _ := pterm.DefaultSpinner.Start("Starting container " + local.Name)

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

		spinnerStarting.Success()
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
