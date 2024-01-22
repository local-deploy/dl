package command

import (
	"context"

	"github.com/spf13/cobra"
)

func recreateServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "recreate",
		Aliases: []string{"restart"},
		Short:   "Recreate containers",
		Long:    `Recreate running service containers.`,
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
