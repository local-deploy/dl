package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Down project",
	Long: `Stop and remove running project containers and network.  
Analogue of the "docker-compose down" command.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func down() {
	project.LoadEnv()

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	pterm.FgGreen.Printfln("Stopping project...")

	cmdCompose := &exec.Cmd{
		Path:   compose,
		Dir:    project.Env.GetString("PWD"),
		Args:   []string{compose, "-p", project.Env.GetString("NETWORK_NAME"), "down"},
		Env:    project.CmdEnv(),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	err := cmdCompose.Run()
	if err != nil {
		pterm.FgGreen.Printfln(fmt.Sprint(err))
		return
	}
	pterm.FgGreen.Printfln("Project has been successfully stopped")
}
