package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(restartServiceCmd)
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart containers",
	Long:  `Restarts running service containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		restart = true
		upService()
	},
}
