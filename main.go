package main

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"github.com/varrcan/dl/helper"
	"log"
)

var version = "0.1.2"

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	if !helper.IsConfigDirExists() {
		pterm.FgRed.Printfln("The application has not been initialized. Please run the command:\nwget --no-check-certificate https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh && chmod +x ./install_dl.sh && ./install_dl.sh")

		return
	}

	if !helper.IsConfigFileExists() {
		firstStart()
	}

	cobra.OnInitialize(initConfig)
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
		pterm.FgRed.Printfln("Error config file: %w \n", err)
	}

	viper.AutomaticEnv()
}

func firstStart() {
	err := createConfigFile()

	if err != nil {
		pterm.FgRed.Printfln("Unable to create config file: %w \n", err)
	}
}

func createConfigFile() error {
	configDir, _ := helper.ConfigDir()

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	viper.Set("version", version)
	viper.Set("locale", "en")

	errWrite := viper.SafeWriteConfig()

	if errWrite != nil {
		return errWrite
	}

	return errWrite
}
