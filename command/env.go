package command

import (
	"bufio"
	"github.com/dixonwille/wmenu/v5"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"github.com/varrcan/dl/project"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(envCmd)
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Create env file",
	Long:  `Create or replace an .env file. If the .env.example file is located in the root project directory, it will be used.`,
	Run: func(cmd *cobra.Command, args []string) {
		env()
	},
}

func env() {
	if project.IsEnvFileExists() {
		showEnvMenu()
	} else {
		if copyEnv() {
			pterm.FgGreen.Println("The .env file has been created successfully. Please specify the necessary variables.")
		}
	}
}

func showEnvMenu() {
	pterm.FgYellow.Println("The .env file exists!")
	menu := wmenu.NewMenu("Select the necessary action:")
	menu.LoopOnInvalid()

	menu.Option("Replace file", "replace", false, func(opt wmenu.Opt) error {
		deleteEnv()
		copyEnv()
		pterm.FgGreen.Println("File replaced successfully.")
		return nil
	})
	//menu.Option("Merge (dangerous)", "merge", false, func(opt wmenu.Opt) error {
	//	mergeEnv()
	//	return nil
	//})
	menu.Option("Just show", "show", false, func(opt wmenu.Opt) error {
		printEnvConfig()
		return nil
	})
	menu.Option("Abort", "abort", true, func(opt wmenu.Opt) error {
		return nil
	})

	err := menu.Run()
	if err != nil {
		pterm.FgRed.Println(err)

		return
	}
}

func printEnvConfig() {
	configDir, _ := helper.ConfigDir()
	src := filepath.Join(configDir, "/config-files/.env.example")

	file, err := os.Open(src)
	if err != nil {
		pterm.FgRed.Println(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			pterm.FgRed.Println(err)
		}
	}()

	scanner := bufio.NewScanner(file)

	pterm.Println()
	pterm.FgGreen.Println("Copy the variables to your .env file and adjust the values")
	pterm.Println()
	for scanner.Scan() {
		pterm.FgCyan.Println(scanner.Text())
	}
}

func deleteEnv() {
	err := os.Remove(".env")

	if err != nil {
		pterm.FgRed.Println(err)
		return
	}
}

func copyEnv() bool {
	var src string

	currentDir, _ := os.Getwd()
	configDir, _ := helper.ConfigDir()

	if project.IsEnvExampleFileExists() {
		src = filepath.Join(currentDir, ".env.example")
	} else {
		src = filepath.Join(configDir, "/config-files/.env.example")
	}

	dest := filepath.Join(currentDir, ".env")

	bytesRead, err := ioutil.ReadFile(src)

	if err != nil {
		pterm.FgRed.Println(err)

		return false
	}

	err = ioutil.WriteFile(dest, bytesRead, 0644)

	if err != nil {
		pterm.FgRed.Println(err)

		return false
	}

	return true
}
