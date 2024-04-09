package command

import (
	"path/filepath"

	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/cert"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func uninstallCertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Removing CA certificate",
		Long:  `Removing a self-signed CA certificate from the Firefox and/or Chrome/Chromium browsers.`,
		Run: func(cmd *cobra.Command, args []string) {
			uninstallCertRun()
		},
	}
	return cmd
}

func uninstallCertRun() {
	certutilPath, err := utils.CertutilPath()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	c := &cert.Cert{
		CertutilPath:  certutilPath,
		CaFileName:    cert.CaRootName,
		CaFileKeyName: cert.CaRootKeyName,
		CaPath:        utils.CertDir(),
	}

	err = c.LoadCA()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	ca := viper.GetBool("ca")
	if !ca {
		pterm.FgYellow.Println("The local CA is not installed")
		return
	}

	c.Uninstall()

	utils.RemoveFilesInPath(filepath.Join(utils.CertDir(), "conf"))
	utils.RemoveFilesInPath(utils.CertDir())

	storeCertConfig(false)
	pterm.FgYellow.Println("The local CA is now uninstalled from the browsers trust store!")
}
