package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(recreateCmd)
}

var recreateCmd = &cobra.Command{
	Use:   "recreate",
	Short: "Recreate containers",
	Long:  `Recreate containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
		up()
	},
}
