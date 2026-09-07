package main

import (
	"embed"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/local-deploy/dl/command"
	"github.com/local-deploy/dl/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

var version = "dev"

//go:embed templates/*
var templates embed.FS

func main() {
	pterm.ThemeDefault.SecondaryStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}

	// forwarding file variable to package
	utils.Templates = templates

	if !dockerCheck() {
		return
	}

	if utils.IsNeedInstall() {
		firstStart()
	}

	if utils.IsNeedReInstall() {
		reinstallTemplates()
	}

	if !utils.IsCertPathExists() {
		createCertDirectory()
	}

	initConfig()
	command.Execute()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configDir := utils.ConfigDir()

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		pterm.FgRed.Printfln("Error config file: %s \n", err)
		os.Exit(1)
	}

	if viper.GetString("version") != version {
		viper.Set("version", version)
		err = viper.WriteConfig()
		if err != nil {
			pterm.FgRed.Printfln("Error config file: %s \n", err)
			os.Exit(1)
		}
	}

	updateTemplates()

	viper.AutomaticEnv()
}

// templatesVersion the version of the files in templates/. Bump it whenever a template
// changes: existing installations keep their unpacked copies until the number differs.
// It subsumes the "traefik != 3" migration that used to live in initConfig.
const templatesVersion = "2"

// updateTemplates unpack the templates again when the installation carries an older version
func updateTemplates() {
	if viper.GetString("templates") == templatesVersion {
		return
	}

	// overwrite: templates removed in the new version must not linger in the directory
	if err := utils.CreateTemplates(true); err != nil {
		pterm.FgRed.Printfln("Unable to create template files: %s \n", err)
		os.Exit(1)
	}

	viper.Set("templates", templatesVersion)
	if err := viper.WriteConfig(); err != nil {
		// the templates are already unpacked, so a failed write only means unpacking them
		// again on the next run — not a reason to refuse to run the command
		pterm.FgYellow.Printfln("Unable to save the template version: %s", err)
	}
}

func firstStart() {
	err := createConfigFile()
	if err != nil {
		pterm.FgRed.Printfln("Unable to create config file: %s \n", err)
		os.Exit(1)
	}

	err = utils.CreateTemplates(true)
	if err != nil {
		pterm.FgRed.Printfln("Unable to create template files: %s \n", err)
		os.Exit(1)
	}

	// remove old template directory
	err = utils.RemovePath(filepath.Join(utils.ConfigDir(), "config-files"))
	if err != nil {
		pterm.FgRed.Printfln("Unable to remove old template files: %s \n", err)
	}
}

func reinstallTemplates() {
	err := utils.CreateTemplates(false)
	if err != nil {
		pterm.FgRed.Printfln("Unable to create template files: %s \n", err)
		os.Exit(1)
	}
}

func createConfigFile() error {
	configDir := utils.ConfigDir()

	err := utils.CreateDirectory(configDir)
	if err != nil {
		return err
	}

	// do not overwrite the config if it exists
	if utils.PathExists(filepath.Join(configDir, "config.yaml")) {
		return nil
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	viper.Set("version", version)
	viper.Set("locale", "en")
	viper.Set("repo", "ghcr.io")
	viper.Set("templates", templatesVersion)
	viper.Set("check-updates", time.Now())

	errWrite := viper.SafeWriteConfig()

	if errWrite != nil {
		return errWrite
	}

	return errWrite
}

func createCertDirectory() {
	err := utils.CreateDirectory(filepath.Join(utils.CertDir(), "conf"))
	if err != nil {
		pterm.FgRed.Printfln("Unable to create certs directory: %s \n", err)
		os.Exit(1)
	}
}

func dockerCheck() bool {
	_, err := exec.LookPath("docker")
	if err != nil {
		pterm.FgRed.Printfln("Docker not found. Please install it. https://docs.docker.com/engine/install/")
		return false
	}
	return true
}
