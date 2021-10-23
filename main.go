package main

import (
	"github.com/pterm/pterm"
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

	initConfig()
	command.Execute()

	//viper.Debug()

	//composeProjectName := getProjectName()
	//
	//log.Println(composeProjectName)
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

	//env := viper.New()
	//
	//env.AddConfigPath("./")
	//env.SetConfigFile(".env")
	//env.SetConfigType("env")
	//env.ReadInConfig()
	//env.Debug()

	viper.AutomaticEnv()
}

func getProjectName() interface{} {
	composeProjectName := viper.Get("APP_NAME")

	return composeProjectName
}
