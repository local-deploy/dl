package command

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(downCmd)
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Down project",
	Long:  `Down project.`,
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
