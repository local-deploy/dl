package command

import (
	"github.com/local-deploy/dl/utils"
	"github.com/spf13/cobra"
)

func templateCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Regenerate docker-compose files",
		Long: `Restoring the original docker-compose files in the configuration directory.
The command will not work on Linux systems when installed via apt-manager.`,
		Example: "dl templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runTemplateCreate()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}

func runTemplateCreate() error {
	err := utils.CreateTemplates(true)
	if err != nil {
		return err
	}

	return nil
}
