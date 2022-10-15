package command

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"github.com/varrcan/dl/project"
)

func envCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Create env file",
		Long:  `Create or replace an .env file. If the .env.example file is located in the root project directory, it will be used.`,
		Run: func(cmd *cobra.Command, args []string) {
			runEnv()
		},
	}
	return cmd
}

func runEnv() {
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

	options := []string{"replace file", "just show", "abort"}
	selectedOption, _ := pterm.DefaultInteractiveSelect.
		WithOptions(options).
		Show("Select the necessary action")

	switch selectedOption {
	case "replace file":
		deleteEnv()
		copyEnv()
		pterm.FgGreen.Println("File replaced successfully.")
		break
	case "just show":
		printEnvConfig()
		break
	case "abort":
		break
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
	bytesRead, err := os.ReadFile(src)
	if err != nil {
		pterm.FgRed.Println(err)
		return false
	}

	err = os.WriteFile(dest, bytesRead, 0644) //nolint:gosec
	if err != nil {
		pterm.FgRed.Println(err)

		return false
	}
	return true
}
