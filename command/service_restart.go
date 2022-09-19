package command

import (
	"context"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/spf13/cobra"
)

func restartServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart containers",
		Long:  `Restarts running service containers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			err := progress.Run(ctx, func(ctx context.Context) error {
				restart = true
				return upServiceRun(ctx)
			})
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
