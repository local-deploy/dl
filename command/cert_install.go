package command

import (
	"context"
	"os"
	"path/filepath"

	"github.com/local-deploy/dl/helper"
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
	certutilPath, err := helper.CertutilPath()
	if err != nil {
		pterm.FgRed.Printfln("Error: %s", err)
		return
	}

	err = helper.CreateDirectory(filepath.Join(helper.CertDir(), "conf"))
	if err != nil {
		pterm.FgRed.Printfln("Error: %s \n", err)
		os.Exit(1)
	}

	c := &cert.Cert{
		CertutilPath:  certutilPath,
		CaFileName:    cert.CaRootName,
		CaFileKeyName: cert.CaRootKeyName,
		CaPath:        helper.CertDir(),
	}

	if reinstallCert {
		err = c.LoadCA()
		if err != nil {
			pterm.FgRed.Printfln("Error: %s", err)
			return
		}
		c.Uninstall()
		helper.RemoveFilesInPath(filepath.Join(helper.CertDir(), "conf"))
		helper.RemoveFilesInPath(helper.CertDir())
	}

	_, err = os.Stat(filepath.Join(helper.CertDir(), cert.CaRootName))
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

		// Restart traefik
		// TODO: Will not work with a new client!
		source = "traefik"
		ctx := context.Background()
		_ = downServiceRun(ctx)
		err = upServiceRun(ctx)
		if err != nil {
			pterm.FgYellow.Println("Restart services for changes to take effect: dl service recreate")
		}
	}
}
