package main

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"github.com/varrcan/dl/helper"
	"log"
	"os"
)

func handleError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	args := os.Args
	if len(args[1:]) > 0 && args[1] == "install" {
		installApp()

		return
	}

	isConfig := helper.IsConfigDirExists()
	if !isConfig {
		pterm.FgRed.Printfln("The application has not been initialized. Please run the command: dl install")

		return
	}

	cobra.OnInitialize(initConfig)
	command.Execute()

	//viper.Debug()
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
