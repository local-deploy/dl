package main

import (
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"github.com/varrcan/dl/helper"
	"os"
)

func installApp() {
	if helper.IsConfigDirExists() {
		pterm.FgCyan.Printfln("DL is already installed")

		return
	}

	err := createConfigDir()

	if err != nil {
		pterm.FgRed.Printfln("Unable to create settings directory: %w \n", err)
	}

	err = createConfigFile()

	if err != nil {
		pterm.FgRed.Printfln("Unable to create config file: %w \n", err)
	}

	pterm.FgGreen.Printfln("DL installed successfully")
}

func createConfigDir() error {
	if !helper.IsConfigDirExists() {
		configDir, _ := helper.ConfigDir()
		errMakeDir := os.Mkdir(configDir, 0755)

		if errMakeDir != nil {
			return errMakeDir
		}
	}

	return nil
}

func createConfigFile() error {
	configDir, _ := helper.ConfigDir()

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	viper.Set("version", command.Version)
	viper.Set("locale", "en")

	errWrite := viper.SafeWriteConfig()

	if errWrite != nil {
		return errWrite
	}

	return errWrite
}
