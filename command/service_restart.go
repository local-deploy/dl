package command

import (
	"context"

	"github.com/spf13/cobra"
)

func restartServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart containers",
		Long:  `Restarts running service containers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			recreate = true
			err := upServiceRun(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
