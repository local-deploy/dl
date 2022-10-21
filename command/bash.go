package command

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/varrcan/dl/project"
)

var bashRoot bool

func bashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Login to PHP container",
		Long:  `Login to PHP container as www-data or root user and start bash shell.`,
		Run: func(cmd *cobra.Command, args []string) {
			runBash()
		},
		ValidArgs: []string{"--root"},
	}
	cmd.Flags().BoolVarP(&bashRoot, "root", "r", false, "Login as root")
	return cmd
}

func runBash() {
	project.LoadEnv()

	bash, lookErr := exec.LookPath("bash")
	docker, lookErr := exec.LookPath("docker")
	if lookErr != nil {
		fmt.Println(lookErr)
		return
	}

	site := project.Env.GetString("HOST_NAME")
	container := site + "_php"
	logrus.Infof("Use container name %s", container)

	var root string
	if bashRoot {
		logrus.Info("Login as root user")
		root = "--user root "
	}

	// TODO: rewrite to api
	// github.com/docker/cli@v20.10.18+incompatible/cli/command/container/exec.go
	cmdCompose := &exec.Cmd{
		Path:   bash,
		Args:   []string{bash, "-c", docker + " exec -it " + root + container + " /bin/bash"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}

	err := cmdCompose.Run()
	if err != nil {
		logrus.Error(err)
	}
}
