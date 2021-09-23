package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(restartCmd)
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart containers",
	Long:  `Restart containers.`,
}
