package docker

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/go-connections/nat"
	"golang.org/x/sync/errgroup"
)

func (cli *Client) StartContainers(ctx context.Context, containers Containers, restart bool) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	for _, con := range containers {
		localContainer := con

		// Check running containers
		containerFilter := filters.NewArgs(filters.Arg("name", con.Name))
		isExists, _ := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilter})
		if len(isExists) > 0 {
			eventName := fmt.Sprintf("Container %q", localContainer.Name)
			if !restart {
				w.Event(progress.RunningEvent(eventName))
				continue
			}

			eg.Go(func() error {
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
		containerNameFilter := filters.NewArgs(filters.Arg("name", con.Name))
		isExistsName, _ := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerNameFilter})
		if len(isExistsName) > 0 {
			busyName = true
			w.Event(progress.ErrorMessageEvent(con.Name, "Unable to start container: name already in use"))
		}
		if busyName {
			continue
		}

		eventName := fmt.Sprintf("Container %q", con.Name)
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
					NetworkMode:   container.NetworkMode(localContainer.Network),
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

			if len(localContainer.AddNetwork) > 0 {
				err = cli.AddContainerToNetwork(ctx, resp.ID, localContainer.AddNetwork)
			}

			return nil
		})
	}

	return eg.Wait()
}
