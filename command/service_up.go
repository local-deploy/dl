package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		err := progress.Run(ctx, func(ctx context.Context) error {
			return upService(ctx)
		})
		if err != nil {
			return err
		}

		return nil
	},
	ValidArgs: []string{"--service", "--restart"},
}

func upService(ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Docker", "Failed connect to socket"))
		return err
	}

	// Check for images
	err = pullRequiredImages(cli, ctx)
	if err != nil {
		return err
	}

	// Check network
	if isNotNet(cli) {
		err := createNetwork(cli, ctx)
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
		_, err = cli.VolumeCreate(ctx, volume.VolumeCreateBody{Name: "portainer_data"})
		if err != nil {
			w.Event(progress.ErrorMessageEvent("Volume", fmt.Sprint(err)))
			return nil
		}
		w.Event(progress.CreatedEvent(eventName))
	}

	err = startContainers(cli, ctx)
	if err != nil {
		return err
	}

	return err
}

func pullRequiredImages(cli *client.Client, ctx context.Context) error {

	return progress.Run(ctx, func(ctx context.Context) error {
		w := progress.ContextWriter(ctx)
		eg, ctx := errgroup.WithContext(ctx)

		localContainers := getServicesContainer()
		for _, local := range localContainers {
			localContainer := local
			imageFiler := filters.NewArgs(filters.Arg("reference", localContainer.Image+":"+localContainer.Version))
			isImageExists, _ := cli.ImageList(ctx, types.ImageListOptions{All: true, Filters: imageFiler})

			if len(isImageExists) == 0 {
				eg.Go(func() error {
					w.Event(progress.Event{
						ID:     localContainer.Name,
						Status: progress.Working,
						Text:   "Pulling",
					})

					stream, err := cli.ImagePull(ctx, localContainer.Image+":"+localContainer.Version, types.ImagePullOptions{})
					if err != nil {
						w.TailMsgf(fmt.Sprint(err))
						w.Event(progress.ErrorEvent(localContainer.Name))
						return nil
					}

					dec := json.NewDecoder(stream)
					for {
						var jm jsonmessage.JSONMessage
						if err := dec.Decode(&jm); err != nil {
							if err == io.EOF {
								break
							}
							return err
						}
						if jm.Error != nil {
							return err
						}
						toPullProgressEvent(localContainer.Name, jm, w)
					}

					w.Event(progress.Event{
						ID:     localContainer.Name,
						Status: progress.Done,
						Text:   "Pulled",
					})

					return err
				})
			}
		}

		err := eg.Wait()
		if err != nil {
			return err
		}
		return err
	})
}

func toPullProgressEvent(parent string, jm jsonmessage.JSONMessage, w progress.Writer) {
	if jm.ID == "" || jm.Progress == nil {
		return
	}

	var (
		text   string
		status = progress.Working
	)

	text = jm.Progress.String()

	if jm.Status == "Pull complete" ||
		jm.Status == "Already exists" ||
		strings.Contains(jm.Status, "Image is up to date") ||
		strings.Contains(jm.Status, "Downloaded newer image") {
		status = progress.Done
	}

	if jm.Error != nil {
		status = progress.Error
		text = jm.Error.Message
	}

	w.Event(progress.Event{
		ID:         jm.ID,
		ParentID:   parent,
		Text:       jm.Status,
		Status:     status,
		StatusText: text,
	})
}

func createNetwork(cli *client.Client, ctx context.Context) error {
	w := progress.ContextWriter(ctx)

	eventName := fmt.Sprintf("Network %q", localNetworkName)
	w.Event(progress.CreatingEvent(eventName))
	_, err := cli.NetworkCreate(ctx, localNetworkName, types.NetworkCreate{})
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Network", fmt.Sprint(err)))
		return err
	}
	w.Event(progress.CreatedEvent(eventName))

	return nil
}

func startContainers(cli *client.Client, ctx context.Context) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	localContainers := getServicesContainer()

	for _, local := range localContainers {
		localContainer := local
		if len(source) > 0 && source != local.Name {
			continue
		}

		// Check dl-services running containers
		containerFilter := filters.NewArgs(filters.Arg("name", local.Name), filters.Arg("label", "com.docker.compose.project=dl-services"))
		isExists, _ := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilter})
		if len(isExists) > 0 {
			if !restart {
				continue
			}

			eg.Go(func() error {
				eventName := fmt.Sprintf("Container %q", localContainer.Name)
				w.Event(progress.RestartingEvent(eventName))
				err := cli.ContainerRestart(ctx, isExists[0].ID, nil)
				if err != nil {
					w.TailMsgf(fmt.Sprint(err))
					w.Event(progress.ErrorEvent(eventName))
					return nil
				}

				w.Event(progress.RestartedEvent(eventName))
				return nil
			})

			continue
		}

		// Check name running containers
		busyName := false
		containerNameFilter := filters.NewArgs(filters.Arg("name", local.Name))
		isExistsName, _ := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerNameFilter})
		if len(isExistsName) > 0 {
			busyName = true
			w.Event(progress.ErrorMessageEvent(local.Name, "Unable to start container: name already in use"))
		}
		if busyName {
			continue
		}

		eventName := fmt.Sprintf("Container %q", local.Name)
		w.Event(progress.CreatingEvent(eventName))

		// Create containers
		eg.Go(func() error {
			exposedPorts, portBindings, _ := nat.ParsePortSpecs(localContainer.Ports)

			resp, err := cli.ContainerCreate(ctx,
				&container.Config{
					Cmd:          localContainer.Cmd,
					Image:        localContainer.Image,
					Volumes:      localContainer.Volumes,
					Entrypoint:   localContainer.Entrypoint,
					Labels:       localContainer.Labels,
					ExposedPorts: exposedPorts,
					Env:          localContainer.Env,
				},
				&container.HostConfig{
					NetworkMode:   container.NetworkMode(localNetworkName),
					RestartPolicy: container.RestartPolicy{Name: "always"},
					PortBindings:  portBindings,
					Mounts:        localContainer.Mounts,
				}, nil, nil, localContainer.Name)

			if err != nil {
				w.TailMsgf(fmt.Sprint(err))
				w.Event(progress.ErrorEvent(eventName))
				return nil
			}

			// Start containers
			w.Event(progress.StartingEvent(eventName))
			err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
			if err != nil {
				w.TailMsgf(fmt.Sprint(err))
				w.Event(progress.ErrorEvent(eventName))
				return nil
			}

			w.Event(progress.StartedEvent(eventName))

			return nil
		})
	}

	return eg.Wait()
}
