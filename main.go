package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"github.com/varrcan/dl/helper"
)

var version = "0.3.1"

func main() {
	if !helper.IsConfigDirExists() {
		pterm.FgRed.Printfln("The application has not been initialized. Please run the command:\ncurl -s https://raw.githubusercontent.com/local-deploy/dl/master/install_dl.sh | bash")
		return
	}

	if !wpdeployCheck() {
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

	errWrite := viper.SafeWriteConfig()

	if errWrite != nil {
		return errWrite
	}

	return errWrite
}

func wpdeployCheck() bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		pterm.Fatal.Println("Failed to connect to socket")
		return false
	}

	containerFilter := filters.NewArgs(filters.Arg("label", "com.docker.compose.project=local-services"))
	isExists, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerFilter})
	if err != nil {
		pterm.Fatal.Println(err)
		return false
	}
	if len(isExists) > 0 {
		pterm.Error.Println("An old version of wpdeploy is running. Please stop wpdeploy with the command: wpdeploy local-services down")
		return false
	}
	return true
}
