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
	setComposeFiles()
}

//setNetworkName Set network name from project name
func setDefaultEnv() {
	projectName := Env.GetString("APP_NAME")
	res := strings.ReplaceAll(projectName, ".", "")
	Env.SetDefault("NETWORK_NAME", res)

	dir, _ := os.Getwd()
	Env.SetDefault("PWD", dir)

	Env.SetDefault("REDIS", false)
	Env.SetDefault("REDIS_PASSWORD", "pass")
	Env.SetDefault("MEMCACHED", false)
}

//setComposeFile Set docker-compose files
func setComposeFiles() {
	var files []string
	confDir, _ := helper.ConfigDir()

	images := map[string]string{
		"fpm":       confDir + "/config-files/docker-compose-fpm.yaml",
		"apache":    confDir + "/config-files/docker-compose-apache.yaml",
		"redis":     confDir + "/config-files/docker-compose-redis.yaml",
		"memcached": confDir + "/config-files/docker-compose-memcached.yaml",
	}

	for imageType, imageComposeFile := range images {
		if strings.Contains(Env.GetString("PHP_VERSION"), imageType) {
			files = append(files, imageComposeFile)
		}
	}

	if Env.GetBool("REDIS") == true {
		files = append(files, images["redis"])
	}
	if Env.GetBool("MEMCACHED") == true {
		files = append(files, images["memcached"])
	}

	containers := strings.Join(files, ":")
	Env.SetDefault("COMPOSE_FILE", containers)
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
