package main

import (
	"os/exec"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"github.com/varrcan/dl/helper"
)

var version = "0.4.5"

func main() {
	pterm.ThemeDefault.SecondaryStyle = pterm.Style{pterm.FgDefault, pterm.BgDefault}

	if !helper.IsConfigDirExists() {
		pterm.FgRed.Printfln("The application has not been initialized. Please run the command:\ncurl -s https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh | bash")
		return
	}

	if !dockerCheck() {
		return
	}

	if !helper.WpdeployCheck() {
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
	configDir, _ := helper.ConfigDir()

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		pterm.FgRed.Printfln("Error config file: %s \n", err)
	}

	viper.AutomaticEnv()
}

func firstStart() {
	err := createConfigFile()

	if err != nil {
		pterm.FgRed.Printfln("Unable to create config file: %s \n", err)
	}
}

func createConfigFile() error {
	configDir, _ := helper.ConfigDir()

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
