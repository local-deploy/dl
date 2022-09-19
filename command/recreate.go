package command

import (
	"github.com/spf13/cobra"
)

func recreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recreate",
		Aliases: []string{"restart"},
		Short:   "Recreate containers",
		Long:    `Stop project containers and restart. Alias for sequential execution of "dl down && dl up" commands.`,
		Run: func(cmd *cobra.Command, args []string) {
			downRun()
			upRun()
		},
	}
	return cmd
}
