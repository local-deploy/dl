package command

import (
	"context"
	"os"
	"path/filepath"

	"github.com/local-deploy/dl/utils"
	"github.com/local-deploy/dl/utils/cert"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var reinstallCert bool

func installCertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "Installing CA certificate",
		Long:    `Generating a self-signed CA certificate and installing it in Firefox and/or Chrome/Chromium browsers.`,
		Run: func(cmd *cobra.Command, args []string) {
			installCertRun()
		},
		ValidArgs: []string{"--reinstall", "-r"},
	}
	cmd.Flags().BoolVarP(&reinstallCert, "reinstall", "r", false, "Reinstall certificate")
	return cmd
}

func installCertRun() {
	certutilPath, err := utils.CertutilPath()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	err = utils.CreateDirectory(filepath.Join(utils.CertDir(), "conf"))
	if err != nil {
		pterm.FgRed.Printfln("Error: %s \n", err)
		os.Exit(1)
	}

	c := &cert.Cert{
		CertutilPath:  certutilPath,
		CaFileName:    cert.CaRootName,
		CaFileKeyName: cert.CaRootKeyName,
		CaPath:        utils.CertDir(),
	}

	if reinstallCert {
		err = c.LoadCA()
		if err != nil {
			pterm.FgRed.Printfln("Error: %s", err)
			return
		}
		c.Uninstall()
		utils.RemoveFilesInPath(filepath.Join(utils.CertDir(), "conf"))
		utils.RemoveFilesInPath(utils.CertDir())
	}

	_, err = os.Stat(filepath.Join(utils.CertDir(), cert.CaRootName))
	if err != nil {
		err := c.CreateCA()
		if err != nil {
			pterm.FgRed.Printfln("Error: %s", err)
			return
		}
	}

	err = c.LoadCA()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	if c.Check() {
		pterm.FgGreen.Println("The local CA is already installed in the browsers trust store!")
	} else if c.Install() {
		storeCertConfig(true)
		pterm.FgGreen.Println("The local CA is now installed in the browsers trust store (requires browser restart)!")

		// Recreate local services
		recreate = true
		ctx := context.Background()
		err = upServiceRun(ctx)
		if err != nil {
			pterm.FgYellow.Println("Restart services for changes to take effect: dl service recreate")
		}
	}
}
