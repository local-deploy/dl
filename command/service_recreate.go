package command

import (
	"context"

	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(recreateServiceCmd)
	recreateServiceCmd.Flags().StringVarP(&source, "service", "s", "", "Recreate single service")
}

var recreateServiceCmd = &cobra.Command{
	Use:   "recreate",
	Short: "Recreate containers",
	Long:  `Stop services containers and restart.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		err := downService(ctx)
		if err != nil {
			return err
		}

		err = upService(ctx)
		if err != nil {
			return err
		}

		return nil
	},
	ValidArgs: []string{"--service"},
}
