package command

import (
	"context"

	"github.com/docker/compose/v2/pkg/progress"
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(restartServiceCmd)
}

var restartServiceCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart containers",
	Long:  `Restarts running service containers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		err := progress.Run(ctx, func(ctx context.Context) error {
			restart = true
			return upService(ctx)
		})
		if err != nil {
			return err
		}

		return nil
	},
}
