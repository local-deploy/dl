package command

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:    "config",
	Short:  "Application configuration",
	Long:   `Menu for setting up the application.`,
	Hidden: false,
}

func configCommand() *cobra.Command {
	configCmd.AddCommand(
		configLangCommand(),
		configRepoCommand(),
	)
	return configCmd
}
