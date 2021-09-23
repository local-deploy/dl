package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop and remove services",
	Long:  `Stop and remove services.`,
}
