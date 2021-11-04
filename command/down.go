package command

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
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
	helper.LoadEnv()

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	cmdCompose := &exec.Cmd{
		Path: compose,
		Dir:  helper.ProjectEnv.GetString("PWD"),
		Args: []string{compose, "-p", helper.ProjectEnv.GetString("NETWORK_NAME"), "down"},
		Env:  helper.CmdEnv(),
	}

	stopProject, _ := pterm.DefaultSpinner.Start("Stopping project")
	output, err := cmdCompose.CombinedOutput()
	if err != nil {
		stopProject.Fail(fmt.Sprint(err) + ": " + string(output))
		return
	}
	stopProject.UpdateText("Project has been successfully stopped")
	stopProject.Success()
}
