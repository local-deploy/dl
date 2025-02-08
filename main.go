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

	viper.AutomaticEnv()
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
