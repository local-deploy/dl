package command

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "v.0.0.1"

func init() {
	//rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DL " + version)
	},
}
