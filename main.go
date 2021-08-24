package main

import (
	"github.com/spf13/viper"
	"github.com/varrcan/dl/cmd"
)

var colorRed = "\033[31m"
var colorGreen = "\033[32m"

func main() {
	cmd.Execute()

	viper.Debug()

	//composeProjectName := getProjectName()
	//
	//log.Println(colorGreen, composeProjectName)
}

func getProjectName() interface{} {
	composeProjectName := viper.Get("APP_NAME")

	return composeProjectName
}
