package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/command"
	"log"
	"os"
	"path/filepath"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	initConfig()
}

func main() {
	command.Execute()

	//viper.Debug()

	//composeProjectName := getProjectName()
	//
	//log.Println(composeProjectName)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, err := os.UserHomeDir()
	configDir := filepath.Join(home, ".dl")
	cobra.CheckErr(err)

	viper.SetDefault("locale", "en")

	viper.AddConfigPath(configDir)
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			errMakeDir := os.Mkdir(configDir, 0755)
			handleError(errMakeDir)
			viper.Set("version", command.Version)

			errWrite := viper.SafeWriteConfig()
			handleError(errWrite)
		} else {
			panic(fmt.Errorf("Fatal error config file: %w \n", err))
		}
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
