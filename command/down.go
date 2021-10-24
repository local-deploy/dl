package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Down project",
	Long:  `Down project.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func down() {
	//
}
