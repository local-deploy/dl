package command

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:    "config",
	Short:  "Application configuration",
	Long:   `Menu for setting up the application.`,
	Hidden: false,
}
