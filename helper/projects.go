package helper

import (
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"os"
	"strings"
)

//ProjectEnv Project variables
var ProjectEnv *viper.Viper

//LoadEnv Get variables from .env file
func LoadEnv() {
	ProjectEnv = viper.New()

	ProjectEnv.AddConfigPath("./")
	ProjectEnv.SetConfigFile(".env")
	ProjectEnv.SetConfigType("env")
	err := ProjectEnv.ReadInConfig()
	if err != nil {
		pterm.FgRed.Printfln(".env file not found. Please run the command: dl env")
	}

	setDefaultEnv()
	setComposeFile()
}

//setNetworkName Set network name from project name
func setDefaultEnv() {
	projectName := ProjectEnv.GetString("APP_NAME")
	res := strings.ReplaceAll(projectName, ".", "")
	ProjectEnv.SetDefault("NETWORK_NAME", res)

	dir, _ := os.Getwd()
	ProjectEnv.SetDefault("PWD", dir)
}

//setNetworkName Set network name from project name
func setComposeFile() {
	php := ProjectEnv.GetString("PHP_VERSION")
	confDir, _ := ConfigDir()

	images := map[string]string{
		"fpm":    confDir + "/config-files/docker-compose-fpm.yaml",
		"apache": confDir + "/config-files/docker-compose-apache.yaml",
	}

	for imageType, imageComposeFile := range images {
		if strings.Contains(php, imageType) {
			ProjectEnv.SetDefault("COMPOSE_FILE", imageComposeFile)
		}
	}
}

//CmdEnv Getting variables in the "key=value" format
func CmdEnv() []string {
	keys := ProjectEnv.AllKeys()
	var env []string

	for _, key := range keys {
		value := ProjectEnv.GetString(key)
		env = append(env, strings.ToUpper(key)+"="+value)
	}

	return env
}
