package command

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/local-deploy/dl/project"
	"github.com/local-deploy/dl/utils"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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
	} else if copyEnv() {
		pterm.FgGreen.Println("The .env file has been created successfully. Please specify the necessary variables.")
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
	case "just show":
		printEnvConfig()
	case "abort":
	}
}

func printEnvConfig() {
	src, _ := utils.Templates.Open(filepath.Join("config-files", getEnvName()))
	scanner := bufio.NewScanner(src)

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
	var (
		src       string
		bytesRead []byte
		err       error
	)

	currentDir, _ := os.Getwd()

	if project.IsEnvExampleFileExists() {
		src = filepath.Join(currentDir, ".env.example")
		bytesRead, err = os.ReadFile(src)
		if err != nil {
			pterm.FgRed.Println(err)
			return false
		}
	} else {
		src = filepath.Join("config-files", getEnvName())
		bytesRead, _ = utils.Templates.ReadFile(src)
	}

	dest := filepath.Join(currentDir, ".env")
	err = os.WriteFile(dest, bytesRead, 0644) //nolint:gosec
	if err != nil {
		pterm.FgRed.Println(err)

		return false
	}
	return true
}

func getEnvName() string {
	currentDir, _ := os.Getwd()

	bitrixPath := filepath.Join(currentDir, "bitrix")
	_, err := os.Stat(bitrixPath)
	if err != nil {
		return ".env.example"
	}

	return ".env.example-bitrix"
}
