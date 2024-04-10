package command

import (
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var certCmd = &cobra.Command{
	Use:       "cert",
	Short:     "CA certificate management",
	Long:      `Generating and (un)installing a root certificate in Firefox and/or Chrome/Chromium browsers.`,
	ValidArgs: []string{"install", "i", "uninstall", "delete"},
}

func certCommand() *cobra.Command {
	certCmd.AddCommand(
		installCertCommand(),
		uninstallCertCommand(),
	)
	return certCmd
}

func storeCertConfig(status bool) {
	viper.Set("ca", status)

	logrus.Info("Updating the configuration file")
	err := viper.WriteConfig()
	if err != nil {
		pterm.FgRed.Println(err)
	}
}
