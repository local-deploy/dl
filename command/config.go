package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configLangCmd)
}

var configLangCmd = &cobra.Command{
	Use:   "config",
	Short: "Application configuration",
	Long:  `Menu for setting up the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DL " + version)
	},
}
