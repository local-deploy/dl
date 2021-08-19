package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/varrcan/dl/cmd"
	"log"
	"os"
)

var colorRed = "\033[31m"
var colorGreen = "\033[32m"

func init() {
	// загрузить переменные из .env
	if err := godotenv.Load(); err != nil {
		//log.Println(colorRed, "No .env file found")
		//os.Exit(1)
	}
}

func main() {
	cmd.Execute()

	composeProjectName := getProjectName()

	log.Println(colorGreen, composeProjectName)
}

func getProjectName() string {
	composeProjectName, exists := os.LookupEnv("COMPOSE_PROJECT_NAME")

	if !exists {
		fmt.Println(colorRed, "Key COMPOSE_PROJECT_NAME not found in .env file")
		os.Exit(1)
	}

	return composeProjectName
}
