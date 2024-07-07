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
		RunE: func(_ *cobra.Command, _ []string) error {
			ctx := context.Background()
			recreate = true
			err := upServiceRun(ctx)
			if err != nil {
				return err
			}

			return nil
		},
		ValidArgs: []string{"--services"},
	}
	return cmd
}
