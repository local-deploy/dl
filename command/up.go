package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up project",
	Long:  `Up project.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func up() {
	//
}
