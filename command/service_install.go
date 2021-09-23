package command

import (
	"github.com/spf13/cobra"
)

func init() {
	serviceCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Add services to systemd startup",
	Long:  `Add services to systemd startup. Portainer, mailcatcher and nginx will be launched at system startup.`,
}
