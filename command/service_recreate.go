package command

import (
	"context"

	"github.com/spf13/cobra"
)

func recreateServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recreate",
		Short: "Recreate containers",
		Long:  `Stop services containers and restart.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			err := downServiceRun(ctx)
			if err != nil {
				return err
			}

			err = upServiceRun(ctx)
			if err != nil {
				return err
			}

			return nil
		},
		ValidArgs: []string{"--service"},
	}
	cmd.Flags().StringVarP(&source, "service", "s", "", "Recreate single service")
	return cmd
}
