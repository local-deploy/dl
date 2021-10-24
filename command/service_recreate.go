package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(recreateServiceCmd)
	recreateServiceCmd.Flags().StringVarP(&source, "service", "s", "", "Recreate single service")
}

var recreateServiceCmd = &cobra.Command{
	Use:   "recreate",
	Short: "Recreate containers",
	Long:  `Recreate containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		downService()
		upService()
	},
}
