package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Local services configuration",
	Long:  `Local services configuration (portainer, mailcatcher, nginx).`,
}
