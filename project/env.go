package project

import (
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/helper"
	"os"
	"strings"
)

//Env Project variables
var Env *viper.Viper

//LoadEnv Get variables from .env file
func LoadEnv() {
	Env = viper.New()

	Env.AddConfigPath("./")
	Env.SetConfigFile(".env")
	Env.SetConfigType("env")
	err := Env.ReadInConfig()
	if err != nil {
		pterm.FgRed.Printfln(".env file not found. Please run the command: dl env")
	}

	setDefaultEnv()
	setComposeFile()
}

//setNetworkName Set network name from project name
func setDefaultEnv() {
	projectName := Env.GetString("APP_NAME")
	res := strings.ReplaceAll(projectName, ".", "")
	Env.SetDefault("NETWORK_NAME", res)

	dir, _ := os.Getwd()
	Env.SetDefault("PWD", dir)
}

//setNetworkName Set network name from project name
func setComposeFile() {
	php := Env.GetString("PHP_VERSION")
	confDir, _ := helper.ConfigDir()

	images := map[string]string{
		"fpm":    confDir + "/config-files/docker-compose-fpm.yaml",
		"apache": confDir + "/config-files/docker-compose-apache.yaml",
	}

	for imageType, imageComposeFile := range images {
		if strings.Contains(php, imageType) {
			Env.SetDefault("COMPOSE_FILE", imageComposeFile)
		}
	}
}

//CmdEnv Getting variables in the "key=value" format
func CmdEnv() []string {
	keys := Env.AllKeys()
	var env []string

	for _, key := range keys {
		value := Env.GetString(key)
		env = append(env, strings.ToUpper(key)+"="+value)
	}

	return env
}
