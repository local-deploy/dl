package project

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"github.com/varrcan/dl/helper"
)

// Env Project variables
var Env *viper.Viper

var phpImagesVersion = map[string]string{
	"7.3-apache": "1.1.0",
	"7.3-fpm":    "1.0.0",
	"7.4-apache": "1.0.1",
	"7.4-fpm":    "1.0.0",
	"8.0-apache": "1.0.2",
	"8.0-fpm":    "1.0.1",
}

// LoadEnv Get variables from .env file
func LoadEnv() {
	Env = viper.New()

	Env.AddConfigPath("./")
	Env.SetConfigFile(".env")
	Env.SetConfigType("env")
	err := Env.ReadInConfig()
	if err != nil {
		pterm.FgRed.Printfln(".env file not found. Please run the command: dl env")
		os.Exit(1)
	}

	setDefaultEnv()
	setComposeFiles()
}

// setNetworkName Set network name from project name
func setDefaultEnv() {
	projectName := strings.ToLower(Env.GetString("HOST_NAME"))

	if len(projectName) == 0 {
		pterm.FgRed.Printfln("The HOST_NAME variable is not defined! Please initialize .env file.")
		os.Exit(1)
	}

	var re = regexp.MustCompile(`[[:punct:]]`)
	res := re.ReplaceAllString(projectName, "")
	Env.SetDefault("NETWORK_NAME", res)

	dir, _ := os.Getwd()
	Env.SetDefault("PWD", dir)
	Env.SetDefault("SSH_KEY", "id_rsa")

	confDir, _ := helper.ConfigDir()
	configNginxFile := filepath.Join(confDir, "config-files", "default.conf.template")
	Env.SetDefault("NGINX_CONF", configNginxFile)

	Env.SetDefault("REDIS", false)
	Env.SetDefault("REDIS_PASSWORD", "pass")
	Env.SetDefault("MEMCACHED", false)

	host := getLocalIp()

	Env.SetDefault("LOCAL_IP", host)
	Env.SetDefault("NIP_DOMAIN", fmt.Sprintf("%s.%s.nip.io", projectName, host))
	Env.SetDefault("LOCAL_DOMAIN", fmt.Sprintf("%s.localhost", projectName))
}

// setComposeFile Set docker-compose files
func setComposeFiles() {
	var files []string
	confDir, _ := helper.ConfigDir()
	phpVersion := Env.GetString("PHP_VERSION")

	if len(phpVersion) == 0 {
		pterm.FgRed.Printfln("The PHP_VERSION variable is not defined! Please initialize .env file.")
		os.Exit(1)
	}

	Env.SetDefault("PHP_IMAGE_VERSION", phpImagesVersion[phpVersion])

	images := map[string]string{
		"mysql":     confDir + "/config-files/docker-compose-mysql.yaml",
		"fpm":       confDir + "/config-files/docker-compose-fpm.yaml",
		"apache":    confDir + "/config-files/docker-compose-apache.yaml",
		"redis":     confDir + "/config-files/docker-compose-redis.yaml",
		"memcached": confDir + "/config-files/docker-compose-memcached.yaml",
	}

	for imageType, imageComposeFile := range images {
		if strings.Contains(phpVersion, imageType) {
			files = append(files, imageComposeFile)
		}
	}

	if Env.GetFloat64("MYSQL_VERSION") > 0 {
		files = append(files, images["mysql"])
	}
	if Env.GetBool("REDIS") {
		files = append(files, images["redis"])
	}
	if Env.GetBool("MEMCACHED") {
		files = append(files, images["memcached"])
	}

	containers := strings.Join(files, ":")
	Env.SetDefault("COMPOSE_FILE", containers)
}

// CmdEnv Getting variables in the "key=value" format
func CmdEnv() []string {
	keys := Env.AllKeys()
	var env []string

	for _, key := range keys {
		value := Env.GetString(key)
		env = append(env, strings.ToUpper(key)+"="+value)
	}

	return env
}

// IsEnvFileExists checking for the existence of .env file
func IsEnvFileExists() bool {
	dir, _ := os.Getwd()
	env := filepath.Join(dir, ".env")

	_, err := os.Stat(env)

	if err != nil {
		return false
	}

	return true
}

// IsEnvExampleFileExists checking for the existence of .env.example file
func IsEnvExampleFileExists() bool {
	dir, _ := os.Getwd()
	env := filepath.Join(dir, ".env.example")

	_, err := os.Stat(env)

	if err != nil {
		return false
	}

	return true
}

func getLocalIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
