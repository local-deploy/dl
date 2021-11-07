package command

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number.`,
	Run: func(cmd *cobra.Command, args []string) {
		version := viper.GetString("version")
		fmt.Println("DL v" + version)
	},
}
