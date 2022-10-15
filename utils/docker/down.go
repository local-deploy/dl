package docker

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/sync/errgroup"
)

// RemoveContainers stop and remove docker containers
func (cli *Client) RemoveContainers(ctx context.Context, containers Containers) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	containerFilters := filters.NewArgs()
	for _, container := range containers {
		// if len(source) > 0 && source != container.Name {
		// 	continue
		// }
		containerFilters.Add("name", container.Name)
	}

	if containerFilters.Len() == 0 {
		fmt.Println("Unknown service")
		return nil
	}

	removeContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilters})
	if err != nil {
		return err
	}

	for _, container := range removeContainers {
		container := container

		eg.Go(func() error {
			eventName := getContainerProgressName(container)

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
