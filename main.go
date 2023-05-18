package main

import (
	"embed"
	"os"
	"os/exec"
	"time"

	"github.com/local-deploy/dl/command"
	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

var version = "0.5.7"

//go:embed config-files/*
var templates embed.FS

func main() {
	pterm.ThemeDefault.SecondaryStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}

	// forwarding file variable to package
	utils.Templates = templates

	if !dockerCheck() {
		return
	}

	if !helper.IsConfigFileExists() {
		firstStart()
	}

	initConfig()
	command.Execute()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configDir := helper.ConfigDir()

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

	if !helper.IsAptInstall() {
		err = utils.CreateTemplates()
		if err != nil {
			pterm.FgRed.Printfln("Unable to create template files: %s \n", err)
			os.Exit(1)
		}
	}
}

func createConfigFile() error {
	configDir := helper.ConfigDir()

	err := helper.CreateDirectory(configDir)
	if err != nil {
		return err
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

func dockerCheck() bool {
	_, err := exec.LookPath("docker")
	if err != nil {
		pterm.FgRed.Printfln("Docker not found. Please install it. https://docs.docker.com/engine/install/")
		return false
	}
	return true
}
