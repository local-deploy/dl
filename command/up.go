package command

import (
	"fmt"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/helper"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(upCmd)
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up project",
	Long:  `Up project.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func up() {
	helper.LoadEnv()

	compose, lookErr := exec.LookPath("docker-compose")
	if lookErr != nil {
		pterm.FgRed.Printfln("docker-compose not found. Please install it. https://docs.docker.com/compose/install/")
		return
	}

	cmdCompose := &exec.Cmd{
		Path: compose,
		Dir:  helper.ProjectEnv.GetString("PWD"),
		Args: []string{compose, "-p", helper.ProjectEnv.GetString("NETWORK_NAME"), "up", "-d"},
		Env:  helper.CmdEnv(),
	}

	startProject, _ := pterm.DefaultSpinner.Start("Starting project")
	output, err := cmdCompose.CombinedOutput()
	if err != nil {
		startProject.Fail(fmt.Sprint(err) + ": " + string(output))
		return
	}
	startProject.UpdateText("Project has been successfully started")
	startProject.Success()

	helper.ShowProjectInfo()
}
