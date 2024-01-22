package docker

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	helpers "github.com/docker/docker/integration/network"
	"golang.org/x/sync/errgroup"
)

// IsNetworkAvailable checking that the network exists
// Deprecated
func (cli *Client) IsNetworkAvailable(networkName string) bool {
	net := helpers.IsNetworkAvailable(cli.DockerCli.Client(), networkName)

	return net().Success()
}

// IsNetworkNotAvailable checking that the network does not exist
// Deprecated
func (cli *Client) IsNetworkNotAvailable(networkName string) bool {
	net := helpers.IsNetworkNotAvailable(cli.DockerCli.Client(), networkName)

	return net().Success()
}

// CreateNetwork create a new network
// Deprecated
func (cli *Client) CreateNetwork(ctx context.Context, networkName string) error {
	w := progress.ContextWriter(ctx)

	eventName := fmt.Sprintf("Network %q", networkName)
	w.Event(progress.CreatingEvent(eventName))
	_, err := cli.DockerCli.Client().NetworkCreate(ctx, networkName, types.NetworkCreate{})
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Network", fmt.Sprint(err)))
		return err
	}
	w.Event(progress.CreatedEvent(eventName))

	return nil
}

// RemoveNetwork delete network
// Deprecated
func (cli *Client) RemoveNetwork(ctx context.Context, networkName string) error {
	w := progress.ContextWriter(ctx)
	eg, _ := errgroup.WithContext(ctx)

	eg.Go(func() error {
		eventName := fmt.Sprintf("Network %q", networkName)
		w.Event(progress.RemovingEvent(eventName))

		netFilters := filters.NewArgs(filters.Arg("name", networkName))
		list, _ := cli.DockerCli.Client().NetworkList(ctx, types.NetworkListOptions{Filters: netFilters})
		err := cli.DockerCli.Client().NetworkRemove(ctx, list[0].ID)

		if err != nil {
			w.Event(progress.ErrorMessageEvent(eventName, fmt.Sprint(err)))
			return nil
		}

		w.Event(progress.RemovedEvent(eventName))
		return nil
	})

	return eg.Wait()
}

// addContainerToNetwork add a container to the network
// Deprecated
func (cli *Client) addContainerToNetwork(ctx context.Context, containerID string, networkName string) error {
	err := cli.DockerCli.Client().NetworkConnect(ctx, networkName, containerID, &network.EndpointSettings{})

	return err
}
