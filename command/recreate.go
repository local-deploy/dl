package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(recreateCmd)
	recreateCmd.Flags().StringVarP(&projectContainer, "container", "c", "", "Recreate single container")
}

var recreateCmd = &cobra.Command{
	Use:   "recreate",
	Short: "Recreate containers",
	Long:  `Recreate containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
		up()
	},
	Example: "dl recreate\ndl recreate -c db",
}
