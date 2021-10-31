package helper

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
	"net"
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

//ShowProjectInfo Display project links
func ShowProjectInfo() {
	p := ProjectEnv.GetString("APP_NAME")
	h := getLocalIp()
	pterm.FgCyan.Println()
	panels := pterm.Panels{
		{{Data: pterm.FgYellow.Sprintf("nip.io\nlocal")},
			{Data: pterm.FgYellow.Sprintf("http://%s.%s.nip.io/\nhttp://%s.localhost/", p, h, p)}},
	}

	_ = pterm.DefaultPanel.WithPanels(panels).WithPadding(5).Render()
}

func getLocalIp() string {
	//name, _ := os.Hostname()
	address, err := net.LookupHost("localhost")
	if err != nil {
		fmt.Printf("%v\n", err)
		return ""
	}

	return address[0]
}
