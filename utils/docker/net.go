package docker

import (
	"context"
	"fmt"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	helpers "github.com/docker/docker/integration/network"
)

func (cli *Client) IsNetworkAvailable(networkName string) bool {
	net := helpers.IsNetworkAvailable(cli, networkName)

	return net().Success()
}

func (cli *Client) IsNetworkNotAvailable(networkName string) bool {
	net := helpers.IsNetworkNotAvailable(cli, networkName)

	return net().Success()
}

func (cli *Client) CreateNetwork(ctx context.Context, networkName string) error {
	w := progress.ContextWriter(ctx)

	eventName := fmt.Sprintf("Network %q", networkName)
	w.Event(progress.CreatingEvent(eventName))
	_, err := cli.NetworkCreate(ctx, networkName, types.NetworkCreate{})
	if err != nil {
		w.Event(progress.ErrorMessageEvent("Network", fmt.Sprint(err)))
		return err
	}
	w.Event(progress.CreatedEvent(eventName))

	return nil
}

func (cli *Client) AddContainerToNetwork(ctx context.Context, containerId string, networkName string) error {
	err := cli.NetworkConnect(ctx, networkName, containerId, &network.EndpointSettings{})

	return err
}
