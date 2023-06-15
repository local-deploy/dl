package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/local-deploy/dl/helper"
	"github.com/local-deploy/dl/project"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func downCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Down project",
		Long: `Stop and remove running project containers and network.
Analogue of the "docker-compose down" command.`,
		Run: func(cmd *cobra.Command, args []string) {
			downRun()
		},
	}
	return cmd
}

func downRun() {
	project.LoadEnv()

	pterm.FgGreen.Printfln("Stopping project...")

	if viper.GetBool("ca") {
		_ = helper.RemoveDirectory(filepath.Join(helper.CertDir(), "conf", project.Env.GetString("NETWORK_NAME")+".yaml"))
		_ = helper.RemoveDirectory(filepath.Join(helper.CertDir(), project.Env.GetString("NETWORK_NAME")))
	}

	bin, option := helper.GetCompose()
	Args := []string{bin}
	preArgs := []string{"-p", project.Env.GetString("NETWORK_NAME"), "down"}

	if len(option) > 0 {
		Args = append(Args, option)
	}

	Args = append(Args, preArgs...)

	logrus.Infof("Run command: %s, args: %s", bin, Args)
	cmdCompose := &exec.Cmd{
		Path:   bin,
		Dir:    project.Env.GetString("PWD"),
		Args:   Args,
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
