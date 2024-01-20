package command

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/local-deploy/dl/containers"
	"github.com/spf13/cobra"
)

var source string

var serviceCmd = &cobra.Command{
	Use:       "service",
	Short:     "Local services configuration",
	Long:      `Local services configuration (portainer, mailcatcher, traefik).`,
	ValidArgs: []string{"up", "down", "recreate", "restart"},
}

func serviceCommand() *cobra.Command {
	serviceCmd.AddCommand(
		downServiceCommand(),
		recreateServiceCommand(),
		upServiceCommand(),
	)
	return serviceCmd
}

// getServicesContainer local services containers
func getServicesContainer() []types.ServiceConfig {
	configs := []types.ServiceConfig{
		containers.Traefik(),
		containers.Mail(),
		containers.Portainer(),
	}

	return configs
}
